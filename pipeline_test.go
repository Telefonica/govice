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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPipeline(t *testing.T) {
	mws := []func(http.HandlerFunc) http.HandlerFunc{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("mw1", "hello")
				next(w, r)
			}
		},
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("mw2", "world!")
				next(w, r)
			}
		},
	}
	endpoint := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users", nil)
	pipeline := Pipeline(mws, endpoint)
	pipeline.ServeHTTP(w, r)
	if w.Code != 201 {
		t.Errorf("endpoint was not executed correctly")
	}
	if w.Header().Get("mw1") != "hello" || w.Header().Get("mw2") != "world!" {
		t.Errorf("middlewares were not executed correctly")
	}
}
