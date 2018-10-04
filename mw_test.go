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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithLogContext(t *testing.T) {
	tests := []struct {
		corr string
	}{
		{""},
		{"test corr"},
	}

	ctxt := LogContext{}
	withLogContext := WithLogContext(&ctxt)

	for _, test := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/users", nil)
		r.Header.Add("Unica-Correlator", test.corr)
		handler := func(w http.ResponseWriter, r *http.Request) {
			context := GetLogContext(r)
			if context.TransactionID == "" {
				t.Errorf("Invalid transaction: %s", context.TransactionID)
			}
			if test.corr == "" {
				if context.TransactionID != context.Correlator {
					t.Errorf("Invalid correlator for context: %+v", context)
				}
			} else if test.corr != context.Correlator || context.Correlator != r.Header.Get("Unica-Correlator") {
				t.Errorf("Correlator is not maintained. Header: %s. Context: %s", r.Header.Get("Unica-Correlator"), context.Correlator)
			}
			w.WriteHeader(http.StatusOK)
		}
		withLogContext(http.HandlerFunc(handler))(w, r)
	}
}

func TestWithLog(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users", nil)
	r.Header.Add("Unica-Correlator", "corr")
	var buf bytes.Buffer
	var ctxt LogContext
	logger := NewLogger()
	logger.SetLogContext(InitContext(r, &ctxt))
	logger.out = &buf
	r = r.WithContext(context.WithValue(r.Context(), LoggerContextKey, logger))
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	WithLog(http.HandlerFunc(handler))(w, r)

	records := strings.Split(buf.String(), "\n")
	if len(records) != 3 {
		t.Errorf("Error in log records: %s", buf.String())
		return
	}
	type reqLog struct {
		Level      string `json:"lvl"`
		Correlator string `json:"corr"`
		Method     string `json:"method"`
		Path       string `json:"path"`
		Message    string `json:"msg"`
	}
	actualReqLog := reqLog{}
	if err := json.Unmarshal([]byte(records[0]), &actualReqLog); err != nil {
		t.Errorf("Error processing log record: %s. %s", records[0], err)
		return
	}
	expectedReqLog := reqLog{"INFO", "corr", "GET", "/users", "Request"}
	if actualReqLog != expectedReqLog {
		t.Errorf("Invalid request log. Actual: %+v. Expected: %+v", actualReqLog, expectedReqLog)
	}

	type respLog struct {
		Level      string `json:"lvl"`
		Correlator string `json:"corr"`
		Status     int    `json:"status"`
		Message    string `json:"msg"`
	}
	actualRespLog := respLog{}
	if err := json.Unmarshal([]byte(records[1]), &actualRespLog); err != nil {
		t.Errorf("Error processing log record: %s. %s", records[1], err)
		return
	}
	expectedRespLog := respLog{"INFO", "corr", 200, "Response"}
	if actualRespLog != expectedRespLog {
		t.Errorf("Invalid response log. Actual: %+v. Expected: %+v", actualRespLog, expectedRespLog)
	}
}

func TestWithMethodNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users", nil)
	withMethodNotAllowed := WithMethodNotAllowed("POST", "DELETE")
	withMethodNotAllowed(w, r)

	actual := w.Header().Get("Allow")
	expected := "POST, DELETE"
	if actual != expected {
		t.Errorf("Invalid allow header. Actual: %s. Expected:%s", actual, expected)
	}
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Invalid status code. Actual %d. Expected %d.", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestWithNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users", nil)
	WithNotFound()(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Invalid status code. Actual %d. Expected %d.", w.Code, http.StatusNotFound)
	}
}
