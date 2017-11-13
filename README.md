[![Build Status](https://api.travis-ci.org/Telefonica/govice.svg?branch=master)](https://travis-ci.org/Telefonica/govice)

# govice

Libraries to **serve and protect** your services implemented in golang.

It provides the following functionality:

 - **Configuration**. Read and marshal the configuration read from a JSON file into a struct, and override it with environment variables.
 - **Validation**. Validate your configuration and requests with JSON schemas.
 - **Logging**. Log in JSON format with custom context objects.
 - **Middlewares**. Some http.HandlerFunc middlewares (e.g. to log your requests and responses automatically).
 - **Errors and alarms**.

See **examples** directory with executable applications of these features. The **service** example is a combination of all the features provided by govice.

## Configuration

The approach selected by govice to configure an application/service is based on a default configuration (using a JSON file) that can be override partially with environment variables. This approach is compliant with [The Twelve-Factor App](https://12factor.net/config). It is also very convenient when working with docker containers.

The following example loads **config.json** file into a **Config** struct and overrides the values with environment variables (if registered).

```go
package main

import (
	"fmt"
	"os"

	"github.com/Telefonica/govice"
)

type config struct {
	Address  string `json:"address" env:"ADDRESS"`
	BasePath string `json:"basePath" env:"BASE_PATH"`
	LogLevel string `json:"logLevel" env:"LOG_LEVEL"`
}

func main() {
	var cfg config
	if err := govice.GetConfig("config.json", &cfg); err != nil {
		panic("Invalid configuration.", err)
	}
	fmt.Printf("%+v\n", cfg)
}
```

The function `func GetConfig(configFile string, cfg interface{}) error` receives two parameters: a) path to the JSON configuration file (relative to the execution directory), b) reference to the configuration instance.

The configuration struct uses struct tags to map each field with a JSON element (using the tag **json**) and/or and environment variable (using the tag **env**). The **env** struct tag is implemented by [github.com/caarlos0/env](https://github.com/caarlos0/env) dependency.

## Validation

The govice validation is based on [JSON schemas](http://json-schema.org/) with the library [github.com/xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema). Its main goal is to avoid including this logic as part of the code. This separation of concerns makes the source code more readable and easier to maintain it.

It is recommended to validate any input to the service: a) configuration, b) requests to our service, c) responses received by our clients. Validation leads to safeness and robusness.

The following example extends the configuration example to validate the configuration:

```go
func main() {
	var cfg Config
	if err := govice.GetConfig("config.json", &cfg); err != nil {
		os.Exit(1)
	}
	fmt.Prinft("%+v", cfg)

	validator := govice.NewValidator()
	if err := validator.LoadSchemas("schemas"); err != nil {
		panic(err)
	}
	if err := validator.ValidateConfig("config", &cfg); err != nil {
		panic(err)
	}
	fmt.Println("Configuration validated successfully")
}
```

`func (v *Validator) LoadSchemas(schemasDir string) error` loads all the JSON schemas located in the `schemasDir` directory. Note that this directory may be relative to the execution directory. Each JSON schemas is loaded and indexed with the file name removing the **json** extension. For example, a JSON schema stored as **schemas/config.json** is loaded with the key **config**.

Then it is possible to validate the configuration (stored in a struct) against a JSON schema (using as key the JSON schema filename without extension). `func (v *Validator) ValidateConfig(schemaName string, cfg interface{}) error` validates a configuration object and generates errors aligned to configuration.

The `Validator` type also provides other methods to validate requests, objects or arrays:

 - `func (v *Validator) ValidateRequestBody(schemaName string, r *http.Request, o interface{}) error`. Validates the request body against a JSON schema and unmarshals it into an object.
 - `func (v *Validator) ValidateSafeRequestBody(schemaName string, r *http.Request, o interface{}) error`. As ValidateRequestBody, it also validates the request body against a JSON schema and unmarshals it into an object, but it maintains the request body to be read afterwards. This can be useful if it is required to forward the same request via proxy.
 - `func (v *Validator) ValidateObject(schemaName string, data interface{}) error`. It validates an object against a JSON schema.
 - `func (v *Validator) ValidateBytes(schemaName string, data []byte, o interface{}) error`. It reads a byte array, validates it against a JSON schema, and unmarshall it. This method is used by both **ValidateRequestBody** and **ValidateSafeRequestBody**.

## Logging

Logging writes log records to console using a JSON format to make easier that log aggregators (e.g. splunk) process them.

Logging requires to create an instance of **Logger** (e.g. with `func NewLogger() *Logger`). It is possible to set a log level: `func (l *Logger) SetLevel(levelName string)`. Possible log levels are: **DEBUG**, **INFO**, **WARN**, **ERROR**, and **FATAL**.

The following fields are always written:

| Field | Description |
|---|---|
| time | Timestamp when the log record was registered |
| lvl | Log level: **DEBUG**, **INFO**, **WARN**, **ERROR**, and **FATAL**. |
| msg | Log message |

These fields can be enhanced by using log contexts. A log context includes additional fields in the log record. There are 2 different log contexts: a) context at logger instance which is set with `func (l *Logger) SetLogContext(context interface{})`, and b) context at log record. Log contexts are structs that are marshalled into the log record (it is required to use the json struct tags to make them be marshalled).

Each log level provides two methods: with and without log context (note that a log context at logger instance is complementary).

