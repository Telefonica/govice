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

// Context to support extending the log context with other parameters.
type Context interface {
	Clone() Context
	GetCorrelator() string
	SetCorrelator(corr string)
	GetTransactionID() string
	SetTransactionID(trans string)
}

// LogContext represents the log context for a base service.
// Note that LogContext implements the Context interface.
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

// Clone the log context.
func (c *LogContext) Clone() Context {
	a := *c
	return &a
}

// GetCorrelator returns the log context correlator.
func (c *LogContext) GetCorrelator() string {
	return c.Correlator
}

// SetCorrelator to set a correlator in the log context.
func (c *LogContext) SetCorrelator(corr string) {
	c.Correlator = corr
}

// GetTransactionID returns the log context transactionID (trans).
func (c *LogContext) GetTransactionID() string {
	return c.TransactionID
}

// SetTransactionID to set a transactionID in the log context.
func (c *LogContext) SetTransactionID(trans string) {
	c.TransactionID = trans
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
