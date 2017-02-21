package crypt

import (
	"github.com/goatcms/goatcore/dependency"
	"github.com/goatcms/goatcms/cmsapp/services"
)

// RegisterDependencies is init callback to register module dependencies
func RegisterDependencies(dp dependency.Provider) error {
	if err := dp.AddDefaultFactory(services.CryptService, CryptFactory); err != nil {
		return err
	}
	return nil
}
