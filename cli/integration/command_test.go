package integration_test

import (
	"database/sql"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/onsi/gomega/gbytes"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Migration CLI", func() {

	var (
		migrationDir string
		err          error
		db           *sql.DB
	)

	BeforeEach(func() {
		db, err = sql.Open("postgres", postgresRunner.DataSourceName())
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		_ = db.Close()
	})

	Context("Generate command", func() {
		var (
			migrationName = "create_new_table"
		)

		BeforeEach(func() {
			migrationDir, err = ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(migrationDir)
		})

		Context("sql migrations", func() {
			It("generates up and down migration files", func() {

				cmd := exec.Command(cliPath, "generate", "-d", migrationDir, "-n", migrationName, "-t", "sql")
				session, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())

				sqlMigrationPattern := "^(\\d+)_(.*).(down|up).sql$"

				Eventually(session).Should(gexec.Exit(0))

				expectGeneratedFilesToMatchSpecification(migrationDir, sqlMigrationPattern,
					migrationName, func(migrationID string, actualFileContents string) {
						Expect(actualFileContents).To(BeEmpty())
					})
			})
		})

		Context("go migration", func() {
			It("generates up and down migration files", func() {
				cmd := exec.Command(cliPath, "generate", "-d", migrationDir, "-n", migrationName, "-t", "go")
				session, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())

				goMigrationPattern := "^(\\d+)_(.*).(down|up).go$"

				Eventually(session).Should(gexec.Exit(0))

				expectGeneratedFilesToMatchSpecification(migrationDir, goMigrationPattern,
					migrationName, func(migrationID string, actualFileContents string) {
						lines := strings.Split(actualFileContents, "\n")
						Expect(lines).To(HaveLen(6))

						Expect(lines[0]).To(Equal("package migrations"))
						Expect(lines[1]).To(Equal(""))
						Expect(lines[2]).To(HavePrefix("//"))
						Expect(lines[3]).To(MatchRegexp("^func (Up|Down)_%s\\(\\) error {", migrationID))
						Expect(lines[4]).To(ContainSubstring("return nil"))
						Expect(lines[5]).To(Equal("}"))
					})
			})
		})
	})

	Context("CurrentVersion command", func() {

		It("reports the current version as reported by voyager.CurrentVersion", func() {
			setupMigrationsHistoryTableToExistAtVersion(db, 12345, false)

			cmd := exec.Command(cliPath, "current-version", "--connection-string", postgresRunner.DataSourceName())
			session, err := gexec.Start(cmd, nil, nil)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Last successfully applied migration: 12345"))
		})
	})

	Context("Migrate command", func() {
		Context("When an error occurs", func() {
			It("surfaces the error", func() {
				cmd := exec.Command(cliPath, "migrate", "-v", "20000", "--connection-string", postgresRunner.DataSourceName())

				session, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say("an error occurred while running migrations"))
			})
		})
		Context("When migrating is successful", func() {
			It("migrates to the version requested", func() {
				cmd := exec.Command(cliPath, "migrate", "-v", "2000", "--connection-string", postgresRunner.DataSourceName())

				session, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out).To(gbytes.Say("Successfully migrated to version 2000"))
			})

			It("migrates to the latest version when latest is specified", func() {
				cmd := exec.Command(cliPath, "migrate", "-v", "latest", "--connection-string", postgresRunner.DataSourceName())

				session, err := gexec.Start(cmd, nil, nil)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(0))
				Expect(session.Out).To(gbytes.Say("Successfully migrated to version latest"))
			})
		})

	})
})

//TODO: move to helpers
func setupMigrationsHistoryTableToExistAtVersion(db *sql.DB, version int, dirty bool) {
	_, err := db.Exec(`CREATE TABLE migrations_history(version bigint, tstamp timestamp with time zone, direction varchar, status varchar, dirty boolean)`)
	Expect(err).NotTo(HaveOccurred())

	_, err = db.Exec(`INSERT INTO migrations_history(version, tstamp, direction, status, dirty) VALUES($1, current_timestamp, 'up', 'passed', $2)`, version, dirty)
	Expect(err).NotTo(HaveOccurred())
}

func expectGeneratedFilesToMatchSpecification(migrationDir, fileNamePattern, migrationName string,
	checkContents func(migrationID string, actualFileContents string)) {

	files, err := ioutil.ReadDir(migrationDir)
	Expect(err).ToNot(HaveOccurred())
	var migrationFilesCount = 0
	regex := regexp.MustCompile(fileNamePattern)
	for _, migrationFile := range files {
		var matches []string
		migrationFileName := migrationFile.Name()
		if regex.MatchString(migrationFileName) {
			matches = regex.FindStringSubmatch(migrationFileName)

			Expect(matches).To(HaveLen(4))
			Expect(matches[2]).To(Equal(migrationName))

			filePath := path.Join(migrationDir, migrationFileName)
			info, err := os.Stat(filePath)
			Expect(uint32(info.Mode())).To(Equal(uint32(0644)))
			fileContents, err := ioutil.ReadFile(filePath)
			Expect(err).ToNot(HaveOccurred())
			checkContents(matches[1], string(fileContents))
			migrationFilesCount++
		}
	}
	Expect(migrationFilesCount).To(Equal(2))
}
