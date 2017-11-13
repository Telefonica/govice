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
	"context"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestResponse(t *testing.T) {
	tests := []struct {
		err    *Error
		status int
		body   string
	}{
		{NewServerError("server error"), 500, `{"error":"server_error"}`},
		{NewBadGatewayError("gateway error"), 502, `{"error":"server_error"}`},
		{NewInvalidRequestError("log message", "invalid request"), 400, `{"error":"invalid_request","error_description":"invalid request"}`},
		{NewUnauthorizedClientError("log message", "unauthorized client"), 403, `{"error":"unauthorized_client","error_description":"unauthorized client"}`},
		{NotFoundError, 404, `{"error":"invalid_request","error_description":"not found"}`},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		test.err.Response(w)
		if w.Code != test.status {
			t.Errorf("Invalid status code. Actual: %d. Expected: %d", w.Code, test.status)
		}
		reader, _ := ioutil.ReadAll(w.Body)
		body := string(reader)
		if body != test.body {
			t.Errorf("Invalid body. Actual: %s. Expected: %s", body, test.body)
		}
	}
}

func TestGetResponse(t *testing.T) {
	tests := []struct {
		err    *Error
		status int
		body   string
	}{
		{NewServerError("server error"), 500, `{"error":"server_error"}`},
		{NewBadGatewayError("gateway error"), 502, `{"error":"server_error"}`},
		{NewInvalidRequestError("log message", "invalid request"), 400, `{"error":"invalid_request","error_description":"invalid request"}`},
		{NewUnauthorizedClientError("log message", "unauthorized client"), 403, `{"error":"unauthorized_client","error_description":"unauthorized client"}`},
		{NotFoundError, 404, `{"error":"invalid_request","error_description":"not found"}`},
	}
	for _, test := range tests {
		resp := test.err.GetResponse()
		if resp.StatusCode != test.status {
			t.Errorf("Invalid status code. Actual: %d. Expected: %d", resp.StatusCode, test.status)
		}
		reader, _ := ioutil.ReadAll(resp.Body)
		body := string(reader)
		if body != test.body {
			t.Errorf("Invalid body. Actual: %s. Expected: %s", body, test.body)
		}
	}
}

var alarmError = &Error{
	Message:     "log message",
	Status:      501,
	Alarm:       "ALARM_01",
	Code:        "invalid",
	Description: "alarm error",
}

func TestReplyWithError(t *testing.T) {
	tests := []struct {
		err      error
		status   int
		body     string
		expected string
	}{
		{errors.New("std error"), 500, `{"error":"server_error"}`, `,"lvl":"ERROR","msg":"std error"}`},
		{NewServerError("server error"), 500, `{"error":"server_error"}`, `,"lvl":"ERROR","msg":"server error"}`},
		{NewBadGatewayError("gateway error"), 502, `{"error":"server_error"}`, `,"lvl":"ERROR","msg":"gateway error"}`},
		{NewInvalidRequestError("log message", "invalid request"), 400, `{"error":"invalid_request","error_description":"invalid request"}`, `,"lvl":"INFO","msg":"log message"}`},
		{NewUnauthorizedClientError("log message", "unauthorized client"), 403, `{"error":"unauthorized_client","error_description":"unauthorized client"}`, `,"lvl":"INFO","msg":"log message"}`},
		{NotFoundError, 404, `{"error":"invalid_request","error_description":"not found"}`, `,"lvl":"INFO","msg":"not found"}`},
		{alarmError, 501, `{"error":"invalid","error_description":"alarm error"}`, `,"lvl":"ERROR","alarm":"ALARM_01","msg":"log message"}`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		logger := &Logger{out: &buf, logLevel: infoLevel}
		r := httptest.NewRequest("GET", "/users", nil)
		r = r.WithContext(context.WithValue(r.Context(), LoggerContextKey, logger))
		w := httptest.NewRecorder()
		ReplyWithError(w, r, test.err)
		if w.Code != test.status {
			t.Errorf("Invalid status code. Actual: %d. Expected: %d", w.Code, test.status)
		}
		reader, _ := ioutil.ReadAll(w.Body)
		body := string(reader)
		if body != test.body {
			t.Errorf("Invalid body. Actual: %s. Expected: %s", body, test.body)
		}
		expected := test.expected + "\n"
		if extractFirstField(buf.String()) != expected {
			t.Errorf("Invalid log. Actual: %s. Expected to end with: %s", buf.String(), expected)
		}
	}
}
