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

	// "github.com/gorilla/mux"
)

// DefaultApiController binds http requests to an api service and writes the service results to the http response
type DefaultApiController struct {
	service DefaultApiServicer
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
			"/component/command",
			c.ComponentCommandPost,
		},
		{
			"ComponentGet",
			strings.ToUpper("Get"),
			"/component",
			c.ComponentGet,
		},
		{
			"InstanceDelete",
			strings.ToUpper("Delete"),
			"/instance",
			c.InstanceDelete,
		},
		{
			"InstanceGet",
			strings.ToUpper("Get"),
			"/instance",
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