| Level | Log without context | Log with context |
| ----- | ------------------- | ---------------- |
| DEBUG | `func (l *Logger) Debug(message string, args ...interface{})` | `func (l *Logger) DebugC(context interface{}, message string, args ...interface{})` |
| INFO | `func (l *Logger) Info(message string, args ...interface{})` | `func (l *Logger) InfoC(context interface{}, message string, args ...interface{})` |
| WARN | `func (l *Logger) Warn(message string, args ...interface{})` | `func (l *Logger) WarnC(context interface{}, message string, args ...interface{})` |
| ERROR | `func (l *Logger) Error(message string, args ...interface{})` | `func (l *Logger) ErrorC(context interface{}, message string, args ...interface{})` |
| FATAL | `func (l *Logger) Fatal(message string, args ...interface{})` | `func (l *Logger) FatalC(context interface{}, message string, args ...interface{})` |

There are several context struct defined in govice to work with HTTP:

```go
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
```

The following example demonstrates how to use the govice logger and the contexts:

```go
package main

import (
	"time"

	"github.com/Telefonica/govice"
)

type demoContext struct {
	Feature int `json:"feat,omitempty"`
}

func main() {
	// Force UTC time zone (used in time field of the log records)
	time.Local = time.UTC
	// Create the context for the logger instance
	ctxt := govice.LogContext{Service: "logger", Component: "demo"}

	logger := govice.NewLogger()
	logger.SetLogContext(ctxt)
	logger.Info("Logging without context")
	logger.Warn("Logging with %d %s", 2, "arguments")

	recordCtxt := &demoContext{Feature: 3}
	logger.InfoC(recordCtxt, "Logging with context")
}
```

The output of the previous command is:

```
{"time":"2017-11-12T23:29:55.929198687Z","lvl":"INFO","svc":"logger","comp":"demo","msg":"Logging without context"}
{"time":"2017-11-12T23:29:55.929333925Z","lvl":"WARN","svc":"logger","comp":"demo","msg":"Logging with 2 arguments"}
{"time":"2017-11-12T23:29:55.929355491Z","lvl":"INFO","svc":"logger","comp":"demo","feat":3,"msg":"Logging with context"}
```

Note that the logger context supports other data types beyond strings. This is really important to build up metrics based on logs.

There are additional utilities to dump requests and responses (in **DEBUG** level):

| Method | Description |
| ------ | ----------- |
| `func (l *Logger) DebugResponse(message string, r *http.Response)` | Dump a response |
| `func (l *Logger) DebugRequest(message string, r *http.Request)` | Dump a request received by the app |
| `func (l *Logger) DebugRequestOut(message string, r *http.Request)` | Dump a request generated by the app |

Note that these methods already have a version with context (e.g. **DebugResponseC**).

## Middlewares

| Middleware | Description |
| ---------- | ----------- |
| WithLogContext(ctx *LogContext) | Creates a logger (stored in the context of the request) and prepares the transactionID and correlator in the log context. It is also responsible to include the HTTP header for the correlator in both request and response. |
| WithLog | It logs the request and response |
| WithMethodNotAllowed(allowedMethods []string) | Generates a response with the **Allow** header with the allowed HTTP methods. |
| WithNotFound | Replies with a 404 error |

There are some important utilities related to **WithLogContext**. `func GetLogger(r *http.Request) *Logger` returns the logger created by the **WithLogContext** middleware. `func GetLogContext(r *http.Request) *LogContext` returns the log context from the previous logger.

The following example creates a web server where every request and response is logged in the console. This is achieved by concatenating **WithLogContext** and **WithLog** middlewares.

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Telefonica/govice"
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
```

This example creates the following log records:

```
{"time":"2017-11-13T08:01:51.335480728Z","lvl":"INFO","trans":"e7fc31ab-c848-11e7-8ed5-186590e007bb","corr":"e7fc31ab-c848-11e7-8ed5-186590e007bb","svc":"logger","comp":"demo","method":"GET","path":"/","remoteaddr":"[::1]:49636","msg":"Request"}
{"time":"2017-11-13T08:01:51.335546998Z","lvl":"INFO","trans":"e7fc31ab-c848-11e7-8ed5-186590e007bb","corr":"e7fc31ab-c848-11e7-8ed5-186590e007bb","svc":"logger","comp":"demo","status":200,"msg":"Response"}
```

Note that the log context passed to **WithLogContext** middleware must follow the type **govice.LogContext**. This is required because the middleware sets the transactionID and correlator in this context.

## Errors and alarms

This library defines some custom errors. Errors store information for logging, and to generate the HTTP response.

```go
type Error struct {
	Message     string `json:"-"`
	Alarm       string `json:"-"`
	Status      int    `json:"-"`
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
}
```

| Field | Description |
| ----- | ----------- |
| Message | Message to be logged |
| Alarm | It identifies an optional alarm identifier to be included in the log record if the error requires to trigger an alarm for ops. |
| Status | Status code of the HTTP response |
| Code | Error identifier (or error type). It corresponds to the **error** field in the JSON response body |
| Description | Description of the error (optional). It corresponds to the **error_description** field in the JSON response body |

The format of the response body complies with the error format defined by [OAuth2 standard](https://tools.ietf.org/html/rfc6749#section-5.2).

The function `func ReplyWithError(w http.ResponseWriter, r *http.Request, err error)` has two responsibilities:

 - It generates a HTTP response using a standard error. If the error is of type **govice.Error**, then it is casted to retrieve all the information; otherwise, it replies with a server error.
 - It also logs the error using the logger in the request context. Note that it depends on the **WithLogContext** middleware. If the status code associated to the error is 4xx, then it is logged with **INFO** level; otherwise, with **ERROR** level. If the error contains an alarm identifier, it is also logged.

## License

Copyright 2017 [Telefónica Investigación y Desarrollo, S.A.U](http://www.tid.es)

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
