package controller

import (
	"github.com/patoarvizu/kms-vault-operator/pkg/controller/kmsvaultsecret"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kmsvaultsecret.Add)
}
