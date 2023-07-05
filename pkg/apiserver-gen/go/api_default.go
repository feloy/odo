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

// DefaultApiController binds http requests to an api service and writes the service results to the http response
type DefaultApiController struct {
	service      DefaultApiServicer
	errorHandler ErrorHandler
}

// DefaultApiOption for how the controller is set up.
type DefaultApiOption func(*DefaultApiController)

// WithDefaultApiErrorHandler inject ErrorHandler into controller
func WithDefaultApiErrorHandler(h ErrorHandler) DefaultApiOption {
	return func(c *DefaultApiController) {
		c.errorHandler = h
	}
}

// NewDefaultApiController creates a default api controller
func NewDefaultApiController(s DefaultApiServicer, opts ...DefaultApiOption) Router {
	controller := &DefaultApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the DefaultApiController
func (c *DefaultApiController) Routes() Routes {
	return Routes{
		{
			"ComponentCommandPost",
			strings.ToUpper("Post"),
			"/api/v1/component/command",
			c.ComponentCommandPost,
		},
		{
			"ComponentGet",
			strings.ToUpper("Get"),
			"/api/v1/component",
			c.ComponentGet,
		},
		{
			"DevfileGet",
			strings.ToUpper("Get"),
			"/api/v1/devfile",
			c.DevfileGet,
		},
		{
			"DevfilePut",
			strings.ToUpper("Put"),
			"/api/v1/devfile",
			c.DevfilePut,
		},
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
			"InstanceDelete",
			strings.ToUpper("Delete"),
			"/api/v1/instance",
			c.InstanceDelete,
		},
		{
			"InstanceGet",
			strings.ToUpper("Get"),
			"/api/v1/instance",
			c.InstanceGet,
		},
	}
}

// ComponentCommandPost -
func (c *DefaultApiController) ComponentCommandPost(w http.ResponseWriter, r *http.Request) {
	componentCommandPostRequestParam := ComponentCommandPostRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&componentCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertComponentCommandPostRequestRequired(componentCommandPostRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ComponentCommandPost(r.Context(), componentCommandPostRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ComponentGet -
func (c *DefaultApiController) ComponentGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.ComponentGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevfileGet -
func (c *DefaultApiController) DevfileGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.DevfileGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevfilePut -
func (c *DefaultApiController) DevfilePut(w http.ResponseWriter, r *http.Request) {
	devfilePutRequestParam := DevfilePutRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&devfilePutRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertDevfilePutRequestRequired(devfilePutRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevfilePut(r.Context(), devfilePutRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateApplyCommandPost -
func (c *DefaultApiController) DevstateApplyCommandPost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateChartGet(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateCommandCommandNameDelete(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateCommandCommandNameMovePost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateCommandCommandNameSetDefaultPost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateCommandCommandNameUnsetDefaultPost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateCompositeCommandPost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateContainerContainerNameDelete(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateContainerPost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateDevfileDelete(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateDevfileGet(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateDevfilePut(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateEventsPut(w http.ResponseWriter, r *http.Request) {
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

// DevstateExecCommandPost -
func (c *DefaultApiController) DevstateExecCommandPost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateImageImageNameDelete(w http.ResponseWriter, r *http.Request) {
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

// DevstateImagePost -
func (c *DefaultApiController) DevstateImagePost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateMetadataPut(w http.ResponseWriter, r *http.Request) {
	metadataParam := Metadata{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&metadataParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertMetadataRequired(metadataParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.DevstateMetadataPut(r.Context(), metadataParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DevstateQuantityValidPost -
func (c *DefaultApiController) DevstateQuantityValidPost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateResourcePost(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultApiController) DevstateResourceResourceNameDelete(w http.ResponseWriter, r *http.Request) {
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

// InstanceDelete -
func (c *DefaultApiController) InstanceDelete(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.InstanceDelete(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// InstanceGet -
func (c *DefaultApiController) InstanceGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.InstanceGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
