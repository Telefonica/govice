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
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/caarlos0/env"
	"github.com/imdario/mergo"
)

// NewType creates a new object with the same type using reflection.
// Note that the new object is empty.
func NewType(orig interface{}) interface{} {
	val := reflect.ValueOf(orig)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	return reflect.New(val.Type()).Interface()
}

func loadConfigFile(configFile string, config interface{}) error {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, config)
}

// GetConfig prepares the configuration by merging multiple sources:
// - Default configuration stored in a json file
// - Environment variables
func GetConfig(configFile string, config interface{}) error {
	// Get the environment variables
	if err := env.Parse(config); err != nil {
		return fmt.Errorf("Error processing environment variables. %s", err)
	}

	// Get the default configuration
	defaultConfig := NewType(config)
	if err := loadConfigFile(configFile, defaultConfig); err != nil {
		return fmt.Errorf("Error processing default configuration. %s", err)
	}
	if err := mergo.Merge(config, defaultConfig); err != nil {
		return fmt.Errorf("Error merging the default configuration. %s", err)
	}

	return nil
}
