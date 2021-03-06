package edgestacks

import (
	"net/http"

	"github.com/cloudogu/portainer-ce/api"
	"github.com/cloudogu/portainer-ce/api/http/security"
	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
)

// Handler is the HTTP handler used to handle endpoint group operations.
type Handler struct {
	*mux.Router
	requestBouncer *security.RequestBouncer
	DataStore      portainer.DataStore
	FileService    portainer.FileService
	GitService     portainer.GitService
}

// NewHandler creates a handler to manage endpoint group operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}
	h.Handle("/edge_stacks",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackCreate)))).Methods(http.MethodPost)
	h.Handle("/edge_stacks",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackList)))).Methods(http.MethodGet)
	h.Handle("/edge_stacks/{id}",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackInspect)))).Methods(http.MethodGet)
	h.Handle("/edge_stacks/{id}",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackUpdate)))).Methods(http.MethodPut)
	h.Handle("/edge_stacks/{id}",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackDelete)))).Methods(http.MethodDelete)
	h.Handle("/edge_stacks/{id}/file",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackFile)))).Methods(http.MethodGet)
	h.Handle("/edge_stacks/{id}/status",
		bouncer.PublicAccess(httperror.LoggerHandler(h.edgeStackStatusUpdate))).Methods(http.MethodPut)
	return h
}
