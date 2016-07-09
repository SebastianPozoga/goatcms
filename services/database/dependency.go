package database

import (
	dep "github.com/goatcms/goat-core/dependency"
	"github.com/s3c0nDD/goatcms/services"
)

const (
	defaultDatabasePath = "sqlite3.db"
)

//Factory is a database depondency builder
func Factory(dp dep.Provider) (dep.Instance, error) {
	return NewDatabase(defaultDatabasePath)
}

//InitDep inicjalize a new database dependency
func InitDep(prov dep.Provider) error {
	if err := prov.AddService(services.DBID, Factory); err != nil {
		return err
	}
	return nil
}
