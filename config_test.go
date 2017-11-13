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

import "testing"
import "os"

type config struct {
	Address  string `json:"address" env:"ADDRESS"`
	BasePath string `json:"basePath" env:"BASE_PATH"`
	LogLevel string `json:"logLevel" env:"LOG_LEVEL"`
	Realm    string `json:"realm" env:"REALM"`
}

func TestGetConfig(t *testing.T) {
	expected := config{Address: ":80", BasePath: "/users", LogLevel: "INFO", Realm: "es"}
	var actual config
	if err := GetConfig("testdata/config.json", &actual); err != nil {
		t.Errorf("Error getting config. %s", err)
	}
	if actual != expected {
		t.Errorf("Error getting config. Actual: %+v. Expected: %+v.", actual, expected)
	}
}

func TestGetConfigWithEnv(t *testing.T) {
	os.Setenv("ADDRESS", ":8080")
	os.Setenv("LOG_LEVEL", "DEBUG")
	expected := config{Address: ":8080", BasePath: "/users", LogLevel: "DEBUG", Realm: "es"}
	var actual config
	if err := GetConfig("testdata/config.json", &actual); err != nil {
		t.Errorf("Error getting config. %s", err)
	}
	if actual != expected {
		t.Errorf("Error getting config. Actual: %+v. Expected: %+v.", actual, expected)
	}
}

func TestGetConfigWrongFile(t *testing.T) {
	var actual config
	err := GetConfig("testdata/configNotExistent.json", &actual)
	if err.Error() != "Error processing default configuration. open testdata/configNotExistent.json: no such file or directory" {
		t.Errorf("Invalid error getting configuration. %s", err)
	}
}
