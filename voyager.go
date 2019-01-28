package voyager

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/voyager/runner"
	multierror "github.com/hashicorp/go-multierror"
	_ "github.com/lib/pq"
)

type Migrator interface {
	CurrentVersion() (int, error)
	SupportedVersion() (int, error)
	Migrate(version int) error
	Up() error
	Migrations() ([]migration, []int, error)
}

func NewMigrator(db *sql.DB, lockID int, source Source, migrationsRunner runner.MigrationsRunner, adapter SchemaAdapter) Migrator {
	return &migrator{
		db,
		lockID,
		lager.NewLogger("migrations"),
		source,
		migrationsRunner,
		adapter,
		&sync.Mutex{},
	}
}

type migrator struct {
	db                 *sql.DB
	lockID             int
	logger             lager.Logger
	source             Source
	goMigrationsRunner runner.MigrationsRunner
	adapter            SchemaAdapter
	*sync.Mutex
}

//go:generate counterfeiter . SchemaAdapter
type SchemaAdapter interface {
	MigrateFromOldSchema(*sql.DB) (int, error)
	MigrateToOldSchema(*sql.DB, int) error
	OldSchemaLastVersion() int
}

func (m *migrator) SupportedVersion() (int, error) {
	matches := []migration{}

	assets := m.source.AssetNames()

	var parser = NewParser(m.source)
	for _, match := range assets {
		if migration, err := parser.ParseMigrationFilename(match); err == nil {
			matches = append(matches, migration)
		}
	}
	sortMigrations(matches)
	return matches[len(matches)-1].Version, nil
}

func (m *migrator) CurrentVersion() (int, error) {

	var migrationHistoryExists bool
	err := m.db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables where table_name='migrations_history')").Scan(&migrationHistoryExists)
	if !migrationHistoryExists && m.adapter != nil {
		return m.adapter.OldSchemaLastVersion(), nil
	}

	var currentVersion int
	var direction string
	var dirty bool
	err = m.db.QueryRow("SELECT version, direction, dirty FROM migrations_history WHERE status!='failed' ORDER BY tstamp DESC LIMIT 1").Scan(&currentVersion, &direction, &dirty)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return -1, err
	}

	if dirty {
		return -1, errors.New("could not determine current migration version. Database is in a dirty state")
	}

	for direction == "down" {
		err := m.db.QueryRow("SELECT version, direction FROM migrations_history WHERE status!='failed' AND version < $1 ORDER BY tstamp DESC LIMIT 1", currentVersion).Scan(&currentVersion, &direction)
		if err != nil {
			return -1, multierror.Append(errors.New("could not determine current migration version"), err)
		}
	}
	return currentVersion, nil
}

