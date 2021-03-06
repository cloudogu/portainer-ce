package http

import (
	"net/http"
	"path/filepath"
	"time"

	portainer "github.com/cloudogu/portainer-ce/api"
	"github.com/cloudogu/portainer-ce/api/crypto"
	"github.com/cloudogu/portainer-ce/api/docker"
	"github.com/cloudogu/portainer-ce/api/http/handler"
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
	"github.com/cloudogu/portainer-ce/api/http/proxy"
	"github.com/cloudogu/portainer-ce/api/http/proxy/factory/kubernetes"
	"github.com/cloudogu/portainer-ce/api/http/security"

	"github.com/cloudogu/portainer-ce/api/kubernetes/cli"
)

// Server implements the portainer.Server interface
type Server struct {
	BindAddress                 string
	AssetsPath                  string
	Status                      *portainer.Status
	ReverseTunnelService        portainer.ReverseTunnelService
	ComposeStackManager         portainer.ComposeStackManager
	CryptoService               portainer.CryptoService
	SignatureService            portainer.DigitalSignatureService
	SnapshotService             portainer.SnapshotService
	FileService                 portainer.FileService
	DataStore                   portainer.DataStore
	GitService                  portainer.GitService
	JWTService                  portainer.JWTService
	LDAPService                 portainer.LDAPService
	OAuthService                portainer.OAuthService
	SwarmStackManager           portainer.SwarmStackManager
	ProxyManager                *proxy.Manager
	KubernetesTokenCacheManager *kubernetes.TokenCacheManager
	Handler                     *handler.Handler
	SSL                         bool
	SSLCert                     string
	SSLKey                      string
	DockerClientFactory         *docker.ClientFactory
	KubernetesClientFactory     *cli.ClientFactory
	KubernetesDeployer          portainer.KubernetesDeployer
}

