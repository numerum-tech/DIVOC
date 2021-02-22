// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// SideEffectsResponse SideEffectsResponse
//
// Indian address format
//
// swagger:model sideEffectsResponse
type SideEffectsResponse struct {

	// response
	//
	// response
	// Required: true
	Response *string `json:"response"`

	// symptom
	//
	// symptom
	// Required: true
	Symptom *string `json:"symptom"`
}

// Validate validates this side effects response
func (m *SideEffectsResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateResponse(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSymptom(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SideEffectsResponse) validateResponse(formats strfmt.Registry) error {

	if err := validate.Required("response", "body", m.Response); err != nil {
		return err
	}

	return nil
}

func (m *SideEffectsResponse) validateSymptom(formats strfmt.Registry) error {

	if err := validate.Required("symptom", "body", m.Symptom); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this side effects response based on context it is used
func (m *SideEffectsResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *SideEffectsResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SideEffectsResponse) UnmarshalBinary(b []byte) error {
	var res SideEffectsResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
