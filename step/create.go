package step

import (
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/internal"
	"github.com/lyraproj/wfe/wfe"
)

// Create creates a new wfe.Step based on the given serviceapi.Definition
func Create(c px.Context, def serviceapi.Definition) wfe.Step {
	return internal.CreateStep(c, def)
}