// Start starts the HTTP server
func (server *Server) Start() error {
	kubernetesTokenCacheManager := server.KubernetesTokenCacheManager

	requestBouncer := security.NewRequestBouncer(server.DataStore, server.JWTService)

	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)

	var authHandler = auth.NewHandler(requestBouncer, rateLimiter)
	authHandler.DataStore = server.DataStore
	authHandler.CryptoService = server.CryptoService
	authHandler.JWTService = server.JWTService
	authHandler.LDAPService = server.LDAPService
	authHandler.ProxyManager = server.ProxyManager
	authHandler.KubernetesTokenCacheManager = kubernetesTokenCacheManager
	authHandler.OAuthService = server.OAuthService

	var roleHandler = roles.NewHandler(requestBouncer)
	roleHandler.DataStore = server.DataStore

	var customTemplatesHandler = customtemplates.NewHandler(requestBouncer)
	customTemplatesHandler.DataStore = server.DataStore
	customTemplatesHandler.FileService = server.FileService
	customTemplatesHandler.GitService = server.GitService

	var dockerHubHandler = dockerhub.NewHandler(requestBouncer)
	dockerHubHandler.DataStore = server.DataStore

	var edgeGroupsHandler = edgegroups.NewHandler(requestBouncer)
	edgeGroupsHandler.DataStore = server.DataStore

	var edgeJobsHandler = edgejobs.NewHandler(requestBouncer)
	edgeJobsHandler.DataStore = server.DataStore
	edgeJobsHandler.FileService = server.FileService
	edgeJobsHandler.ReverseTunnelService = server.ReverseTunnelService

	var edgeStacksHandler = edgestacks.NewHandler(requestBouncer)
	edgeStacksHandler.DataStore = server.DataStore
	edgeStacksHandler.FileService = server.FileService
	edgeStacksHandler.GitService = server.GitService

	var edgeTemplatesHandler = edgetemplates.NewHandler(requestBouncer)
	edgeTemplatesHandler.DataStore = server.DataStore

	var endpointHandler = endpoints.NewHandler(requestBouncer)
	endpointHandler.DataStore = server.DataStore
	endpointHandler.FileService = server.FileService
	endpointHandler.ProxyManager = server.ProxyManager
	endpointHandler.SnapshotService = server.SnapshotService
	endpointHandler.ReverseTunnelService = server.ReverseTunnelService
	endpointHandler.ComposeStackManager = server.ComposeStackManager

	var endpointEdgeHandler = endpointedge.NewHandler(requestBouncer)
	endpointEdgeHandler.DataStore = server.DataStore
	endpointEdgeHandler.FileService = server.FileService
	endpointEdgeHandler.ReverseTunnelService = server.ReverseTunnelService

	var endpointGroupHandler = endpointgroups.NewHandler(requestBouncer)
	endpointGroupHandler.DataStore = server.DataStore

	var endpointProxyHandler = endpointproxy.NewHandler(requestBouncer)
	endpointProxyHandler.DataStore = server.DataStore
	endpointProxyHandler.ProxyManager = server.ProxyManager
	endpointProxyHandler.ReverseTunnelService = server.ReverseTunnelService

	var fileHandler = file.NewHandler(filepath.Join(server.AssetsPath, "public"))

	var motdHandler = motd.NewHandler(requestBouncer)

	var registryHandler = registries.NewHandler(requestBouncer)
	registryHandler.DataStore = server.DataStore
	registryHandler.FileService = server.FileService
	registryHandler.ProxyManager = server.ProxyManager

	var resourceControlHandler = resourcecontrols.NewHandler(requestBouncer)
	resourceControlHandler.DataStore = server.DataStore

	var settingsHandler = settings.NewHandler(requestBouncer)
	settingsHandler.DataStore = server.DataStore
	settingsHandler.FileService = server.FileService
	settingsHandler.JWTService = server.JWTService
	settingsHandler.LDAPService = server.LDAPService
	settingsHandler.SnapshotService = server.SnapshotService

	var stackHandler = stacks.NewHandler(requestBouncer)
	stackHandler.DataStore = server.DataStore
	stackHandler.FileService = server.FileService
	stackHandler.SwarmStackManager = server.SwarmStackManager
	stackHandler.ComposeStackManager = server.ComposeStackManager
	stackHandler.KubernetesDeployer = server.KubernetesDeployer
	stackHandler.GitService = server.GitService

	var tagHandler = tags.NewHandler(requestBouncer)
	tagHandler.DataStore = server.DataStore

	var teamHandler = teams.NewHandler(requestBouncer)
	teamHandler.DataStore = server.DataStore

	var teamMembershipHandler = teammemberships.NewHandler(requestBouncer)
	teamMembershipHandler.DataStore = server.DataStore

	var statusHandler = status.NewHandler(requestBouncer, server.Status)

	var templatesHandler = templates.NewHandler(requestBouncer)
	templatesHandler.DataStore = server.DataStore
	templatesHandler.FileService = server.FileService
	templatesHandler.GitService = server.GitService

	var uploadHandler = upload.NewHandler(requestBouncer)
	uploadHandler.FileService = server.FileService

	var userHandler = users.NewHandler(requestBouncer, rateLimiter)
	userHandler.DataStore = server.DataStore
	userHandler.CryptoService = server.CryptoService

	var websocketHandler = websocket.NewHandler(requestBouncer)
	websocketHandler.DataStore = server.DataStore
	websocketHandler.SignatureService = server.SignatureService
	websocketHandler.ReverseTunnelService = server.ReverseTunnelService
	websocketHandler.KubernetesClientFactory = server.KubernetesClientFactory

	var webhookHandler = webhooks.NewHandler(requestBouncer)
	webhookHandler.DataStore = server.DataStore
	webhookHandler.DockerClientFactory = server.DockerClientFactory

	server.Handler = &handler.Handler{
		RoleHandler:            roleHandler,
		AuthHandler:            authHandler,
		CustomTemplatesHandler: customTemplatesHandler,
		DockerHubHandler:       dockerHubHandler,
		EdgeGroupsHandler:      edgeGroupsHandler,
		EdgeJobsHandler:        edgeJobsHandler,
		EdgeStacksHandler:      edgeStacksHandler,
		EdgeTemplatesHandler:   edgeTemplatesHandler,
		EndpointGroupHandler:   endpointGroupHandler,
		EndpointHandler:        endpointHandler,
		EndpointEdgeHandler:    endpointEdgeHandler,
		EndpointProxyHandler:   endpointProxyHandler,
		FileHandler:            fileHandler,
		MOTDHandler:            motdHandler,
		RegistryHandler:        registryHandler,
		ResourceControlHandler: resourceControlHandler,
		SettingsHandler:        settingsHandler,
		StatusHandler:          statusHandler,
		StackHandler:           stackHandler,
		TagHandler:             tagHandler,
		TeamHandler:            teamHandler,
		TeamMembershipHandler:  teamMembershipHandler,
		TemplatesHandler:       templatesHandler,
		UploadHandler:          uploadHandler,
		UserHandler:            userHandler,
		WebSocketHandler:       websocketHandler,
		WebhookHandler:         webhookHandler,
	}

	httpServer := &http.Server{
		Addr:    server.BindAddress,
		Handler: server.Handler,
	}

	if server.SSL {
		httpServer.TLSConfig = crypto.CreateServerTLSConfiguration()
		return httpServer.ListenAndServeTLS(server.SSLCert, server.SSLKey)
	}
	return httpServer.ListenAndServe()
}
