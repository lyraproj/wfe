package identity

import (
	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/wfe/internal"
	"github.com/lyraproj/wfe/wfe"
)

// StartEra starts a new era in the Identity store. If lazy is set, then no service will
// be started until the service is requested to map identifiers
func StartEra(c px.Context, lazy bool) (service serviceapi.Identity) {
	if lazy {
		service = internal.GetLazyIdentity(c)
	} else {
		service = internal.GetIdentity(c)
	}
	service.BumpEra(c)
	return service
}

// SweepAndGC performs a sweep of the Identity store, retrieves all garbage, and
// then tells the handler for each garbage entry to delete the resource. The entry
// is then purged from the Identity store
func SweepAndGC(c px.Context, service serviceapi.Identity, prefix string) {
	log := hclog.Default()
	log.Debug("Identity Sweep", "prefix", prefix)
	service.Sweep(c, prefix)
	log.Debug("Identity Collect garbage", "prefix", prefix)
	gl := service.Garbage(c, prefix)
	ng := gl.Len()
	log.Debug("Identity Collect garbage", "prefix", prefix, "count", ng)
	rs := make([]px.List, ng)

	// Store in reverse order
	ng--
	gl.EachWithIndex(func(t px.Value, i int) {
		rs[ng-i] = t.(px.List)
	})

	for _, l := range rs {
		uri := types.ParseURI(l.At(0).String())
		hid := uri.Query().Get(`hid`)
		if hid == `` {
			continue
		}
		handlerDef := wfe.GetHandler(c, px.NewTypedName(px.NsHandler, hid))
		handler := wfe.GetService(c, handlerDef.ServiceId())

		extId := l.At(1)
		log.Debug("Identity delete", "prefix", prefix, "intId", uri.String(), "extId", extId)
		handler.Invoke(c, handlerDef.Identifier().Name(), `delete`, extId)
		service.PurgeExternal(c, extId.String())
	}
	service.PurgeReferences(c, prefix)
}
