// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProgramUpdateRequest program update request
//
// swagger:model ProgramUpdateRequest
type ProgramUpdateRequest struct {
	ProgramRequest

	// osid
	Osid string `json:"osid,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *ProgramUpdateRequest) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 ProgramRequest
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.ProgramRequest = aO0

	// AO1
	var dataAO1 struct {
		Osid string `json:"osid,omitempty"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.Osid = dataAO1.Osid

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m ProgramUpdateRequest) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.ProgramRequest)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)
	var dataAO1 struct {
		Osid string `json:"osid,omitempty"`
	}

	dataAO1.Osid = m.Osid

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)
	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this program update request
func (m *ProgramUpdateRequest) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with ProgramRequest
	if err := m.ProgramRequest.Validate(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validate this program update request based on the context it is used
func (m *ProgramUpdateRequest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with ProgramRequest
	if err := m.ProgramRequest.ContextValidate(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// MarshalBinary interface implementation
func (m *ProgramUpdateRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProgramUpdateRequest) UnmarshalBinary(b []byte) error {
	var res ProgramUpdateRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
