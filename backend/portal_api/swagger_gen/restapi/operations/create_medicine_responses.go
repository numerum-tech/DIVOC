// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// CreateMedicineOKCode is the HTTP code returned for type CreateMedicineOK
const CreateMedicineOKCode int = 200

/*CreateMedicineOK OK

swagger:response createMedicineOK
*/
type CreateMedicineOK struct {
}

// NewCreateMedicineOK creates CreateMedicineOK with default headers values
func NewCreateMedicineOK() *CreateMedicineOK {

	return &CreateMedicineOK{}
}

// WriteResponse to the client
func (o *CreateMedicineOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// CreateMedicineBadRequestCode is the HTTP code returned for type CreateMedicineBadRequest
const CreateMedicineBadRequestCode int = 400

/*CreateMedicineBadRequest Invalid input

swagger:response createMedicineBadRequest
*/
type CreateMedicineBadRequest struct {
}

// NewCreateMedicineBadRequest creates CreateMedicineBadRequest with default headers values
func NewCreateMedicineBadRequest() *CreateMedicineBadRequest {

	return &CreateMedicineBadRequest{}
}

// WriteResponse to the client
func (o *CreateMedicineBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// CreateMedicineUnauthorizedCode is the HTTP code returned for type CreateMedicineUnauthorized
const CreateMedicineUnauthorizedCode int = 401

/*CreateMedicineUnauthorized Unauthorized

swagger:response createMedicineUnauthorized
*/
type CreateMedicineUnauthorized struct {
}

// NewCreateMedicineUnauthorized creates CreateMedicineUnauthorized with default headers values
func NewCreateMedicineUnauthorized() *CreateMedicineUnauthorized {

	return &CreateMedicineUnauthorized{}
}

// WriteResponse to the client
func (o *CreateMedicineUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}
