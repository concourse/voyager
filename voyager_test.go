package voyager_test

import (
	"database/sql"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/concourse/voyager"
	"github.com/concourse/voyager/migrations"
	"github.com/concourse/voyager/runner"
	"github.com/concourse/voyager/voyagerfakes"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Voyager Migration", func() {
	var (
		err      error
		db       *sql.DB
		source   *voyagerfakes.FakeSource
		runner   runner.MigrationsRunner
		lockID   int
		migrator voyager.Migrator
	)

	BeforeEach(func() {
		db, err = sql.Open("postgres", postgresRunner.DataSourceName())
		Expect(err).ToNot(HaveOccurred())

		source = new(voyagerfakes.FakeSource)
		source.AssetStub = asset
		source.AssetNamesStub = assetNames
		runner = migrations.NewMigrationsRunner()
		lockID = 12345
	})

	JustBeforeEach(func() {
	})

	AfterEach(func() {
		_ = db.Close()
	})

	Context("Migration test run", func() {
		It("Runs all the migrations", func() {
			migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
			err := migrator.Up(db)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Current Version", func() {
		BeforeEach(func() {
			SetupMigrationsHistoryTableToExistAtVersion(db, 2000, false)
		})

		Context("when the latest migration was an up migration", func() {
			It("reports the current version stored in the database", func() {
				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				version, err := migrator.CurrentVersion(db)
				Expect(err).ToNot(HaveOccurred())
				Expect(version).To(Equal(2000))
			})
		})

		Context("when the latest migration was a down migration", func() {
			BeforeEach(func() {
				_, err = db.Exec(`INSERT INTO migrations_history(version, tstamp, direction, status, dirty) VALUES($1, current_timestamp, 'up', 'passed', false)`, 3000)
				Expect(err).ToNot(HaveOccurred())
				_, err = db.Exec(`INSERT INTO migrations_history(version, tstamp, direction, status, dirty) VALUES($1, current_timestamp, 'up', 'passed', false)`, 4000)
				Expect(err).ToNot(HaveOccurred())
				_, err = db.Exec(`INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES ($1, current_timestamp, 'down', 'passed', false)`, 4000)
				Expect(err).ToNot(HaveOccurred())
			})

			It("reports the version before the latest down migration", func() {

				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				version, err := migrator.CurrentVersion(db)
				Expect(err).ToNot(HaveOccurred())
				Expect(version).To(Equal(3000))
			})
		})

		Context("when the latest migration was dirty", func() {
			BeforeEach(func() {
				_, err = db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES (3000, current_timestamp, 'down', 'passed', true)")
				Expect(err).ToNot(HaveOccurred())
			})
			It("throws an error", func() {

				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				_, err := migrator.CurrentVersion(db)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Database is in a dirty state"))
			})
		})

		Context("when the latest migration failed", func() {
			BeforeEach(func() {
				_, err = db.Exec("INSERT INTO migrations_history (version, tstamp, direction, status, dirty) VALUES (3000, current_timestamp, 'down', 'failed', false)")
				Expect(err).ToNot(HaveOccurred())
			})
			It("reports the version before the failed migration", func() {
				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				version, err := migrator.CurrentVersion(db)
				Expect(err).ToNot(HaveOccurred())
				Expect(version).To(Equal(2000))
			})
		})
	})

	Context("Supported Version", func() {
		BeforeEach(func() {
			source.AssetNamesReturns([]string{
				"1000_some_migration.up.sql",
				"3000_this_is_to_prove_we_dont_use_string_sort.up.sql",
				"20000_latest_migration.up.sql",
				"1000_some_migration.down.sql",
				"3000_this_is_to_prove_we_dont_use_string_sort.down.sql",
				"20000_latest_migration.down.sql",
				"migrations.go",
			})
		})
		It("SupportedVersion reports the highest supported migration version", func() {
			migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
			version, err := migrator.SupportedVersion()
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal(20000))
		})

		It("Ignores files it can't parse", func() {
			migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
			version, err := migrator.SupportedVersion()
			Expect(err).ToNot(HaveOccurred())
			Expect(version).To(Equal(20000))
		})
	})

	Context("Upgrade", func() {
		Context("old database schema table exist", func() {
			var fakeAdapter *voyagerfakes.FakeSchemaAdapter
			BeforeEach(func() {
				SetupOldDatabaseSchema(db, 8878, false)
				fakeAdapter = new(voyagerfakes.FakeSchemaAdapter)
				fakeAdapter.MigrateFromOldSchemaReturns(8878, nil)

				source.AssetNamesReturns([]string{
					"1000_initial_migration.up.sql",
					"1000_initial_migration.down.sql",
					"2000_update_some_table.up.sql",
					"2000_update_some_table.down.sql",
				})
			})

			It("populate migrations_history table with starting version from old schema table", func() {
				migrator = voyager.NewMigrator(logger, lockID, source, runner, fakeAdapter)
				startTime := time.Now()

				err = migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

				var (
					version   int
					isDirty   bool
					timeStamp pq.NullTime
					status    string
					direction string
				)
				err = db.QueryRow("SELECT * from migrations_history ORDER BY tstamp ASC LIMIT 1").Scan(&version, &timeStamp, &direction, &status, &isDirty)
				Expect(version).To(Equal(8878))
				Expect(isDirty).To(BeFalse())
				Expect(timeStamp.Time.After(startTime)).To(Equal(true))
				Expect(direction).To(Equal("up"))
				Expect(status).To(Equal("passed"))
			})

			Context("when the migrations_history table already exists", func() {
				It("does not repopulate the migrations_history table", func() {
					SetupMigrationsHistoryTableToExistAtVersion(db, 8878, false)
					startTime := time.Now()

					migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
					err = migrator.Up(db)
					Expect(err).ToNot(HaveOccurred())

					var timeStamp pq.NullTime
					rows, err := db.Query("SELECT tstamp FROM migrations_history WHERE version=8878")
					Expect(err).ToNot(HaveOccurred())
					var numRows = 0
					for rows.Next() {
						err = rows.Scan(&timeStamp)
						Expect(err).ToNot(HaveOccurred())
						numRows++
					}
					Expect(numRows).To(Equal(1))
					Expect(timeStamp.Time.Before(startTime)).To(Equal(true))
				})
			})

			Context("when migrating to a version that doesn't exist", func() {
				It("returns an error and doesn't perform any migration", func() {
					migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
					err = migrator.Migrate(db, 1000)
					Expect(err).ToNot(HaveOccurred())

					err = migrator.Migrate(db, 20000)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not find migration version 20000. No changes were made"))

					var timeStamp pq.NullTime
					rows, err := db.Query("SELECT tstamp FROM migrations_history WHERE version=1000")
					Expect(err).ToNot(HaveOccurred())
					var numRows = 0
					for rows.Next() {
						err = rows.Scan(&timeStamp)
						Expect(err).ToNot(HaveOccurred())
						numRows++
					}
					Expect(numRows).To(Equal(1))
				})
			})
		})

		Context("sql migrations", func() {
			var simpleMigrationFilename string
			BeforeEach(func() {
				simpleMigrationFilename = "1000_test_table_created.up.sql"
				source.AssetReturns([]byte(`
						BEGIN;
						CREATE TABLE some_table (id integer);
						COMMIT;
						`), nil)

				source.AssetNamesReturns([]string{
					simpleMigrationFilename,
				})
			})
			It("runs a migration", func() {
				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				migrations, _, err := migrator.Migrations()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(migrations)).To(Equal(1))

				err = migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

				By("Creating the table in the database")
				var exists string
				err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.tables where table_name = 'some_table')").Scan(&exists)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(Equal("true"))

				By("Updating the migrations_history table")
				ExpectDatabaseMigrationVersionToEqual(db, migrator, 1000)
			})

			It("ignores migrations before the current version", func() {
				SetupMigrationsHistoryTableToExistAtVersion(db, 1000, false)
				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				err := migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

				By("Not creating the database referenced in the migration")
				var exists string
				err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.tables where table_name = 'some_table')").Scan(&exists)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(Equal("false"))
			})

			It("runs the up migrations in ascending order", func() {
				addTableMigrationFilename := "1000_test_table_created.up.sql"
				removeTableMigrationFilename := "1001_test_table_created.up.sql"

				source.AssetStub = func(name string) ([]byte, error) {
					if name == addTableMigrationFilename {
						return []byte(`
						BEGIN;
						CREATE TABLE some_table (id integer);
						COMMIT;
						`), nil
					} else if name == removeTableMigrationFilename {
						return []byte(`
						BEGIN;
						DROP TABLE some_table;
						COMMIT;
						`), nil
					}
					return asset(name)
				}

				source.AssetNamesReturns([]string{
					removeTableMigrationFilename,
					addTableMigrationFilename,
				})

				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				err := migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

			})

			Context("With a transactional migration", func() {
				It("leaves the database clean after a failure", func() {
					SetupMigrationsHistoryTableToExistAtVersion(db, 1000, false)
					source.AssetNamesReturns([]string{
						"1200_delete_nonexistent_table.up.sql",
					})

					source.AssetReturns([]byte(`
						DROP TABLE nonexistent;
					`), nil)

					migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
					err := migrator.Up(db)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("rolled back the migration"))
					ExpectDatabaseMigrationVersionToEqual(db, migrator, 1000)
					ExpectMigrationToHaveFailed(db, 1200, false)
				})
			})

			It("Doesn't fail if there are no migrations to run", func() {
				SetupMigrationsHistoryTableToExistAtVersion(db, 1000, false)
				source.AssetNamesReturns([]string{
					"1000_initial_migration.up.sql",
				})

				source.AssetReturns([]byte(`
						CREATE table some_table(id int, tstamp timestamp);
				`), nil)

				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				err = migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

				ExpectDatabaseMigrationVersionToEqual(db, migrator, 1000)

				var someTableExists bool
				err = db.QueryRow(`SELECT EXISTS ( SELECT 1 FROM information_schema.tables WHERE table_name='some_table')`).Scan(&someTableExists)
				Expect(err).ToNot(HaveOccurred())
				Expect(someTableExists).To(Equal(false))
			})

			It("Locks the database so multiple ATCs don't all run migrations at the same time", func() {
				SetupMigrationsHistoryTableToExistAtVersion(db, 900, false)

				source.AssetNamesReturns([]string{
					"1000_initial_migration.up.sql",
				})

				source.AssetReturns(ioutil.ReadFile("migrations/1000_initial_migration.up.sql"))

				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)

				var wg sync.WaitGroup
				wg.Add(3)

				go TryRunUpAndVerifyResult(db, migrator, &wg)
				go TryRunUpAndVerifyResult(db, migrator, &wg)
				go TryRunUpAndVerifyResult(db, migrator, &wg)

				wg.Wait()

				var numRows int
				err := db.QueryRow(`SELECT COUNT(*) from some_table`).Scan(&numRows)
				Expect(err).ToNot(HaveOccurred())
				Expect(numRows).To(Equal(12))
			})

			Context("With a non-transactional migration", func() {
				It("fails if the migration version is in a dirty state", func() {
					source.AssetNamesReturns([]string{
						"1200_delete_nonexistent_table.up.sql",
					})

					source.AssetReturns([]byte(`
							-- NO_TRANSACTION
							DROP TABLE nonexistent;
					`), nil)

					migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
					err := migrator.Up(db)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(MatchRegexp("Migration.*failed"))

					ExpectMigrationToHaveFailed(db, 1200, true)
				})

				It("successfully runs a non-transactional migration", func() {
					source.AssetNamesReturns(
						[]string{
							"30000_no_transaction_migration.up.sql",
						},
					)
					source.AssetReturnsOnCall(1, []byte(`
							-- NO_TRANSACTION
							CREATE TYPE enum_type AS ENUM ('blue_type', 'green_type');
							ALTER TYPE enum_type ADD VALUE 'some_type';
					`), nil)
					startTime := time.Now()
					migrator := voyager.NewMigrator(logger, lockID, source, runner, nil)
					err = migrator.Up(db)
					Expect(err).ToNot(HaveOccurred())

					var (
						version   int
						isDirty   bool
						timeStamp pq.NullTime
						status    string
						direction string
					)
					err = db.QueryRow("SELECT * from migrations_history ORDER BY tstamp DESC").Scan(&version, &timeStamp, &direction, &status, &isDirty)
					Expect(version).To(Equal(30000))
					Expect(isDirty).To(BeFalse())
					Expect(timeStamp.Time.After(startTime)).To(Equal(true))
					Expect(direction).To(Equal("up"))
					Expect(status).To(Equal("passed"))
				})

				It("gracefully fails on a failing non-transactional migration", func() {
					source.AssetNamesReturns(
						[]string{
							"50000_failing_no_transaction_migration.up.sql",
						},
					)
					source.AssetReturns([]byte(`
							-- NO_TRANSACTION
							CREATE TYPE enum_type AS ENUM ('blue_type', 'green_type');
							ALTER TYPE nonexistent_enum_type ADD VALUE 'some_type';
					`), nil)
					startTime := time.Now()
					migrator := voyager.NewMigrator(logger, lockID, source, runner, nil)
					err = migrator.Up(db)
					Expect(err).To(HaveOccurred())

					var (
						version   int
						isDirty   bool
						timeStamp pq.NullTime
						status    string
						direction string
					)
					err = db.QueryRow("SELECT * from migrations_history ORDER BY tstamp DESC").Scan(&version, &timeStamp, &direction, &status, &isDirty)
					Expect(version).To(Equal(50000))
					Expect(isDirty).To(BeTrue())
					Expect(timeStamp.Time.After(startTime)).To(Equal(true))
					Expect(direction).To(Equal("up"))
					Expect(status).To(Equal("failed"))
				})
			})

		})

		Context("Golang migrations", func() {
			var args []interface{}
			BeforeEach(func() {
				args = make([]interface{}, 2)
				args[0] = "some_table"
				args[1] = "metadata"
			})
			It("runs a migration with Migrate", func() {
				source.AssetNamesReturns([]string{
					"1000_initial_migration.up.sql",
					"4000_go_migration.up.go",
				})

				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				By("applying the initial migration")
				err := migrator.Migrate(db, 1000)
				Expect(err).ToNot(HaveOccurred())
				var columnExists string
				err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.columns where table_name = 'some_table' AND column_name = 'name')").Scan(&columnExists)
				Expect(err).ToNot(HaveOccurred())
				Expect(columnExists).To(Equal("false"))

				err = migrator.Migrate(db, 4000)
				Expect(err).ToNot(HaveOccurred())

				By("applying the go migration")
				err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.columns where table_name = 'some_table' AND column_name='name')").Scan(&columnExists)
				Expect(err).ToNot(HaveOccurred())
				Expect(columnExists).To(Equal("true"))

				By("updating the migrations history table")
				ExpectDatabaseMigrationVersionToEqual(db, migrator, 4000)
			})

			It("runs a migration with Up", func() {
				source.AssetNamesReturns([]string{
					"1000_initial_migration.up.sql",
					"1000_initial_migration.down.sql",
					"4000_go_migration.up.go",
					"4000_go_migration.down.go",
				})

				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				err := migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

				By("applying the migration")
				var columnExists string
				err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.columns where table_name = 'some_table' AND column_name = 'name')").Scan(&columnExists)
				Expect(err).ToNot(HaveOccurred())
				Expect(columnExists).To(Equal("true"))

				By("updating the schema migrations table")
				ExpectDatabaseMigrationVersionToEqual(db, migrator, 4000)
			})
		})
	})

	Context("Downgrade", func() {

		Context("Downgrades to a version with new migrations_history table", func() {
			BeforeEach(func() {
				source.AssetNamesReturns([]string{
					"1000_initial_migration.up.sql",
					"1000_initial_migration.down.sql",
					"2000_update_some_table.up.sql",
					"2000_update_some_table.down.sql",
				})
			})
			It("Downgrades to a given version", func() {
				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				err := migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

				currentVersion, err := migrator.CurrentVersion(db)
				Expect(err).ToNot(HaveOccurred())
				Expect(currentVersion).To(Equal(2000))

				err = migrator.Migrate(db, 1000)
				Expect(err).ToNot(HaveOccurred())

				currentVersion, err = migrator.CurrentVersion(db)
				Expect(err).ToNot(HaveOccurred())
				Expect(currentVersion).To(Equal(1000))

				ExpectToBeAbleToInsertData(db)
			})

			It("Doesn't fail if already at the requested version", func() {
				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				err := migrator.Migrate(db, 2000)
				Expect(err).ToNot(HaveOccurred())

				currentVersion, err := migrator.CurrentVersion(db)
				Expect(err).ToNot(HaveOccurred())
				Expect(currentVersion).To(Equal(2000))

				err = migrator.Migrate(db, 2000)
				Expect(err).ToNot(HaveOccurred())

				currentVersion, err = migrator.CurrentVersion(db)
				Expect(err).ToNot(HaveOccurred())
				Expect(currentVersion).To(Equal(2000))

				ExpectToBeAbleToInsertData(db)
			})

			It("Locks the database so multiple consumers don't run downgrade at the same time", func() {
				migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
				err := migrator.Up(db)
				Expect(err).ToNot(HaveOccurred())

				var wg sync.WaitGroup
				wg.Add(3)

				go TryRunMigrateAndVerifyResult(db, migrator, 1000, &wg)
				go TryRunMigrateAndVerifyResult(db, migrator, 1000, &wg)
				go TryRunMigrateAndVerifyResult(db, migrator, 1000, &wg)

				wg.Wait()
			})
		})

		Context("downgrades to a version with the old database schema", func() {
			var dirty bool

			Context("dirty state is true", func() {
				It("errors", func() {
					dirty = true
					SetupMigrationsHistoryTableToExistAtVersion(db, 4000, dirty)

					migrator = voyager.NewMigrator(logger, lockID, source, runner, nil)
					err = migrator.Migrate(db, 1000)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Database is in a dirty state"))

					var oldSchemaCreated bool
					err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='old_schema')").Scan(&oldSchemaCreated)
					Expect(oldSchemaCreated).To(BeFalse())
				})
			})

			Context("dirty state is false", func() {
				var fakeAdapter *voyagerfakes.FakeSchemaAdapter
				BeforeEach(func() {
					dirty = false
					source.AssetNamesReturns([]string{
						"1000_initial_migration.down.sql",
						"1000_initial_migration.up.sql",
						"2000_update_some_table.down.sql",
						"2000_update_some_table.up.sql",
					})

					fakeAdapter = new(voyagerfakes.FakeSchemaAdapter)
					fakeAdapter.MigrateToOldSchemaStub = func(db *sql.DB, version int) error {

						_, err := db.Exec("CREATE TABLE old_schema (version bigint, dirty boolean)")
						if err != nil {
							return err
						}

						_, err = db.Exec("INSERT INTO old_schema (version, dirty) VALUES (3456, false)")
						if err != nil {
							return err
						}
						return nil
					}
				})

				It("populates old schema table with corresponding first version from migrations_history table", func() {
					migrator = voyager.NewMigrator(logger, lockID, source, runner, fakeAdapter)
					err = migrator.Up(db)
					Expect(err).ToNot(HaveOccurred())

					var newSchemaExists bool
					err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='migrations_history')").Scan(&newSchemaExists)
					Expect(newSchemaExists).To(BeTrue())

					err = migrator.Migrate(db, 1000)
					Expect(err).ToNot(HaveOccurred())

					err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='migrations_history')").Scan(&newSchemaExists)
					Expect(err).ToNot(HaveOccurred())
					Expect(newSchemaExists).To(BeTrue())

					var oldSchemaCreated bool
					err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='old_schema')").Scan(&oldSchemaCreated)
					Expect(err).ToNot(HaveOccurred())
					Expect(oldSchemaCreated).To(BeTrue())
					ExpectDatabaseVersionToEqual(db, 3456, "old_schema")
				})
			})
		})
	})

})

