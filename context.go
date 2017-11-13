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

// LogContext represents the log context for a base service.
type LogContext struct {
	TransactionID string `json:"trans,omitempty"`
	Correlator    string `json:"corr,omitempty"`
	Operation     string `json:"op,omitempty"`
	Service       string `json:"svc,omitempty"`
	Component     string `json:"comp,omitempty"`
	User          string `json:"user,omitempty"`
	Realm         string `json:"realm,omitempty"`
	Alarm         string `json:"alarm,omitempty"`
}

// ReqLogContext is a complementary LogContext to log information about the request (e.g. path).
type ReqLogContext struct {
	Method     string `json:"method,omitempty"`
	Path       string `json:"path,omitempty"`
	RemoteAddr string `json:"remoteaddr,omitempty"`
}

// RespLogContext is a complementary LogContext to log information about the response (e.g. status code).
type RespLogContext struct {
	Status   int    `json:"status,omitempty"`
	Latency  int    `json:"latency,omitempty"`
	Location string `json:"location,omitempty"`
}
