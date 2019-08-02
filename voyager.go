package voyager

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/hashicorp/go-multierror"
	_ "github.com/lib/pq"

	"github.com/concourse/voyager/runner"
)

type Migrator interface {
	CurrentVersion(db *sql.DB) (int, error)
	SupportedVersion() (int, error)
	Migrate(db *sql.DB, version int) error
	Up(db *sql.DB) error
	Migrations() ([]migration, []int, error)
}

func NewMigrator(logger lager.Logger, lockID int, source Source, migrationsRunner runner.MigrationsRunner, adapter SchemaAdapter) Migrator {
	if logger == nil {
		logger = lager.NewLogger("voyager")
		logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	}

	return &migrator{
		logger,
		lockID,
		source,
		migrationsRunner,
		adapter,
		&sync.Mutex{},
	}
}

type migrator struct {
	logger             lager.Logger
	lockID             int
	source             Source
	goMigrationsRunner runner.MigrationsRunner
	adapter            SchemaAdapter
	*sync.Mutex
}

//go:generate counterfeiter . SchemaAdapter
type SchemaAdapter interface {
	MigrateFromOldSchema(*sql.DB, int) (int, error)
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

func (m *migrator) CurrentVersion(db *sql.DB) (int, error) {
	var migrationHistoryExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables where table_name='migrations_history')").Scan(&migrationHistoryExists)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return -1, err
	}

	if !migrationHistoryExists && m.adapter != nil {
		return m.adapter.OldSchemaLastVersion(), nil
	}

	var currentVersion int
	var direction string
	var dirty bool
	err = db.QueryRow("SELECT version, direction, dirty FROM migrations_history WHERE status!='failed' ORDER BY tstamp DESC LIMIT 1").Scan(&currentVersion, &direction, &dirty)
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
		err := db.QueryRow("SELECT version, direction FROM migrations_history WHERE status!='failed' AND version < $1 ORDER BY tstamp DESC LIMIT 1", currentVersion).Scan(&currentVersion, &direction)
		if err != nil {
			return -1, multierror.Append(errors.New("could not determine current migration version"), err)
		}
	}
	return currentVersion, nil
}

func (m *migrator) Migrate(db *sql.DB, toVersion int) error {

	acquired, err := m.acquireLock(db)
	if err != nil {
		return err
	}

	if acquired {
		//TODO: Add log line here once logging is added
		defer func() { _, _ = m.releaseLock(db) }()
	}

	var existingDBVersion int
	if m.adapter != nil {
		existingDBVersion, err = m.adapter.MigrateFromOldSchema(db, toVersion)
		if err != nil {
			return err
		}
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS migrations_history (version bigint, tstamp timestamp with time zone, direction varchar, status varchar, dirty boolean)")
	if err != nil {
		return err
	}

	if existingDBVersion > 0 {
		var containsOldMigrationInfo bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations_history where version=$1)", existingDBVersion).Scan(&containsOldMigrationInfo)
		if err != nil {
			return err
		}

		if !containsOldMigrationInfo {
			_, err = db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES ($1, current_timestamp, 'up', 'passed', false)", existingDBVersion)
			if err != nil {
				return err
			}
		}
	}

	currentVersion, err := m.CurrentVersion(db)
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
				err = m.runMigration(db, migration)
				if err != nil {
					return err
				}
			}
		}
	} else {
		for i := len(migrations) - 1; i >= 0; i-- {
			if currentVersion >= migrations[i].Version && migrations[i].Version > toVersion && migrations[i].Direction == "down" {
				err = m.runMigration(db, migrations[i])
				if err != nil {
					return err
				}

			}
		}

		if m.adapter != nil {
			err = m.adapter.MigrateToOldSchema(db, toVersion)
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

func (m *migrator) recordMigrationFailure(db *sql.DB, migration migration, err error, dirty bool) error {
	_, dbErr := db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES ($1, current_timestamp, $2, 'failed', $3)", migration.Version, migration.Direction, dirty)
	return multierror.Append(fmt.Errorf("Migration '%s' failed: %v", migration.Name, err), dbErr)
}

func (m *migrator) runMigration(db *sql.DB, migration migration) error {
	var err error

	switch migration.Strategy {
	case GoMigration:
		err = m.goMigrationsRunner.Run(db, migration.Name)
		if err != nil {
			return m.recordMigrationFailure(db, migration, err, false)
		}
	case SQLTransaction:
		tx, err := db.Begin()
		if err != nil {
			return m.recordMigrationFailure(db, migration, err, false)
		}

		for _, statement := range migration.Statements {
			_, err = tx.Exec(statement)
			if err != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					err = multierror.Append(fmt.Errorf("Transaction %v failed, failed to roll back the migration", statement), rollbackErr)
					return m.recordMigrationFailure(db, migration, err, false)
				}

				err = multierror.Append(fmt.Errorf("Transaction %v failed, rolled back the migration", statement), err)
				if err != nil {
					return m.recordMigrationFailure(db, migration, err, false)
				}
			}
		}
		err = tx.Commit()
		if err != nil {
			return m.recordMigrationFailure(db, migration, err, true)
		}
	case SQLNoTransaction:
		for _, statement := range migration.Statements {
			_, err = db.Exec(statement)
			if err != nil {
				return m.recordMigrationFailure(db, migration, err, true)
			}
		}
	}

	_, err = db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES ($1, current_timestamp, $2, 'passed', false)", migration.Version, migration.Direction)
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

func (m *migrator) Up(db *sql.DB) error {
	_, versions, err := m.Migrations()
	if err != nil {
		return err
	}
	return m.Migrate(db, versions[len(versions)-1])
}

func (m *migrator) acquireLock(db *sql.DB) (bool, error) {
	var acquired bool
	for {
		m.Lock()
		err := db.QueryRow(`SELECT pg_try_advisory_lock($1)`, m.lockID).Scan(&acquired)

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

func (m *migrator) releaseLock(db *sql.DB) (bool, error) {
	var released bool
	for {
		m.Lock()
		err := db.QueryRow(`SELECT pg_advisory_unlock($1)`, m.lockID).Scan(&released)

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

func sortMigrations(migrationList []migration) {
	sort.Slice(migrationList, func(i, j int) bool {
		return migrationList[i].Version < migrationList[j].Version
	})
}
