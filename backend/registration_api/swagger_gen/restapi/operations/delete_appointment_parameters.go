// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
<<<<<<< HEAD
	"context"
=======
>>>>>>> d67f4a22968fc0d8f5e31a903c140990031f5bbe
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
<<<<<<< HEAD
	"github.com/go-openapi/validate"
)

// NewDeleteAppointmentParams creates a new DeleteAppointmentParams object
//
// There are no default values defined in the spec.
=======
)

// NewDeleteAppointmentParams creates a new DeleteAppointmentParams object
// no default values defined in spec.
>>>>>>> d67f4a22968fc0d8f5e31a903c140990031f5bbe
func NewDeleteAppointmentParams() DeleteAppointmentParams {

	return DeleteAppointmentParams{}
}

// DeleteAppointmentParams contains all the bound params for the delete appointment operation
// typically these are obtained from a http.Request
//
// swagger:parameters deleteAppointment
type DeleteAppointmentParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: body
	*/
	Body DeleteAppointmentBody
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewDeleteAppointmentParams() beforehand.
func (o *DeleteAppointmentParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body DeleteAppointmentBody
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			res = append(res, errors.NewParseError("body", "body", "", err))
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

<<<<<<< HEAD
			ctx := validate.WithOperationRequest(context.Background())
			if err := body.ContextValidate(ctx, route.Formats); err != nil {
				res = append(res, err)
			}

=======
>>>>>>> d67f4a22968fc0d8f5e31a903c140990031f5bbe
			if len(res) == 0 {
				o.Body = body
			}
		}
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
