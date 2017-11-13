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
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

type request struct {
	User  string `json:"user"`
	Realm string `json:"realm"`
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		cfg config
		err string
	}{
		{config{Address: ":80", BasePath: "/users", LogLevel: "INFO", Realm: "es"}, ""},
		{config{Address: "80", BasePath: "/users", LogLevel: "INFO", Realm: "es"}, `Invalid configuration according to JSON schema: address: Does not match pattern '^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})?:\d{2,4}$'`},
	}

	v := NewValidator()
	v.LoadSchemas("testdata/schemas")
	for _, test := range tests {
		if err := v.ValidateConfig("config", &test.cfg); err != nil {
			if err.Error() != test.err {
				t.Errorf("Invalid validation. Actual: %s. Expected: %s.", err, test.err)
			}
		} else {
			if test.err != "" {
				t.Errorf("Invalid validation. No error raised but expected: %s.", test.err)
			}
		}
	}
}

func TestValidateRequestBody(t *testing.T) {
	tests := []struct {
		body        string
		contentType string
		err         string
	}{
		{`{"name":"niji","realm":"es"}`, "application/json", ""},
		{`{"name":"niji"}`, "application/json", "Invalid request body"},
		{`"name":"niji"`, "application/json", "Invalid request body"},
		{`{"name":"niji","realm":"es"}`, "text/xml", "Invalid Content-Type header"},
	}

	v := NewValidator()
	v.LoadSchemas("testdata/schemas")

	for _, test := range tests {
		var req request
		r := httptest.NewRequest("POST", "/users", strings.NewReader(test.body))
		r.Header.Add("Content-Type", test.contentType)
		if err := v.ValidateRequestBody("request", r, &req); err != nil {
			if err.Error() != test.err {
				t.Errorf("Invalid request validation. Actual: %s. Expected: %s.", err, test.err)
			}
		} else {
			if test.err != "" {
				t.Errorf("Invalid request validation. No error raised but expected: %s.", test.err)
			}
		}
	}
}

func TestValidateSafeRequestBody(t *testing.T) {
	tests := []struct {
		body string
		err  string
	}{
		{`{"name":"niji","realm":"es"}`, ""},
		{`{"name":"niji"}`, "Invalid request body"},
	}

	v := NewValidator()
	v.LoadSchemas("testdata/schemas")

	for _, test := range tests {
		var req request
		r := httptest.NewRequest("POST", "/users", strings.NewReader(test.body))
		r.Header.Add("Content-Type", "application/json")
		if err := v.ValidateSafeRequestBody("request", r, &req); err != nil {
			if err.Error() != test.err {
				t.Errorf("Invalid request validation. Actual: %s. Expected: %s.", err, test.err)
			}
		} else {
			if test.err != "" {
				t.Errorf("Invalid request validation. No error raised but expected: %s.", test.err)
			}
		}
		data, _ := ioutil.ReadAll(r.Body)
		if test.body != string(data) {
			t.Errorf("The body cannot be read again: %s", string(data))
		}
	}
}
