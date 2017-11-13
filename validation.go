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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// Validator type.
type Validator struct {
	schemas map[string]*gojsonschema.Schema
}

// NewValidator is the constructor for Validator.
func NewValidator() *Validator {
	return &Validator{schemas: make(map[string]*gojsonschema.Schema)}
}

// LoadSchemas to load all the JSON schemas stored in schemasDir directory (it may be an absolute path or relative to
// the current working directory)
func (v *Validator) LoadSchemas(schemasDir string) error {
	schemasPath, err := getAbsolutePath(schemasDir)
	if err != nil {
		return fmt.Errorf("Error getting schemas directory: %s", err)
	}
	files, err := ioutil.ReadDir(schemasPath)
	if err != nil {
		return fmt.Errorf("Error reading schemas directory: %s. %s", schemasPath, err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			schemaURI := "file://" + path.Join(schemasPath, file.Name())
			schemaLoader := gojsonschema.NewReferenceLoader(schemaURI)
			schema, err := gojsonschema.NewSchema(schemaLoader)
			if err != nil {
				return fmt.Errorf("Invalid JSON schema file: %s. %s", schemaURI, err)
			}
			schemaName := strings.TrimSuffix(file.Name(), ".json")
			v.schemas[schemaName] = schema
		}
	}
	return nil
}

// ValidateConfig to validate the configuration against config.json schema.
func (v *Validator) ValidateConfig(schemaName string, config interface{}) error {
	if err := v.ValidateObject(schemaName, config); err != nil {
		return fmt.Errorf("Invalid configuration according to JSON schema: %s", err)
	}
	return nil
}

// ValidateRequestBody reads the request body, validates it against a JSON schema, and unmarshall it.
func (v *Validator) ValidateRequestBody(schemaName string, r *http.Request, o interface{}) error {
	return v.validateRequestBody(schemaName, r, o, false)
}

// ValidateSafeRequestBody reads the request body, validates it against a JSON schema, and unmarshall it.
// Unlike ValidateRequestBody, it maintains the body in the request object (e.g. to be forwarded via proxy).
func (v *Validator) ValidateSafeRequestBody(schemaName string, r *http.Request, o interface{}) error {
	return v.validateRequestBody(schemaName, r, o, true)
}

func (v *Validator) validateRequestBody(schemaName string, r *http.Request, o interface{}, safe bool) error {
	// Get the request body and load it to be validated with the JSON schema
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Error reading the request body. %s", err)
	}

	// Force that body can be read again from request object
	if safe {
		r.Body = ioutil.NopCloser(bytes.NewReader(data))
	}

	// Check if the content type is application/json
	contentTypeHeader := r.Header["Content-Type"]
	if contentTypeHeader != nil && len(contentTypeHeader) == 1 {
		contentType := contentTypeHeader[0]
		if contentType == "application/json" || strings.HasPrefix(contentType, "application/json;") {
			if err := v.ValidateBytes(schemaName, data, o); err != nil {
				return NewInvalidRequestError("Invalid request body", err.Error())
			}
			return nil
		}
	}
	logMsg := "Invalid Content-Type header"
	errorDescription := "content-type header must be application/json"
	return NewInvalidRequestError(logMsg, errorDescription)
}

// ValidateObject to validate an object against a JSON schema.
func (v *Validator) ValidateObject(schemaName string, data interface{}) error {
	documentLoader := gojsonschema.NewGoLoader(data)
	return v.validate(schemaName, documentLoader)
}

// ValidateBytes reads a byte array, validates it against a JSON schema, and unmarshall it.
func (v *Validator) ValidateBytes(schemaName string, data []byte, o interface{}) error {
	documentLoader := gojsonschema.NewStringLoader(string(data))
	if err := v.validate(schemaName, documentLoader); err != nil {
		return err
	}
	// Validation succeeded
	if err := json.Unmarshal(data, o); err != nil {
		return fmt.Errorf("Error unmarshalling data. %s", err)
	}
	return nil
}

// validate validates a document (documentLoader) with a schema.
func (v *Validator) validate(schemaName string, documentLoader gojsonschema.JSONLoader) error {
	// Retrieve the JSON schema
	schema := v.schemas[schemaName]
	if schema == nil {
		return fmt.Errorf("schema %s not found", schemaName)
	}
	// Validate document against JSON schema
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return err
	}
	if !result.Valid() {
		if len(result.Errors()) > 0 {
			return fmt.Errorf("%s", result.Errors()[0])
		}
		return fmt.Errorf("invalid")
	}
	return nil
}

// getAbsolutePath returns an absolute path. If relpath is absolute, it returns the same value. If relpath is relative, it
// returns an absolute path relative to current working directory.
func getAbsolutePath(relpath string) (string, error) {
	if path.IsAbs(relpath) {
		return relpath, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Error getting current working directory. %s", err)
	}
	return path.Join(cwd, relpath), nil
}
