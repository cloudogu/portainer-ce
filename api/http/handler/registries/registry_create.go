package registries

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/cloudogu/portainer-ce/api"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type registryCreatePayload struct {
	Name           string
	Type           portainer.RegistryType
	URL            string
	Authentication bool
	Username       string
	Password       string
	Gitlab         portainer.GitlabRegistryData
}

func (payload *registryCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid registry name")
	}
	if govalidator.IsNull(payload.URL) {
		return errors.New("Invalid registry URL")
	}
	if payload.Authentication && (govalidator.IsNull(payload.Username) || govalidator.IsNull(payload.Password)) {
		return errors.New("Invalid credentials. Username and password must be specified when authentication is enabled")
	}
	if payload.Type != portainer.QuayRegistry && payload.Type != portainer.AzureRegistry && payload.Type != portainer.CustomRegistry && payload.Type != portainer.GitlabRegistry {
		return errors.New("Invalid registry type. Valid values are: 1 (Quay.io), 2 (Azure container registry), 3 (custom registry) or 4 (Gitlab registry)")
	}
	return nil
}

func (handler *Handler) registryCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload registryCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	registry := &portainer.Registry{
		Type:               portainer.RegistryType(payload.Type),
		Name:               payload.Name,
		URL:                payload.URL,
		Authentication:     payload.Authentication,
		Username:           payload.Username,
		Password:           payload.Password,
		UserAccessPolicies: portainer.UserAccessPolicies{},
		TeamAccessPolicies: portainer.TeamAccessPolicies{},
		Gitlab:             payload.Gitlab,
	}

	err = handler.DataStore.Registry().CreateRegistry(registry)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the registry inside the database", err}
	}

	hideFields(registry)
	return response.JSON(w, registry)
}
