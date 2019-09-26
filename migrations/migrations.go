package migrations

import (
	"database/sql"
	"reflect"

	"github.com/concourse/voyager/runner"
)

type TestGoMigrationsRunner struct {
	DB       *sql.DB
	metadata string
}

func NewMigrationsRunner() runner.MigrationsRunner {
	return &TestGoMigrationsRunner{metadata: "sample-metadata"}
}

func (runner *TestGoMigrationsRunner) Run(db *sql.DB, name string) error {
	runner.DB = db
	res := reflect.ValueOf(runner).MethodByName(name).Call(nil)

	ret := res[0].Interface()

	if ret != nil {
		return ret.(error)
	}

	return nil
}