func TryRunUpAndVerifyResult(db *sql.DB, migrator voyager.Migrator, wg *sync.WaitGroup) {
	defer GinkgoRecover()
	defer wg.Done()

	err := migrator.Up(db)
	Expect(err).ToNot(HaveOccurred())

	ExpectDatabaseMigrationVersionToEqual(db, migrator, 1000)
	ExpectToBeAbleToInsertData(db)
}

func TryRunMigrateAndVerifyResult(db *sql.DB, migrator voyager.Migrator, version int, wg *sync.WaitGroup) {
	defer GinkgoRecover()
	defer wg.Done()

	err := migrator.Migrate(db, version)
	Expect(err).ToNot(HaveOccurred())

	ExpectDatabaseMigrationVersionToEqual(db, migrator, version)

	ExpectToBeAbleToInsertData(db)
}

func SetupMigrationsHistoryTableToExistAtVersion(db *sql.DB, version int, dirty bool) {
	_, err := db.Exec(`CREATE TABLE migrations_history(version bigint, tstamp timestamp with time zone, direction varchar, status varchar, dirty boolean)`)
	Expect(err).ToNot(HaveOccurred())

	_, err = db.Exec(`INSERT INTO migrations_history(version, tstamp, direction, status, dirty) VALUES($1, current_timestamp, 'up', 'passed', $2)`, version, dirty)
	Expect(err).ToNot(HaveOccurred())
}

