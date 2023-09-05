/*
 * odo dev
 *
 * API interface for 'odo dev'
 *
 * API version: 0.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// DevstateApiController binds http requests to an api service and writes the service results to the http response
type DevstateApiController struct {
	service      DevstateApiServicer
	errorHandler ErrorHandler
}

// DevstateApiOption for how the controller is set up.
type DevstateApiOption func(*DevstateApiController)

// WithDevstateApiErrorHandler inject ErrorHandler into controller
func WithDevstateApiErrorHandler(h ErrorHandler) DevstateApiOption {
	return func(c *DevstateApiController) {
		c.errorHandler = h
	}
}

// NewDevstateApiController creates a default api controller
func NewDevstateApiController(s DevstateApiServicer, opts ...DevstateApiOption) Router {
	controller := &DevstateApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the DevstateApiController
func (c *DevstateApiController) Routes() Routes {
	return Routes{
		{
			"DevstateApplyCommandPost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/applyCommand",
			c.DevstateApplyCommandPost,
		},
		{
			"DevstateChartGet",
			strings.ToUpper("Get"),
			"/api/v1/devstate/chart",
			c.DevstateChartGet,
		},
		{
			"DevstateCommandCommandNameDelete",
			strings.ToUpper("Delete"),
			"/api/v1/devstate/command/{commandName}",
			c.DevstateCommandCommandNameDelete,
		},
		{
			"DevstateCommandCommandNameMovePost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/command/{commandName}/move",
			c.DevstateCommandCommandNameMovePost,
		},
		{
			"DevstateCommandCommandNameSetDefaultPost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/command/{commandName}/setDefault",
			c.DevstateCommandCommandNameSetDefaultPost,
		},
		{
			"DevstateCommandCommandNameUnsetDefaultPost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/command/{commandName}/unsetDefault",
			c.DevstateCommandCommandNameUnsetDefaultPost,
		},
		{
			"DevstateCompositeCommandPost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/compositeCommand",
			c.DevstateCompositeCommandPost,
		},
		{
			"DevstateContainerContainerNameDelete",
			strings.ToUpper("Delete"),
			"/api/v1/devstate/container/{containerName}",
			c.DevstateContainerContainerNameDelete,
		},
		{
			"DevstateContainerPost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/container",
			c.DevstateContainerPost,
		},
		{
			"DevstateDevfileDelete",
			strings.ToUpper("Delete"),
			"/api/v1/devstate/devfile",
			c.DevstateDevfileDelete,
		},
		{
			"DevstateDevfileGet",
			strings.ToUpper("Get"),
			"/api/v1/devstate/devfile",
			c.DevstateDevfileGet,
		},
		{
			"DevstateDevfilePut",
			strings.ToUpper("Put"),
			"/api/v1/devstate/devfile",
			c.DevstateDevfilePut,
		},
		{
			"DevstateEventsPut",
			strings.ToUpper("Put"),
			"/api/v1/devstate/events",
			c.DevstateEventsPut,
		},
		{
			"DevstateExecCommandCommandNamePatch",
			strings.ToUpper("Patch"),
			"/api/v1/devstate/execCommand/{commandName}",
			c.DevstateExecCommandCommandNamePatch,
		},
		{
			"DevstateExecCommandPost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/execCommand",
			c.DevstateExecCommandPost,
		},
		{
			"DevstateImageImageNameDelete",
			strings.ToUpper("Delete"),
			"/api/v1/devstate/image/{imageName}",
			c.DevstateImageImageNameDelete,
		},
		{
			"DevstateImageImageNamePatch",
			strings.ToUpper("Patch"),
			"/api/v1/devstate/image/{imageName}",
			c.DevstateImageImageNamePatch,
		},
		{
			"DevstateImagePost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/image",
			c.DevstateImagePost,
		},
		{
			"DevstateMetadataPut",
			strings.ToUpper("Put"),
			"/api/v1/devstate/metadata",
			c.DevstateMetadataPut,
		},
		{
			"DevstateQuantityValidPost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/quantityValid",
			c.DevstateQuantityValidPost,
		},
		{
			"DevstateResourcePost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/resource",
			c.DevstateResourcePost,
		},
		{
			"DevstateResourceResourceNameDelete",
			strings.ToUpper("Delete"),
			"/api/v1/devstate/resource/{resourceName}",
			c.DevstateResourceResourceNameDelete,
		},
		{
			"DevstateResourceResourceNamePatch",
			strings.ToUpper("Patch"),
			"/api/v1/devstate/resource/{resourceName}",
			c.DevstateResourceResourceNamePatch,
		},
		{
			"DevstateVolumePost",
			strings.ToUpper("Post"),
			"/api/v1/devstate/volume",
			c.DevstateVolumePost,
		},
		{
			"DevstateVolumeVolumeNameDelete",
			strings.ToUpper("Delete"),
			"/api/v1/devstate/volume/{volumeName}",
			c.DevstateVolumeVolumeNameDelete,
		},
		{
			"DevstateVolumeVolumeNamePatch",
			strings.ToUpper("Patch"),
			"/api/v1/devstate/volume/{volumeName}",
			c.DevstateVolumeVolumeNamePatch,
		},
	}
}

// DevstateApplyCommandPost -
func (c *DevstateApiController) DevstateApplyCommandPost(w http.ResponseWriter, r *http.Request) {
	devstateApplyCommandPostRequestParam := DevstateApplyCommandPostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateApplyCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateApplyCommandPostRequestRequired(devstateApplyCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateApplyCommandPost(r.Context(), devstateApplyCommandPostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateChartGet -
func (c *DevstateApiController) DevstateChartGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.DevstateChartGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateCommandCommandNameDelete -
func (c *DevstateApiController) DevstateCommandCommandNameDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	commandNameParam := params["commandName"]
	result, err := c.service.DevstateCommandCommandNameDelete(r.Context(), commandNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateCommandCommandNameMovePost -
func (c *DevstateApiController) DevstateCommandCommandNameMovePost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	commandNameParam := params["commandName"]
	devstateCommandCommandNameMovePostRequestParam := DevstateCommandCommandNameMovePostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateCommandCommandNameMovePostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateCommandCommandNameMovePostRequestRequired(devstateCommandCommandNameMovePostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateCommandCommandNameMovePost(r.Context(), commandNameParam, devstateCommandCommandNameMovePostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateCommandCommandNameSetDefaultPost -
func (c *DevstateApiController) DevstateCommandCommandNameSetDefaultPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	commandNameParam := params["commandName"]
	devstateCommandCommandNameSetDefaultPostRequestParam := DevstateCommandCommandNameSetDefaultPostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateCommandCommandNameSetDefaultPostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateCommandCommandNameSetDefaultPostRequestRequired(devstateCommandCommandNameSetDefaultPostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateCommandCommandNameSetDefaultPost(r.Context(), commandNameParam, devstateCommandCommandNameSetDefaultPostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateCommandCommandNameUnsetDefaultPost -
func (c *DevstateApiController) DevstateCommandCommandNameUnsetDefaultPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	commandNameParam := params["commandName"]
	result, err := c.service.DevstateCommandCommandNameUnsetDefaultPost(r.Context(), commandNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateCompositeCommandPost -
func (c *DevstateApiController) DevstateCompositeCommandPost(w http.ResponseWriter, r *http.Request) {
	devstateCompositeCommandPostRequestParam := DevstateCompositeCommandPostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateCompositeCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateCompositeCommandPostRequestRequired(devstateCompositeCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateCompositeCommandPost(r.Context(), devstateCompositeCommandPostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateContainerContainerNameDelete -
func (c *DevstateApiController) DevstateContainerContainerNameDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	containerNameParam := params["containerName"]
	result, err := c.service.DevstateContainerContainerNameDelete(r.Context(), containerNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateContainerPost -
func (c *DevstateApiController) DevstateContainerPost(w http.ResponseWriter, r *http.Request) {
	devstateContainerPostRequestParam := DevstateContainerPostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateContainerPostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateContainerPostRequestRequired(devstateContainerPostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateContainerPost(r.Context(), devstateContainerPostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateDevfileDelete -
func (c *DevstateApiController) DevstateDevfileDelete(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.DevstateDevfileDelete(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateDevfileGet -
func (c *DevstateApiController) DevstateDevfileGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.DevstateDevfileGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateDevfilePut -
func (c *DevstateApiController) DevstateDevfilePut(w http.ResponseWriter, r *http.Request) {
	devstateDevfilePutRequestParam := DevstateDevfilePutRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateDevfilePutRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateDevfilePutRequestRequired(devstateDevfilePutRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateDevfilePut(r.Context(), devstateDevfilePutRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateEventsPut -
func (c *DevstateApiController) DevstateEventsPut(w http.ResponseWriter, r *http.Request) {
	devstateEventsPutRequestParam := DevstateEventsPutRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateEventsPutRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateEventsPutRequestRequired(devstateEventsPutRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateEventsPut(r.Context(), devstateEventsPutRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateExecCommandCommandNamePatch -
func (c *DevstateApiController) DevstateExecCommandCommandNamePatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	commandNameParam := params["commandName"]
	devstateExecCommandCommandNamePatchRequestParam := DevstateExecCommandCommandNamePatchRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateExecCommandCommandNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateExecCommandCommandNamePatchRequestRequired(devstateExecCommandCommandNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateExecCommandCommandNamePatch(r.Context(), commandNameParam, devstateExecCommandCommandNamePatchRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateExecCommandPost -
func (c *DevstateApiController) DevstateExecCommandPost(w http.ResponseWriter, r *http.Request) {
	devstateExecCommandPostRequestParam := DevstateExecCommandPostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateExecCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateExecCommandPostRequestRequired(devstateExecCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateExecCommandPost(r.Context(), devstateExecCommandPostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateImageImageNameDelete -
func (c *DevstateApiController) DevstateImageImageNameDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	imageNameParam := params["imageName"]
	result, err := c.service.DevstateImageImageNameDelete(r.Context(), imageNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateImageImageNamePatch -
func (c *DevstateApiController) DevstateImageImageNamePatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	imageNameParam := params["imageName"]
	devstateImageImageNamePatchRequestParam := DevstateImageImageNamePatchRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateImageImageNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateImageImageNamePatchRequestRequired(devstateImageImageNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateImageImageNamePatch(r.Context(), imageNameParam, devstateImageImageNamePatchRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateImagePost -
func (c *DevstateApiController) DevstateImagePost(w http.ResponseWriter, r *http.Request) {
	devstateImagePostRequestParam := DevstateImagePostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateImagePostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateImagePostRequestRequired(devstateImagePostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateImagePost(r.Context(), devstateImagePostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateMetadataPut -
func (c *DevstateApiController) DevstateMetadataPut(w http.ResponseWriter, r *http.Request) {
	metadataRequestParam := MetadataRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&metadataRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertMetadataRequestRequired(metadataRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateMetadataPut(r.Context(), metadataRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateQuantityValidPost -
func (c *DevstateApiController) DevstateQuantityValidPost(w http.ResponseWriter, r *http.Request) {
	devstateQuantityValidPostRequestParam := DevstateQuantityValidPostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateQuantityValidPostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateQuantityValidPostRequestRequired(devstateQuantityValidPostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateQuantityValidPost(r.Context(), devstateQuantityValidPostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateResourcePost -
func (c *DevstateApiController) DevstateResourcePost(w http.ResponseWriter, r *http.Request) {
	devstateResourcePostRequestParam := DevstateResourcePostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateResourcePostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateResourcePostRequestRequired(devstateResourcePostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateResourcePost(r.Context(), devstateResourcePostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateResourceResourceNameDelete -
func (c *DevstateApiController) DevstateResourceResourceNameDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	resourceNameParam := params["resourceName"]
	result, err := c.service.DevstateResourceResourceNameDelete(r.Context(), resourceNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateResourceResourceNamePatch -
func (c *DevstateApiController) DevstateResourceResourceNamePatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	resourceNameParam := params["resourceName"]
	devstateResourceResourceNamePatchRequestParam := DevstateResourceResourceNamePatchRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateResourceResourceNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateResourceResourceNamePatchRequestRequired(devstateResourceResourceNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateResourceResourceNamePatch(r.Context(), resourceNameParam, devstateResourceResourceNamePatchRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateVolumePost -
func (c *DevstateApiController) DevstateVolumePost(w http.ResponseWriter, r *http.Request) {
	devstateVolumePostRequestParam := DevstateVolumePostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateVolumePostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateVolumePostRequestRequired(devstateVolumePostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateVolumePost(r.Context(), devstateVolumePostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateVolumeVolumeNameDelete -
func (c *DevstateApiController) DevstateVolumeVolumeNameDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	volumeNameParam := params["volumeName"]
	result, err := c.service.DevstateVolumeVolumeNameDelete(r.Context(), volumeNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateVolumeVolumeNamePatch -
func (c *DevstateApiController) DevstateVolumeVolumeNamePatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	volumeNameParam := params["volumeName"]
	devstateVolumeVolumeNamePatchRequestParam := DevstateVolumeVolumeNamePatchRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devstateVolumeVolumeNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevstateVolumeVolumeNamePatchRequestRequired(devstateVolumeVolumeNamePatchRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateVolumeVolumeNamePatch(r.Context(), volumeNameParam, devstateVolumeVolumeNamePatchRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
