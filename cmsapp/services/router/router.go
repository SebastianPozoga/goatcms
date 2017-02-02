package router

import (
	"fmt"
	"net/http"

	"github.com/goatcms/goat-core/app"
	"github.com/goatcms/goat-core/dependency"
	"github.com/goatcms/goat-core/filesystem"
	"github.com/goatcms/goatcms/cmsapp/services"
	"github.com/gorilla/mux"
)

const (
	DefaultHost = ":5555"
)

type Router struct {
	deps struct {
		EventScope   app.Scope            `dependency:"EngineScope"`
		AppScope     app.Scope            `dependency:"AppScope"`
		Host         string               `config:"?mux.host"`
		TmpFilespace filesystem.Filespace `filespace:"tmp"`
	}
	dependencyFactories map[string]dependency.Factory
	grouter             *mux.Router
	dp                  dependency.Provider
}

func RouterFactory(dp dependency.Provider) (interface{}, error) {
	m := &Router{
		dependencyFactories: map[string]dependency.Factory{},
		grouter:             mux.NewRouter(),
		dp:                  dp,
	}
	if err := dp.InjectTo(&m.deps); err != nil {
		return nil, err
	}
	return services.Router(m), nil
}

// Get append http get routing to global pool
func (router *Router) OnGet(path string, handler services.ScopeHandler) {
	router.grouter.HandleFunc(path, router.scopeHandlerToMuxHandler(handler)).Methods("GET")
}

// Post append http post routing to global pool
func (router *Router) OnPost(path string, handler services.ScopeHandler) {
	router.grouter.HandleFunc(path, router.scopeHandlerToMuxHandler(handler)).Methods("POST")
}

// Put append http put routing to global pool
func (router *Router) OnPut(path string, handler services.ScopeHandler) {
	router.grouter.HandleFunc(path, router.scopeHandlerToMuxHandler(handler)).Methods("PUT")
}

// Delete append http delete routing to global pool
func (router *Router) OnDelete(path string, handler services.ScopeHandler) {
	router.grouter.HandleFunc(path, router.scopeHandlerToMuxHandler(handler)).Methods("DELETE")
}

// Delete append http delete routing to global pool
func (router *Router) On(methods []string, path string, handler services.ScopeHandler) {
	router.grouter.HandleFunc(path, router.scopeHandlerToMuxHandler(handler)).Methods(methods...)
}

// Host return current host value
func (router *Router) Host() string {
	if router.deps.Host != "" {
		return router.deps.Host
	}
	return DefaultHost
}

// Start add routing to global pool
func (router *Router) Start() error {
	http.Handle("/", router.grouter)
	if err := http.ListenAndServe(router.Host(), nil); err != nil {
		return err
	}
	return nil
}

// Start add routing to global pool
func (router *Router) AddFactory(name string, factory dependency.Factory) error {
	if _, ok := router.dependencyFactories[name]; ok {
		return fmt.Errorf("router.AddFactory: Add %s dependency twice", name)
	}
	router.dependencyFactories[name] = factory
	return nil
}

func (router *Router) newRequestScope(w http.ResponseWriter, r *http.Request) app.Scope {
	rs := NewRequestScope(router.dp, router.deps.TmpFilespace, router.dependencyFactories, w, r)
	router.deps.EventScope.Trigger(services.CreateRequestScope, rs)
	return rs
}

func (router *Router) scopeHandlerToMuxHandler(handler services.ScopeHandler) services.MuxHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := router.newRequestScope(w, r)
		handler(scope)
	}
}
