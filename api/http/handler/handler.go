package handler

import (
	"net/http"
	"strings"

	"github.com/cloudogu/portainer-ce/api/http/handler/auth"
	"github.com/cloudogu/portainer-ce/api/http/handler/customtemplates"
	"github.com/cloudogu/portainer-ce/api/http/handler/dockerhub"
	"github.com/cloudogu/portainer-ce/api/http/handler/edgegroups"
	"github.com/cloudogu/portainer-ce/api/http/handler/edgejobs"
	"github.com/cloudogu/portainer-ce/api/http/handler/edgestacks"
	"github.com/cloudogu/portainer-ce/api/http/handler/edgetemplates"
	"github.com/cloudogu/portainer-ce/api/http/handler/endpointedge"
	"github.com/cloudogu/portainer-ce/api/http/handler/endpointgroups"
	"github.com/cloudogu/portainer-ce/api/http/handler/endpointproxy"
	"github.com/cloudogu/portainer-ce/api/http/handler/endpoints"
	"github.com/cloudogu/portainer-ce/api/http/handler/file"
	"github.com/cloudogu/portainer-ce/api/http/handler/motd"
	"github.com/cloudogu/portainer-ce/api/http/handler/registries"
	"github.com/cloudogu/portainer-ce/api/http/handler/resourcecontrols"
	"github.com/cloudogu/portainer-ce/api/http/handler/roles"
	"github.com/cloudogu/portainer-ce/api/http/handler/settings"
	"github.com/cloudogu/portainer-ce/api/http/handler/stacks"
	"github.com/cloudogu/portainer-ce/api/http/handler/status"
	"github.com/cloudogu/portainer-ce/api/http/handler/tags"
	"github.com/cloudogu/portainer-ce/api/http/handler/teammemberships"
	"github.com/cloudogu/portainer-ce/api/http/handler/teams"
	"github.com/cloudogu/portainer-ce/api/http/handler/templates"
	"github.com/cloudogu/portainer-ce/api/http/handler/upload"
	"github.com/cloudogu/portainer-ce/api/http/handler/users"
	"github.com/cloudogu/portainer-ce/api/http/handler/webhooks"
	"github.com/cloudogu/portainer-ce/api/http/handler/websocket"
)

// Handler is a collection of all the service handlers.
type Handler struct {
	AuthHandler            *auth.Handler
	CustomTemplatesHandler *customtemplates.Handler
	DockerHubHandler       *dockerhub.Handler
	EdgeGroupsHandler      *edgegroups.Handler
	EdgeJobsHandler        *edgejobs.Handler
	EdgeStacksHandler      *edgestacks.Handler
	EdgeTemplatesHandler   *edgetemplates.Handler
	EndpointEdgeHandler    *endpointedge.Handler
	EndpointGroupHandler   *endpointgroups.Handler
	EndpointHandler        *endpoints.Handler
	EndpointProxyHandler   *endpointproxy.Handler
	FileHandler            *file.Handler
	MOTDHandler            *motd.Handler
	RegistryHandler        *registries.Handler
	ResourceControlHandler *resourcecontrols.Handler
	RoleHandler            *roles.Handler
	SettingsHandler        *settings.Handler
	StackHandler           *stacks.Handler
	StatusHandler          *status.Handler
	TagHandler             *tags.Handler
	TeamMembershipHandler  *teammemberships.Handler
	TeamHandler            *teams.Handler
	TemplatesHandler       *templates.Handler
	UploadHandler          *upload.Handler
	UserHandler            *users.Handler
	WebSocketHandler       *websocket.Handler
	WebhookHandler         *webhooks.Handler
}

// ServeHTTP delegates a request to the appropriate subhandler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/auth"):
		http.StripPrefix("/api", h.AuthHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/dockerhub"):
		http.StripPrefix("/api", h.DockerHubHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/custom_templates"):
		http.StripPrefix("/api", h.CustomTemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_stacks"):
		http.StripPrefix("/api", h.EdgeStacksHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_groups"):
		http.StripPrefix("/api", h.EdgeGroupsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_jobs"):
		http.StripPrefix("/api", h.EdgeJobsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_stacks"):
		http.StripPrefix("/api", h.EdgeStacksHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/edge_templates"):
		http.StripPrefix("/api", h.EdgeTemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/endpoint_groups"):
		http.StripPrefix("/api", h.EndpointGroupHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/endpoints"):
		switch {
		case strings.Contains(r.URL.Path, "/docker/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/kubernetes/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/storidge/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/azure/"):
			http.StripPrefix("/api/endpoints", h.EndpointProxyHandler).ServeHTTP(w, r)
		case strings.Contains(r.URL.Path, "/edge/"):
			http.StripPrefix("/api/endpoints", h.EndpointEdgeHandler).ServeHTTP(w, r)
		default:
			http.StripPrefix("/api", h.EndpointHandler).ServeHTTP(w, r)
		}
	case strings.HasPrefix(r.URL.Path, "/api/motd"):
		http.StripPrefix("/api", h.MOTDHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/registries"):
		http.StripPrefix("/api", h.RegistryHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/resource_controls"):
		http.StripPrefix("/api", h.ResourceControlHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/roles"):
		http.StripPrefix("/api", h.RoleHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/settings"):
		http.StripPrefix("/api", h.SettingsHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/stacks"):
		http.StripPrefix("/api", h.StackHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/status"):
		http.StripPrefix("/api", h.StatusHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/tags"):
		http.StripPrefix("/api", h.TagHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/templates"):
		http.StripPrefix("/api", h.TemplatesHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/upload"):
		http.StripPrefix("/api", h.UploadHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/users"):
		http.StripPrefix("/api", h.UserHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/teams"):
		http.StripPrefix("/api", h.TeamHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/team_memberships"):
		http.StripPrefix("/api", h.TeamMembershipHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/websocket"):
		http.StripPrefix("/api", h.WebSocketHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/webhooks"):
		http.StripPrefix("/api", h.WebhookHandler).ServeHTTP(w, r)
	case strings.HasPrefix(r.URL.Path, "/"):
		h.FileHandler.ServeHTTP(w, r)
	}
}
