package status

import (
	"net/http"

	"github.com/cloudogu/portainer-ce/api"
	"github.com/cloudogu/portainer-ce/api/http/security"
	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
)

// Handler is the HTTP handler used to handle status operations.
type Handler struct {
	*mux.Router
	Status *portainer.Status
}

// NewHandler creates a handler to manage status operations.
func NewHandler(bouncer *security.RequestBouncer, status *portainer.Status) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
		Status: status,
	}
	h.Handle("/status",
		bouncer.PublicAccess(httperror.LoggerHandler(h.statusInspect))).Methods(http.MethodGet)
	h.Handle("/status/version",
		bouncer.AuthenticatedAccess(http.HandlerFunc(h.statusInspectVersion))).Methods(http.MethodGet)

	return h
}
