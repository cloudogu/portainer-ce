package edgestacks

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/cloudogu/portainer-ce/api"
	"github.com/cloudogu/portainer-ce/api/filesystem"
	"github.com/cloudogu/portainer-ce/api/internal/edge"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

// POST request on /api/endpoint_groups
func (handler *Handler) edgeStackCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: method", err}
	}

	edgeStack, err := handler.createSwarmStack(method, r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create Edge stack", err}
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoints from database", err}
	}

	endpointGroups, err := handler.DataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoint groups from database", err}
	}

	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge groups from database", err}
	}

	relatedEndpoints, err := edge.EdgeStackRelatedEndpoints(edgeStack.EdgeGroups, endpoints, endpointGroups, edgeGroups)

	for _, endpointID := range relatedEndpoints {
		relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find endpoint relation in database", err}
		}

		relation.EdgeStacks[edgeStack.ID] = true

		err = handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist endpoint relation in database", err}
		}
	}

	return response.JSON(w, edgeStack)
}

func (handler *Handler) createSwarmStack(method string, r *http.Request) (*portainer.EdgeStack, error) {
	switch method {
	case "string":
		return handler.createSwarmStackFromFileContent(r)
	case "repository":
		return handler.createSwarmStackFromGitRepository(r)
	case "file":
		return handler.createSwarmStackFromFileUpload(r)
	}
	return nil, errors.New("Invalid value for query parameter: method. Value must be one of: string, repository or file")
}

type swarmStackFromFileContentPayload struct {
	Name             string
	StackFileContent string
	EdgeGroups       []portainer.EdgeGroupID
}

func (payload *swarmStackFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	if payload.EdgeGroups == nil || len(payload.EdgeGroups) == 0 {
		return errors.New("Edge Groups are mandatory for an Edge stack")
	}
	return nil
}

func (handler *Handler) createSwarmStackFromFileContent(r *http.Request) (*portainer.EdgeStack, error) {
	var payload swarmStackFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	err = handler.validateUniqueName(payload.Name)
	if err != nil {
		return nil, err
	}

	stackID := handler.DataStore.EdgeStack().GetNextIdentifier()
	stack := &portainer.EdgeStack{
		ID:           portainer.EdgeStackID(stackID),
		Name:         payload.Name,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		CreationDate: time.Now().Unix(),
		EdgeGroups:   payload.EdgeGroups,
		Status:       make(map[portainer.EndpointID]portainer.EdgeStackStatus),
		Version:      1,
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return nil, err
	}
	stack.ProjectPath = projectPath

	err = handler.DataStore.EdgeStack().CreateEdgeStack(stack)
	if err != nil {
		return nil, err
	}

	return stack, nil
}

type swarmStackFromGitRepositoryPayload struct {
	Name                        string
	RepositoryURL               string
	RepositoryReferenceName     string
	RepositoryAuthentication    bool
	RepositoryUsername          string
	RepositoryPassword          string
	ComposeFilePathInRepository string
	EdgeGroups                  []portainer.EdgeGroupID
}

func (payload *swarmStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && (govalidator.IsNull(payload.RepositoryUsername) || govalidator.IsNull(payload.RepositoryPassword)) {
		return errors.New("Invalid repository credentials. Username and password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.ComposeFilePathInRepository) {
		payload.ComposeFilePathInRepository = filesystem.ComposeFileDefaultName
	}
	if payload.EdgeGroups == nil || len(payload.EdgeGroups) == 0 {
		return errors.New("Edge Groups are mandatory for an Edge stack")
	}
	return nil
}

func (handler *Handler) createSwarmStackFromGitRepository(r *http.Request) (*portainer.EdgeStack, error) {
	var payload swarmStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	err = handler.validateUniqueName(payload.Name)
	if err != nil {
		return nil, err
	}

	stackID := handler.DataStore.EdgeStack().GetNextIdentifier()
	stack := &portainer.EdgeStack{
		ID:           portainer.EdgeStackID(stackID),
		Name:         payload.Name,
		EntryPoint:   payload.ComposeFilePathInRepository,
		CreationDate: time.Now().Unix(),
		EdgeGroups:   payload.EdgeGroups,
		Status:       make(map[portainer.EndpointID]portainer.EdgeStackStatus),
		Version:      1,
	}

	projectPath := handler.FileService.GetEdgeStackProjectPath(strconv.Itoa(int(stack.ID)))
	stack.ProjectPath = projectPath

	gitCloneParams := &cloneRepositoryParameters{
		url:            payload.RepositoryURL,
		referenceName:  payload.RepositoryReferenceName,
		path:           projectPath,
		authentication: payload.RepositoryAuthentication,
		username:       payload.RepositoryUsername,
		password:       payload.RepositoryPassword,
	}

	err = handler.cloneGitRepository(gitCloneParams)
	if err != nil {
		return nil, err
	}

	err = handler.DataStore.EdgeStack().CreateEdgeStack(stack)
	if err != nil {
		return nil, err
	}

	return stack, nil
}

type swarmStackFromFileUploadPayload struct {
	Name             string
	StackFileContent []byte
	EdgeGroups       []portainer.EdgeGroupID
}

func (payload *swarmStackFromFileUploadPayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return errors.New("Invalid stack name")
	}
	payload.Name = name

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.StackFileContent = composeFileContent

	var edgeGroups []portainer.EdgeGroupID
	err = request.RetrieveMultiPartFormJSONValue(r, "EdgeGroups", &edgeGroups, false)
	if err != nil || len(edgeGroups) == 0 {
		return errors.New("Edge Groups are mandatory for an Edge stack")
	}
	payload.EdgeGroups = edgeGroups
	return nil
}

func (handler *Handler) createSwarmStackFromFileUpload(r *http.Request) (*portainer.EdgeStack, error) {
	payload := &swarmStackFromFileUploadPayload{}
	err := payload.Validate(r)
	if err != nil {
		return nil, err
	}

	err = handler.validateUniqueName(payload.Name)
	if err != nil {
		return nil, err
	}

	stackID := handler.DataStore.EdgeStack().GetNextIdentifier()
	stack := &portainer.EdgeStack{
		ID:           portainer.EdgeStackID(stackID),
		Name:         payload.Name,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		CreationDate: time.Now().Unix(),
		EdgeGroups:   payload.EdgeGroups,
		Status:       make(map[portainer.EndpointID]portainer.EdgeStackStatus),
		Version:      1,
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return nil, err
	}
	stack.ProjectPath = projectPath

	err = handler.DataStore.EdgeStack().CreateEdgeStack(stack)
	if err != nil {
		return nil, err
	}

	return stack, nil
}

func (handler *Handler) validateUniqueName(name string) error {
	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return err
	}

	for _, stack := range edgeStacks {
		if strings.EqualFold(stack.Name, name) {
			return errors.New("Edge stack name must be unique")
		}
	}
	return nil
}
