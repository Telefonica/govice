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

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/telefonica/govice"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world")
}

func main() {
	// Force UTC time zone (used in time field of the log records)
	time.Local = time.UTC
	// Create the context for the logger instance
	ctxt := govice.LogContext{Service: "logger", Component: "demo"}

	http.HandleFunc("/", govice.WithLogContext(&ctxt)(govice.WithLog(handler)))
	http.ListenAndServe(":8080", nil)
}
