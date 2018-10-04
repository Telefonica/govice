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
	"encoding/json"
	"net/http"
)

// WriteJSON generates a HTTP response by marshalling an object to JSON.
// If the marshalling fails, it generates a govice error.
// It also sets the HTTP header "Content-Type" to "application/json".
func WriteJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		ReplyWithError(w, r, err)
	}
}
