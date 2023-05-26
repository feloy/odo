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
	"context"
	"net/http"
)

// DefaultApiRouter defines the required methods for binding the api requests to a responses for the DefaultApi
// The DefaultApiRouter implementation should parse necessary information from the http request,
// pass the data to a DefaultApiServicer to perform the required actions, then write the service results to the http response.
type DefaultApiRouter interface {
	ComponentCommandPost(http.ResponseWriter, *http.Request)
	ComponentGet(http.ResponseWriter, *http.Request)
	InstanceDelete(http.ResponseWriter, *http.Request)
	InstanceGet(http.ResponseWriter, *http.Request)
}

// DefaultApiServicer defines the api actions for the DefaultApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can be ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type DefaultApiServicer interface {
	ComponentCommandPost(context.Context, ComponentCommandPostRequest) (ImplResponse, error)
	ComponentGet(context.Context) (ImplResponse, error)
	InstanceDelete(context.Context) (ImplResponse, error)
	InstanceGet(context.Context) (ImplResponse, error)
}
