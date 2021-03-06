package registries

import (
	"net/http"

	"github.com/cloudogu/portainer-ce/api"
	"github.com/cloudogu/portainer-ce/api/http/proxy"
	"github.com/cloudogu/portainer-ce/api/http/security"
	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
)

func hideFields(registry *portainer.Registry) {
	registry.Password = ""
	registry.ManagementConfiguration = nil
}

// Handler is the HTTP handler used to handle registry operations.
type Handler struct {
	*mux.Router
	requestBouncer *security.RequestBouncer
	DataStore      portainer.DataStore
	FileService    portainer.FileService
	ProxyManager   *proxy.Manager
}

// NewHandler creates a handler to manage registry operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}

	h.Handle("/registries",
		bouncer.AdminAccess(httperror.LoggerHandler(h.registryCreate))).Methods(http.MethodPost)
	h.Handle("/registries",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryList))).Methods(http.MethodGet)
	h.Handle("/registries/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.registryInspect))).Methods(http.MethodGet)
	h.Handle("/registries/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.registryUpdate))).Methods(http.MethodPut)
	h.Handle("/registries/{id}/configure",
		bouncer.AdminAccess(httperror.LoggerHandler(h.registryConfigure))).Methods(http.MethodPost)
	h.Handle("/registries/{id}",
		bouncer.AdminAccess(httperror.LoggerHandler(h.registryDelete))).Methods(http.MethodDelete)
	h.PathPrefix("/registries/proxies/gitlab").Handler(
		bouncer.AdminAccess(httperror.LoggerHandler(h.proxyRequestsToGitlabAPIWithoutRegistry)))
	return h
}
