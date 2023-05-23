/*
 * odo dev
 *
 * API interface for 'odo dev'
 *
 * API version: 0.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ComponentGet200Response struct {

	// Description of the component. This is the same as output of 'odo describe component -o json'
	Component map[string]interface{} `json:"component,omitempty"`
}

// AssertComponentGet200ResponseRequired checks if the required fields are not zero-ed
func AssertComponentGet200ResponseRequired(obj ComponentGet200Response) error {
	return nil
}

// AssertRecurseComponentGet200ResponseRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of ComponentGet200Response (e.g. [][]ComponentGet200Response), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseComponentGet200ResponseRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aComponentGet200Response, ok := obj.(ComponentGet200Response)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertComponentGet200ResponseRequired(aComponentGet200Response)
	})
}
