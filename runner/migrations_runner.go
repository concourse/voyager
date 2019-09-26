package runner

import "database/sql"

type MigrationsRunner interface {
	Run(db *sql.DB, name string) error
}
