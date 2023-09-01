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
	DevfileGet(http.ResponseWriter, *http.Request)
	DevfilePut(http.ResponseWriter, *http.Request)
	InstanceDelete(http.ResponseWriter, *http.Request)
	InstanceGet(http.ResponseWriter, *http.Request)
	TelemetryGet(http.ResponseWriter, *http.Request)
}

// DevstateApiRouter defines the required methods for binding the api requests to a responses for the DevstateApi
// The DevstateApiRouter implementation should parse necessary information from the http request,
// pass the data to a DevstateApiServicer to perform the required actions, then write the service results to the http response.
type DevstateApiRouter interface {
	DevstateApplyCommandPost(http.ResponseWriter, *http.Request)
	DevstateChartGet(http.ResponseWriter, *http.Request)
	DevstateCommandCommandNameDelete(http.ResponseWriter, *http.Request)
	DevstateCommandCommandNameMovePost(http.ResponseWriter, *http.Request)
	DevstateCommandCommandNameSetDefaultPost(http.ResponseWriter, *http.Request)
	DevstateCommandCommandNameUnsetDefaultPost(http.ResponseWriter, *http.Request)
	DevstateCompositeCommandPost(http.ResponseWriter, *http.Request)
	DevstateContainerContainerNameDelete(http.ResponseWriter, *http.Request)
	DevstateContainerPost(http.ResponseWriter, *http.Request)
	DevstateDevfileDelete(http.ResponseWriter, *http.Request)
	DevstateDevfileGet(http.ResponseWriter, *http.Request)
	DevstateDevfilePut(http.ResponseWriter, *http.Request)
	DevstateEventsPut(http.ResponseWriter, *http.Request)
	DevstateExecCommandPost(http.ResponseWriter, *http.Request)
	DevstateImageImageNameDelete(http.ResponseWriter, *http.Request)
	DevstateImagePost(http.ResponseWriter, *http.Request)
	DevstateMetadataPut(http.ResponseWriter, *http.Request)
	DevstateQuantityValidPost(http.ResponseWriter, *http.Request)
	DevstateResourcePost(http.ResponseWriter, *http.Request)
	DevstateResourceResourceNameDelete(http.ResponseWriter, *http.Request)
	DevstateResourceResourceNamePatch(http.ResponseWriter, *http.Request)
	DevstateVolumePost(http.ResponseWriter, *http.Request)
	DevstateVolumeVolumeNameDelete(http.ResponseWriter, *http.Request)
	DevstateVolumeVolumeNamePatch(http.ResponseWriter, *http.Request)
}

// DefaultApiServicer defines the api actions for the DefaultApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can be ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type DefaultApiServicer interface {
	ComponentCommandPost(context.Context, ComponentCommandPostRequest) (ImplResponse, error)
	ComponentGet(context.Context) (ImplResponse, error)
	DevfileGet(context.Context) (ImplResponse, error)
	DevfilePut(context.Context, DevfilePutRequest) (ImplResponse, error)
	InstanceDelete(context.Context) (ImplResponse, error)
	InstanceGet(context.Context) (ImplResponse, error)
	TelemetryGet(context.Context) (ImplResponse, error)
}

// DevstateApiServicer defines the api actions for the DevstateApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can be ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type DevstateApiServicer interface {
	DevstateApplyCommandPost(context.Context, DevstateApplyCommandPostRequest) (ImplResponse, error)
	DevstateChartGet(context.Context) (ImplResponse, error)
	DevstateCommandCommandNameDelete(context.Context, string) (ImplResponse, error)
	DevstateCommandCommandNameMovePost(context.Context, string, DevstateCommandCommandNameMovePostRequest) (ImplResponse, error)
	DevstateCommandCommandNameSetDefaultPost(context.Context, string, DevstateCommandCommandNameSetDefaultPostRequest) (ImplResponse, error)
	DevstateCommandCommandNameUnsetDefaultPost(context.Context, string) (ImplResponse, error)
	DevstateCompositeCommandPost(context.Context, DevstateCompositeCommandPostRequest) (ImplResponse, error)
	DevstateContainerContainerNameDelete(context.Context, string) (ImplResponse, error)
	DevstateContainerPost(context.Context, DevstateContainerPostRequest) (ImplResponse, error)
	DevstateDevfileDelete(context.Context) (ImplResponse, error)
	DevstateDevfileGet(context.Context) (ImplResponse, error)
	DevstateDevfilePut(context.Context, DevstateDevfilePutRequest) (ImplResponse, error)
	DevstateEventsPut(context.Context, DevstateEventsPutRequest) (ImplResponse, error)
	DevstateExecCommandPost(context.Context, DevstateExecCommandPostRequest) (ImplResponse, error)
	DevstateImageImageNameDelete(context.Context, string) (ImplResponse, error)
	DevstateImagePost(context.Context, DevstateImagePostRequest) (ImplResponse, error)
	DevstateMetadataPut(context.Context, MetadataRequest) (ImplResponse, error)
	DevstateQuantityValidPost(context.Context, DevstateQuantityValidPostRequest) (ImplResponse, error)
	DevstateResourcePost(context.Context, DevstateResourcePostRequest) (ImplResponse, error)
	DevstateResourceResourceNameDelete(context.Context, string) (ImplResponse, error)
	DevstateResourceResourceNamePatch(context.Context, string, DevstateResourceResourceNamePatchRequest) (ImplResponse, error)
	DevstateVolumePost(context.Context, DevstateVolumePostRequest) (ImplResponse, error)
	DevstateVolumeVolumeNameDelete(context.Context, string) (ImplResponse, error)
	DevstateVolumeVolumeNamePatch(context.Context, string, DevstateVolumeVolumeNamePatchRequest) (ImplResponse, error)
}
