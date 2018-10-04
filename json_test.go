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
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tcs := []struct {
		v            interface{}
		expectedBody string
		expectedCode int
	}{
		{&LogContext{TransactionID: "txid-01", Correlator: "corr-02"}, `{"trans":"txid-01","corr":"corr-02"}` + "\n", 200},
		{make(chan int), `{"error":"server_error"}`, 500},
	}
	for _, tc := range tcs {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/users", nil)
		WriteJSON(w, r, tc.v)
		if w.Body.String() != tc.expectedBody {
			t.Errorf("Invalid JSON body. Expected: %s. Got: %s.", tc.expectedBody, w.Body)
		}
		if w.Code != tc.expectedCode {
			t.Errorf("Invalid status code. Expected: %d. Got: %d.", tc.expectedCode, w.Code)
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Invalid Content-Type HTTP header. Got: %s.", w.Header().Get("Content-Type"))
		}
	}
}
