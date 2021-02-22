// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// UserGroup user group
//
// swagger:model UserGroup
type UserGroup struct {

	// group id
	ID string `json:"id,omitempty"`

	// group name
	Name string `json:"name,omitempty"`
}

// Validate validates this user group
func (m *UserGroup) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this user group based on context it is used
func (m *UserGroup) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *UserGroup) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UserGroup) UnmarshalBinary(b []byte) error {
	var res UserGroup
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
