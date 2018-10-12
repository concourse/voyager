package migrations

import (
	"database/sql"
	"reflect"

	"github.com/concourse/voyager/runner"
)

type TestGoMigrationsRunner struct {
	*sql.DB
}

func NewMigrationsRunner(db *sql.DB) runner.MigrationsRunner {
	return &TestGoMigrationsRunner{db}
}

func (runner *TestGoMigrationsRunner) Run(name string) error {

	res := reflect.ValueOf(runner).MethodByName(name).Call(nil)

	ret := res[0].Interface()

	if ret != nil {
		return ret.(error)
	}

	return nil
}
