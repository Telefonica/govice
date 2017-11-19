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
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/imdario/mergo"
)

// CorrelatorHTTPHeader contains the name of the HTTP header that transports the correlator.
// The correlator enables to match all the HTTP requests and responses for a same web flow.
var (
	CorrelatorHTTPHeader = "Unica-Correlator"
	RequestLogMessage    = "Request"
	ResponseLogMessage   = "Response"
)

type loggerContextKey string

// LoggerContextKey is a unique key to store the logger in the golang context.
var LoggerContextKey = loggerContextKey("logger")

// LoggableResponseWriter is a ResponseWriter wrapper to log the response status code.
type LoggableResponseWriter struct {
	Status int
	http.ResponseWriter
}

// WriteHeader overwrites ResponseWriter's WriteHeader to store the response status code.
func (w *LoggableResponseWriter) WriteHeader(statusCode int) {
	w.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func newTransactionID() string {
	UUID, err := uuid.NewUUID()
	if err != nil {
		return ""
	}
	return UUID.String()
}

func newContextLogger(r *http.Request, ctx *LogContext) *Logger {
	logger := NewLogger()
	newCtx := NewType(ctx).(*LogContext)
	logger.SetLogContext(newCtx)
	if err := mergo.Merge(newCtx, ctx); err != nil {
		return logger
	}
	newCtx.TransactionID = newTransactionID()
	if newCtx.Correlator = r.Header.Get(CorrelatorHTTPHeader); newCtx.Correlator == "" {
		newCtx.Correlator = newCtx.TransactionID
		r.Header.Add(CorrelatorHTTPHeader, newCtx.Correlator)
	}
	return logger
}

// WithLogContext is a middleware constructor to initialize the log context with the
// transactionID and correlator. It also stores the logger in the golang context.
// Note that the context is initialized with an initial context (see ctx).
func WithLogContext(ctx *LogContext) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logger := newContextLogger(r, ctx)
			next(w, r.WithContext(context.WithValue(r.Context(), LoggerContextKey, logger)))
		}
	}
}

// WithLog is a middleware to log the request and response.
// Note that WithContext middleware is required to initialize the logger with a context.
func WithLog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		logger := GetLogger(r)
		isNewLogger := false
		if logger == nil {
			logger = newContextLogger(r, &LogContext{})
			isNewLogger = true
		}
		logContext := logger.GetLogContext().(*LogContext)
		reqContext := ReqLogContext{
			Path:       r.RequestURI,
			Method:     r.Method,
			RemoteAddr: r.RemoteAddr,
		}
		logger.InfoC(reqContext, RequestLogMessage)
		logger.DebugRequest(RequestLogMessage, r)
		lw := &LoggableResponseWriter{Status: http.StatusOK, ResponseWriter: w}
		lw.Header().Set(CorrelatorHTTPHeader, logContext.Correlator)
		if isNewLogger {
			next(lw, r.WithContext(context.WithValue(r.Context(), LoggerContextKey, logger)))
		} else {
			next(lw, r)
		}
		respContext := RespLogContext{
			Status:   lw.Status,
			Latency:  int(time.Since(now).Nanoseconds() / 1000000),
			Location: lw.Header().Get("Location"),
		}
		logger.InfoC(respContext, ResponseLogMessage)
	}
}

// WithMethodNotAllowed is a middleware to reply with an error when the HTTP method is not supported.
// The allowedMethods must be a list of HTTP methods.
func WithMethodNotAllowed(allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// WithNotFound is a middleware to reply with a not found error (status code: 404).
func WithNotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		NotFoundError.Response(w)
	}
}

// GetLogger to get the logger from the request context.
func GetLogger(r *http.Request) *Logger {
	return r.Context().Value(LoggerContextKey).(*Logger)
}

// GetLogContext gets the log context associated to a request.
func GetLogContext(r *http.Request) *LogContext {
	if logger := GetLogger(r); logger != nil {
		return logger.GetLogContext().(*LogContext)
	}
	return nil
}
