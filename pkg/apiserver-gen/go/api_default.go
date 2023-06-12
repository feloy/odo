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
)

// DefaultAPIController binds http requests to an api service and writes the service results to the http response
type DefaultAPIController struct {
	service      DefaultAPIServicer
	errorHandler ErrorHandler
}

// DefaultAPIOption for how the controller is set up.
type DefaultAPIOption func(*DefaultAPIController)

// WithDefaultAPIErrorHandler inject ErrorHandler into controller
func WithDefaultAPIErrorHandler(h ErrorHandler) DefaultAPIOption {
	return func(c *DefaultAPIController) {
		c.errorHandler = h
	}
}

// NewDefaultAPIController creates a default api controller
func NewDefaultAPIController(s DefaultAPIServicer, opts ...DefaultAPIOption) Router {
	controller := &DefaultAPIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the DefaultAPIController
func (c *DefaultAPIController) Routes() Routes {
	return Routes{
		"ComponentCommandPost": Route{
			strings.ToUpper("Post"),
			"/api/v1/component/command",
			c.ComponentCommandPost,
		},
		"ComponentGet": Route{
			strings.ToUpper("Get"),
			"/api/v1/component",
			c.ComponentGet,
		},
		"InstanceDelete": Route{
			strings.ToUpper("Delete"),
			"/api/v1/instance",
			c.InstanceDelete,
		},
		"InstanceGet": Route{
			strings.ToUpper("Get"),
			"/api/v1/instance",
			c.InstanceGet,
		},
	}
}

// ComponentCommandPost -
func (c *DefaultAPIController) ComponentCommandPost(w http.ResponseWriter, r *http.Request) {
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
	if err := AssertComponentCommandPostRequestConstraints(componentCommandPostRequestParam); err != nil {
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
func (c *DefaultAPIController) ComponentGet(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultAPIController) InstanceDelete(w http.ResponseWriter, r *http.Request) {
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
func (c *DefaultAPIController) InstanceGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.InstanceGet(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}
