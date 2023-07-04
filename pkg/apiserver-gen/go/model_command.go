/*
 * odo dev
 *
 * API interface for 'odo dev'
 *
 * API version: 0.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type Command struct {
	Name string `json:"name"`

	Group string `json:"group"`

	Default bool `json:"default,omitempty"`

	Type string `json:"type"`

	Exec ExecCommand `json:"exec,omitempty"`

	Apply ApplyCommand `json:"apply,omitempty"`

	Image ImageCommand `json:"image,omitempty"`

	Composite CompositeCommand `json:"composite,omitempty"`
}

// AssertCommandRequired checks if the required fields are not zero-ed
func AssertCommandRequired(obj Command) error {
	elements := map[string]interface{}{
		"name":  obj.Name,
		"group": obj.Group,
		"type":  obj.Type,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	if err := AssertExecCommandRequired(obj.Exec); err != nil {
		return err
	}
	if err := AssertApplyCommandRequired(obj.Apply); err != nil {
		return err
	}
	if err := AssertImageCommandRequired(obj.Image); err != nil {
		return err
	}
	if err := AssertCompositeCommandRequired(obj.Composite); err != nil {
		return err
	}
	return nil
}

// AssertRecurseCommandRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of Command (e.g. [][]Command), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseCommandRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aCommand, ok := obj.(Command)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertCommandRequired(aCommand)
	})
}
