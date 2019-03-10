package wfe

import (
	"sync"

	"github.com/lyraproj/pcore/px"
)

type serviceLoader struct {
	px.ParentedLoader
	lock      sync.Mutex
	services  map[string]px.LoaderEntry
	loadLocks map[string]sync.Mutex
}

// LoadEntry returns the requested entry or nil if no such entry can be found
func (l *serviceLoader) LoadEntry(c px.Context, name px.TypedName) px.LoaderEntry {
	if name.Namespace() != px.NsService {
		return l.ParentedLoader.Parent().LoadEntry(c, name)
	}

	key := name.MapKey()
	l.lock.Lock()
	s, ok := l.services[key]
	if ok {
		l.lock.Unlock()
		return s
	}

	m, mk := l.loadLocks[key]
	if mk {
		l.lock.Unlock()

		// A pending lock was found for the desired service. Wait for it
		m.Lock()
		m.Unlock()

		l.lock.Lock()
		s = l.services[key]
		l.lock.Unlock()
		return s
	}

	// Insert a lock that is specific to the requested service (requests for other
	// services will not need to wait while this service is loaded).
	m = sync.Mutex{}
	m.Lock()
	defer m.Unlock()

	l.loadLocks[key] = m
	l.lock.Unlock()

	s = l.ParentedLoader.Parent().LoadEntry(c, name)

	l.lock.Lock()
	l.services[key] = s
	delete(l.loadLocks, key)
	l.lock.Unlock()
	return s
}

func (l *serviceLoader) NameAuthority() px.URI {
	return l.ParentedLoader.Parent().NameAuthority()
}

func (l *serviceLoader) Parent() px.Loader {
	return l.ParentedLoader.Parent()
}

func ServiceLoader(parent px.Loader) *serviceLoader {
	return &serviceLoader{ParentedLoader: px.NewParentedLoader(parent).(px.ParentedLoader), services: make(map[string]px.LoaderEntry), loadLocks: make(map[string]sync.Mutex)}
}
