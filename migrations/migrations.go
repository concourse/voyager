package migrations

import (
	"database/sql"
	"reflect"

	"github.com/ddadlani/voyager/runner"
)

type TestGoMigrationsRunner struct {
	*sql.DB
	metadata string
}

func NewMigrationsRunner(db *sql.DB) runner.MigrationsRunner {
	return &TestGoMigrationsRunner{db, "sample-metadata"}
}

func (runner *TestGoMigrationsRunner) Run(name string) error {
	res := reflect.ValueOf(runner).MethodByName(name).Call(nil)

	ret := res[0].Interface()

	if ret != nil {
		return ret.(error)
	}

	return nil
}
