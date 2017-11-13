/**
 * @license
 * Copyright 2017 Telefónica Investigación y Desarrollo, S.A.U
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package govice

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Error is a custom error. This struct stores information to generate an HTTP error response if required.
type Error struct {
	Message     string `json:"-"`
	Status      int    `json:"-"`
	Alarm       string `json:"-"`
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// Response generates a JSON document for an Error.
// JSON is in the form: {"error": "invalid_request", "error_description": "xxx"}
func (e *Error) Response(w http.ResponseWriter) {
	data, err := json.Marshal(e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	w.Write(data)
}

// GetResponse to get a http.Response object from an Error.
func (e *Error) GetResponse() *http.Response {
	r := &http.Response{}
	if data, err := json.Marshal(e); err != nil {
		r.StatusCode = http.StatusInternalServerError
	} else {
		r.Header = make(http.Header)
		r.Header.Add("Content-Type", "application/json")
		r.StatusCode = e.Status
		r.Body = ioutil.NopCloser(bytes.NewReader(data))
	}
	return r
}

// NewServerError to create a server_error Error
func NewServerError(message string) *Error {
	return &Error{
		Message: message,
		Status:  http.StatusInternalServerError,
		Code:    "server_error",
	}
}

// NewBadGatewayError to create a bad gateway error.
func NewBadGatewayError(message string) *Error {
	return &Error{
		Message: message,
		Status:  http.StatusBadGateway,
		Code:    "server_error",
	}
}

// NewInvalidRequestError to create an invalid_request Error
func NewInvalidRequestError(message string, description string) *Error {
	return &Error{
		Message:     message,
		Status:      http.StatusBadRequest,
		Code:        "invalid_request",
		Description: description,
	}
}

// NewUnauthorizedClientError to create an unauthorized_client Error
func NewUnauthorizedClientError(message string, description string) *Error {
	return &Error{
		Message:     message,
		Status:      http.StatusForbidden,
		Code:        "unauthorized_client",
		Description: description,
	}
}

// NotFoundError with a not found Error
var NotFoundError = &Error{
	Message:     "not found",
	Status:      http.StatusNotFound,
	Code:        "invalid_request",
	Description: "not found",
}

// ReplyWithError to send a HTTP response with the error document.
func ReplyWithError(w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) {
	case *Error:
		if e.Status >= http.StatusBadRequest && e.Status < http.StatusInternalServerError {
			GetLogger(r).Info(err.Error())
		} else if e.Alarm != "" {
			logContext := LogContext{Alarm: e.Alarm}
			GetLogger(r).ErrorC(logContext, err.Error())
		} else {
			GetLogger(r).Error(err.Error())
		}
		e.Response(w)
	default:
		GetLogger(r).Error(err.Error())
		NewServerError("").Response(w)
	}
}
