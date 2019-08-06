package cli

import (
	"code.cloudfoundry.org/lager"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"text/template"
	"time"

	"github.com/concourse/voyager"
	"github.com/concourse/voyager/migrations"
	"github.com/gobuffalo/packr"
)

var defaultMigrationDir = "../../migrations/"
var advisoryLockID = 14656

type MigrationType string

const (
	SQL MigrationType = "sql"
	Go  MigrationType = "go"
)

type MigrationCommand struct {
	//TODO : use commands shortcuts.
	GenerateCommand       GenerateCommand       `command:"generate"`
	CurrentVersionCommand CurrentVersionCommand `command:"current-version"`
	MigrateCommand        MigrateCommand        `command:"migrate"`
}

type MigrateCommand struct {
	MigrationDirectory string `short:"d" long:"directory" default:"migrations" description:"The directory to which the migration files should be written"`
	Version            string `short:"v" long:"version" default:"latest" description:"The migration version to migrate up or down to. Use 'latest' to migrate to the latest available version"`
	ConnectionString   string `short:"c" long:"connection-string" description:"connection string for db"`
}

type GenerateCommand struct {
	MigrationDirectory string        `short:"d" long:"directory" default:"migrations" description:"The directory to which the migration files should be written"`
	MigrationName      string        `short:"n" long:"name" description:"The name of the migration"`
	Type               MigrationType `short:"t" long:"type" description:"The file type of the migration"`
}

type CurrentVersionCommand struct {
	ConnectionString string `short:"c" long:"connection-string"`
}

func (c *CurrentVersionCommand) Execute(args []string) error {
	logger := lager.NewLogger("voyager")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	db, err := sql.Open("postgres", c.ConnectionString)
	if err != nil {
		return err
	}

	migrator := voyager.NewMigrator(advisoryLockID, nil, nil, nil)

	version, err := migrator.CurrentVersion(logger, db)

	if err != nil {
		return err
	}

	fmt.Println("Last successfully applied migration:", version)

	return nil
}

func (c *GenerateCommand) Execute(args []string) error {
	if c.Type == SQL {
		return c.generateSQLMigration()
	} else if c.Type == Go {
		return c.generateGoMigration()
	}
	return fmt.Errorf("unsupported migration type %s. Supported types include %s and %s", c.Type, SQL, Go)
}

func (c *GenerateCommand) generateSQLMigration() error {
	currentTime := time.Now().Unix()
	fileNameFormat := "%d_%s.%s.sql"

	upMigrationFileName := fmt.Sprintf(fileNameFormat, currentTime, c.MigrationName, "up")
	downMigrationFileName := fmt.Sprintf(fileNameFormat, currentTime, c.MigrationName, "down")

	contents := ""

	err := ioutil.WriteFile(path.Join(c.MigrationDirectory, upMigrationFileName), []byte(contents), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(c.MigrationDirectory, downMigrationFileName), []byte(contents), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *MigrateCommand) Execute(args []string) error {
	logger := lager.NewLogger("voyager")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	db, err := sql.Open("postgres", c.ConnectionString)

	if err != nil {
		return err
	}

	box := packr.NewBox(defaultMigrationDir)
	migrator := voyager.NewMigrator(0, source{box}, migrations.NewMigrationsRunner(), nil)

	var toVersion int

	if c.Version == "latest" {
		toVersion = 0
	} else {
		toVersion, err = strconv.Atoi(c.Version)
		if err != nil {
			return err
		}
	}

	err = migrator.Migrate(logger, db, toVersion)

	if err != nil {
		return fmt.Errorf("an error occurred while running migrations: %s", err.Error())
	}

	fmt.Println("Successfully migrated to version", c.Version)
	return nil
}

type migrationInfo struct {
	MigrationId int64
	Direction   string
}

const goMigrationTemplate = `package migrations

// implement the migration in this function
func {{ .Direction }}_{{ .MigrationId }}() error {
	return nil
}`

func renderGoMigrationToFile(filePath string, state migrationInfo) error {
	migrationFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer migrationFile.Close()

	tmpl, err := template.New("go-migration-template").Parse(goMigrationTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(migrationFile, state)
	if err != nil {
		return err
	}

	return nil
}

func (c *GenerateCommand) generateGoMigration() error {
	currentTime := time.Now().Unix()
	fileNameFormat := "%d_%s.%s.go"

	upMigrationFileName := fmt.Sprintf(fileNameFormat, currentTime, c.MigrationName, "up")
	downMigrationFileName := fmt.Sprintf(fileNameFormat, currentTime, c.MigrationName, "down")

	err := renderGoMigrationToFile(path.Join(c.MigrationDirectory, upMigrationFileName), migrationInfo{
		MigrationId: currentTime,
		Direction:   "Up",
	})
	if err != nil {
		return err
	}

	err = renderGoMigrationToFile(path.Join(c.MigrationDirectory, downMigrationFileName), migrationInfo{
		MigrationId: currentTime,
		Direction:   "Down",
	})
	if err != nil {
		return err
	}

	return nil
}

type source struct {
	packr.Box
}

func (s source) AssetNames() []string {
	migrations := make([]string, 0)
	for _, name := range s.List() {
		if name != "migrations.go" {
			migrations = append(migrations, name)
		}
	}

	return migrations
}

func (s source) Asset(name string) ([]byte, error) {
	return s.MustBytes(name)
}