func SetupOldDatabaseSchema(db *sql.DB, version int, dirty bool) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS old_schema (version bigint, dirty boolean)")
	Expect(err).ToNot(HaveOccurred())
	_, err = db.Exec("INSERT INTO old_schema (version, dirty) VALUES ($1, $2)", version, dirty)
	Expect(err).ToNot(HaveOccurred())
}

func ExpectDatabaseMigrationVersionToEqual(db *sql.DB, migrator voyager.Migrator, expectedVersion int) {
	var dbVersion int
	dbVersion, err := migrator.CurrentVersion(db)
	Expect(err).ToNot(HaveOccurred())
	Expect(dbVersion).To(Equal(expectedVersion))
}

func ExpectToBeAbleToInsertData(dbConn *sql.DB) {
	rand.Seed(time.Now().UnixNano())

	teamID := rand.Intn(10000)
	_, err := dbConn.Exec("INSERT INTO some_table(id, tstamp) VALUES ($1, current_timestamp)", teamID)
	Expect(err).ToNot(HaveOccurred())
}

func ExpectMigrationToHaveFailed(dbConn *sql.DB, failedVersion int, expectDirty bool) {
	var status string
	var dirty bool
	err := dbConn.QueryRow("SELECT status, dirty FROM migrations_history WHERE version=$1 ORDER BY tstamp desc LIMIT 1", failedVersion).Scan(&status, &dirty)
	Expect(err).ToNot(HaveOccurred())
	Expect(status).To(Equal("failed"))
	Expect(dirty).To(Equal(expectDirty))
}

func ExpectDatabaseVersionToEqual(db *sql.DB, version int, table string) {
	var dbVersion int
	query := "SELECT version from " + table + " LIMIT 1"
	err := db.QueryRow(query).Scan(&dbVersion)
	Expect(err).ToNot(HaveOccurred())
	Expect(dbVersion).To(Equal(version))
}
