package controller

import (
	"github.com/cswarm/ck8sd/pkg/controller/computerdeployment"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, computerdeployment.Add)
}
