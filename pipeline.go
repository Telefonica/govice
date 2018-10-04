/**
 * @license
 * Copyright 2018 Telefónica Investigación y Desarrollo, S.A.U
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

import "net/http"

// Pipeline returns a HandlerFunc resulting of a list of middlewares and a final endpoint.
//
// The following example creates a pipeline of 2 middlewares:
//
//	mws := []func(http.HandlerFunc) http.HandlerFunc{
// 		govice.WithLogContext(&logContext),
// 		govice.WithLog,
// 	}
//	p := govice.Pipeline(mws, next)
//
func Pipeline(mws []func(http.HandlerFunc) http.HandlerFunc, endpoint http.HandlerFunc) http.HandlerFunc {
	h := endpoint
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
