package voyager

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"code.cloudfoundry.org/lager"
)

var noTxPrefix = regexp.MustCompile(`^\s*--\s+(NO_TRANSACTION)`)
var migrationDirection = regexp.MustCompile(`\.(up|down)\.`)
var goMigrationFuncName = regexp.MustCompile(`(Up|Down)_[0-9]*`)

var ErrCouldNotParseDirection = errors.New("could not parse direction for migration")

type parser struct {
	source Source
}

func NewParser(source Source) *parser {
	return &parser{
		source: source,
	}
}

func (p *parser) ParseMigrationFilename(logger lager.Logger, fileName string) (Migration, error) {
	var (
		migration Migration
		err       error
	)

	logger.WithData(lager.Data{
		"fileName": fileName,
	})
	migration.Direction, err = determineDirection(fileName)
	if err != nil {
		logger.Error("failed-to-parse-migration-direction", err)
		return migration, err
	}

	migration.Version, err = schemaVersion(fileName)
	if err != nil {
		logger.Error("failed-to-parse-migration-version", err)
		return migration, err
	}

	return migration, nil
}

func (p *parser) ParseFileToMigration(logger lager.Logger, migrationName string) (Migration, error) {
	var migrationContents string

	migration, err := p.ParseMigrationFilename(logger, migrationName)
	if err != nil {
		return migration, err
	}

	migrationBytes, err := p.source.Asset(migrationName)
	if err != nil {
		return migration, err
	}

	migrationContents = string(migrationBytes)
	migration.Strategy = determineMigrationStrategy(migrationName, migrationContents)

	switch migration.Strategy {
	case GoMigration:
		migration.Name = goMigrationFuncName.FindString(migrationContents)
	case SQLNoTransaction:
		migration.Statements = splitStatements(migrationContents)
		migration.Name = migrationName
	case SQLTransaction:
		migration.Statements = splitStatements(migrationContents)
		migration.Name = migrationName
	}

	return migration, nil
}

func schemaVersion(assetName string) (int, error) {
	regex := regexp.MustCompile(`(\d+)`)
	match := regex.FindStringSubmatch(assetName)
	return strconv.Atoi(match[1])
}

func determineDirection(migrationName string) (string, error) {
	matches := migrationDirection.FindStringSubmatch(migrationName)
	if len(matches) < 2 {
		return "", ErrCouldNotParseDirection
	}

	return matches[1], nil
}

func determineMigrationStrategy(migrationName string, migrationContents string) Strategy {
	if strings.HasSuffix(migrationName, ".go") {
		return GoMigration
	} else {
		if noTxPrefix.MatchString(migrationContents) {
			return SQLNoTransaction
		}
	}
	return SQLTransaction
}

func splitStatements(migrationContents string) []string {
	var (
		fileStatements      []string
		migrationStatements []string
	)
	fileStatements = append(fileStatements, strings.Split(migrationContents, ";")...)
	// last string is empty
	if strings.TrimSpace(fileStatements[len(fileStatements)-1]) == "" {
		fileStatements = fileStatements[:len(fileStatements)-1]
	}

	var isSqlStatement bool = false
	var sqlStatement string
	for _, statement := range fileStatements {
		statement = strings.TrimSpace(statement)

		if statement == "BEGIN" || statement == "COMMIT" {
			continue
		}
		if strings.Contains(statement, "BEGIN") {
			isSqlStatement = true
			sqlStatement = statement + ";"
		} else {

			if isSqlStatement {
				sqlStatement = strings.Join([]string{sqlStatement, statement, ";"}, "")
				if strings.HasPrefix(statement, "$$") {
					migrationStatements = append(migrationStatements, sqlStatement)
					isSqlStatement = false
				}
			} else {
				migrationStatements = append(migrationStatements, statement)
			}
		}
	}
	return migrationStatements
}
