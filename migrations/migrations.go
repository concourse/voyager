package migrations

import (
	"database/sql"
	"reflect"

	"github.com/ddadlani/voyager/runner"
)

type TestGoMigrationsRunner struct {
	*sql.DB
}

func NewMigrationsRunner(db *sql.DB) runner.MigrationsRunner {
	return &TestGoMigrationsRunner{db}
}

func (runner *TestGoMigrationsRunner) Run(name string, args []interface{}) error {

	var valueArgs []reflect.Value
	if len(args) < 1 {
		valueArgs = nil
	} else {
		valueArgs = make([]reflect.Value, len(args))
		for i, arg := range args {
			valueArgs[i] = reflect.ValueOf(arg)
		}
	}

	res := reflect.ValueOf(runner).MethodByName(name).Call(valueArgs)

	ret := res[0].Interface()

	if ret != nil {
		return ret.(error)
	}

	return nil
}