func (m *migrator) Migrate(toVersion int) error {

	acquired, err := m.acquireLock()
	if err != nil {
		return err
	}

	if acquired {
		defer m.releaseLock()
	}

	var existingDBVersion int
	if m.adapter != nil {
		existingDBVersion, err = m.adapter.MigrateFromOldSchema(m.db)
		if err != nil {
			return err
		}
	}

	_, err = m.db.Exec("CREATE TABLE IF NOT EXISTS migrations_history (version bigint, tstamp timestamp with time zone, direction varchar, status varchar, dirty boolean)")
	if err != nil {
		return err
	}

	if existingDBVersion > 0 {
		var containsOldMigrationInfo bool
		err = m.db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations_history where version=$1)", existingDBVersion).Scan(&containsOldMigrationInfo)

		if !containsOldMigrationInfo {
			_, err = m.db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES ($1, current_timestamp, 'up', 'passed', false)", existingDBVersion)
			if err != nil {
				return err
			}
		}
	}

	currentVersion, err := m.CurrentVersion()
	if err != nil {
		return err
	}

	migrations, versions, err := m.Migrations()
	if err != nil {
		return err
	}

	isValidVersion := false
	if toVersion == 0 {
		toVersion = versions[len(versions)-1]
		isValidVersion = true
	} else {
		for _, v := range versions {
			if v == toVersion {
				isValidVersion = true
				break
			}
		}
	}

	if !isValidVersion {
		return fmt.Errorf("could not find migration version %v. No changes were made", toVersion)
	}

	if currentVersion <= toVersion {
		for _, migration := range migrations {
			if currentVersion < migration.Version && migration.Version <= toVersion && migration.Direction == "up" {
				err = m.runMigration(migration)
				if err != nil {
					return err
				}
			}
		}
	} else {
		for i := len(migrations) - 1; i >= 0; i-- {
			if currentVersion >= migrations[i].Version && migrations[i].Version > toVersion && migrations[i].Direction == "down" {
				err = m.runMigration(migrations[i])
				if err != nil {
					return err
				}

			}
		}

		if m.adapter != nil {
			err = m.adapter.MigrateToOldSchema(m.db, toVersion)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type Strategy int

const (
	GoMigration Strategy = iota
	SQLTransaction
	SQLNoTransaction
)

type migration struct {
	Name       string
	Version    int
	Direction  string
	Statements []string
	Strategy   Strategy
}

func (m *migrator) recordMigrationFailure(migration migration, err error, dirty bool) error {
	_, dbErr := m.db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES ($1, current_timestamp, $2, 'failed', $3)", migration.Version, migration.Direction, dirty)
	return multierror.Append(fmt.Errorf("Migration '%s' failed: %v", migration.Name, err), dbErr)
}

func (m *migrator) runMigration(migration migration) error {
	var err error

	switch migration.Strategy {
	case GoMigration:
		err = m.goMigrationsRunner.Run(migration.Name)
		if err != nil {
			return m.recordMigrationFailure(migration, err, false)
		}
	case SQLTransaction:
		tx, err := m.db.Begin()
		for _, statement := range migration.Statements {
			_, err = tx.Exec(statement)
			if err != nil {
				tx.Rollback()
				err = multierror.Append(fmt.Errorf("Transaction %v failed, rolled back the migration", statement), err)
				if err != nil {
					return m.recordMigrationFailure(migration, err, false)
				}
			}
		}
		err = tx.Commit()
	case SQLNoTransaction:
		for _, statement := range migration.Statements {
			_, err = m.db.Exec(statement)
			if err != nil {
				return m.recordMigrationFailure(migration, err, true)
			}
		}
	}

	_, err = m.db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES ($1, current_timestamp, $2, 'passed', false)", migration.Version, migration.Direction)
	return err
}

func (m *migrator) Migrations() ([]migration, []int, error) {
	migrationList := []migration{}
	versionList := []int{}
	assets := m.source.AssetNames()
	var parser = NewParser(m.source)
	for _, assetName := range assets {
		parsedMigration, err := parser.ParseFileToMigration(assetName)
		if err != nil {
			return nil, nil, err
		}
		versionList = append(versionList, parsedMigration.Version)
		migrationList = append(migrationList, parsedMigration)
	}

	sortMigrations(migrationList)
	sort.Ints(versionList)

	return migrationList, versionList, nil
}

func (m *migrator) Up() error {
	_, versions, err := m.Migrations()
	if err != nil {
		return err
	}
	return m.Migrate(versions[len(versions)-1])
}

func (m *migrator) acquireLock() (bool, error) {
	var acquired bool
	for {
		m.Lock()
		err := m.db.QueryRow(`SELECT pg_try_advisory_lock($1)`, m.lockID).Scan(&acquired)

		if err != nil {
			m.Unlock()
			return false, err
		}

		if acquired {
			m.Unlock()
			return acquired, nil
		}

		m.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (m *migrator) releaseLock() (bool, error) {

	var released bool
	for {
		m.Lock()
		err := m.db.QueryRow(`SELECT pg_advisory_unlock($1)`, m.lockID).Scan(&released)

		if err != nil {
			m.Unlock()
			return false, err
		}

		if released {
			m.Unlock()
			return released, nil
		}

		m.Unlock()
		time.Sleep(1 * time.Second)
	}
}

type filenames []string

func sortMigrations(migrationList []migration) {
	sort.Slice(migrationList, func(i, j int) bool {
		return migrationList[i].Version < migrationList[j].Version
	})
}
