package runner

type MigrationsRunner interface {
	Run(name string, args []interface{}) error
}
