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
	"strings"
	"testing"
	"time"
)

func TestLevelByName(t *testing.T) {
	tests := []struct {
		logLevel string
		expected level
	}{
		{"", defaultLogLevel},
		{"invalid", defaultLogLevel},
		{"Debug", debugLevel},
		{"INFO", infoLevel},
		{"inFo", infoLevel},
		{"wArN", warnLevel},
		{"eRRoR", errorLevel},
		{"fatAL", fatalLevel},
	}
	for _, test := range tests {
		actual := levelByName(test.logLevel)
		if actual != test.expected {
			t.Errorf("Invalid level for %s. Actual: %d. Expected: %d.", test.logLevel, actual, test.expected)
		}
	}
}

func TestLoggerLevel(t *testing.T) {
	logger := NewLogger()
	if logLevel := logger.GetLevel(); logLevel != "INFO" {
		t.Errorf("Invalid logger level. Actual: %s. Expected: %s", logLevel, "INFO")
	}
	logger.SetLevel("DEBUG")
	if logLevel := logger.GetLevel(); logLevel != "DEBUG" {
		t.Errorf("Invalid logger level. Actual: %s. Expected: %s", logLevel, "DEBUG")
	}
	SetDefaultLogLevel("WARN")
	if logLevel := logger.GetLevel(); logLevel != "DEBUG" {
		t.Errorf("Invalid logger level after setting default. Actual: %s. Expected: %s", logLevel, "DEBUG")
	}
	logger = NewLogger()
	if logLevel := logger.GetLevel(); logLevel != "WARN" {
		t.Errorf("Invalid logger level after setting default. Actual: %s. Expected: %s", logLevel, "WARN")
	}
	// Restore the default log level for other tests
	SetDefaultLogLevel("INFO")
}

func TestLoggerContext(t *testing.T) {
	logger := NewLogger()
	if logger.GetLogContext() != nil {
		t.Errorf("Expected nil log context")
	}
	ctxt := LogContext{TransactionID: "txid"}
	logger.SetLogContext(&ctxt)
	if logger.GetLogContext() != &ctxt {
		t.Errorf("Expected valid log context")
	}
}

func TestWriteDoc(t *testing.T) {
	time.Local = time.UTC
	now := time.Now()
	nowBytes, _ := json.Marshal(now)
	nowStr := string(nowBytes)

	ctxtA := LogContext{TransactionID: "txid", Operation: "opA"}
	ctxtB := ReqLogContext{Method: "GET"}

	tests := []struct {
		logLevel string
		ctxtA    interface{}
		ctxtB    interface{}
		message  string
		expected string
	}{
		{"DEBUG", nil, nil, "This is a test", `{"time":` + nowStr + `,"lvl":"DEBUG","msg":"This is a test"}`},
		{"INFO", ctxtA, nil, "This is a test", `{"time":` + nowStr + `,"lvl":"INFO","trans":"txid","op":"opA","msg":"This is a test"}`},
		{"WARN", nil, ctxtB, "This is a test", `{"time":` + nowStr + `,"lvl":"WARN","method":"GET","msg":"This is a test"}`},
		{"ERROR", ctxtA, ctxtB, "This is a test", `{"time":` + nowStr + `,"lvl":"ERROR","trans":"txid","op":"opA","method":"GET","msg":"This is a test"}`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		writeDoc(&buf, now, test.logLevel, test.ctxtA, test.ctxtB, test.message)
		expected := test.expected + "\n"
		if buf.String() != expected {
			t.Errorf("Invalid writeDoc. Actual: %s. Expected: %s", buf.String(), expected)
		}
	}

}

func TestWriteField(t *testing.T) {
	tests := []struct {
		key      string
		value    interface{}
		expected string
	}{
		{"demoNil", nil, `"demoNil":null`},
		{"demoBool", false, `"demoBool":false`},
		{"demoBool", true, `"demoBool":true`},
		{"demoInt", 0, `"demoInt":0`},
		{"demoInt", 102, `"demoInt":102`},
		{"demoStr", "", `"demoStr":""`},
		{"demoStr", "this is a demo", `"demoStr":"this is a demo"`},
		{"demoStr", `this is a "demo"`, `"demoStr":"this is a \"demo\""`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		writeField(&buf, test.key, test.value)
		if buf.String() != test.expected {
			t.Errorf("Invalid writeField. Actual: %s. Expected: %s", buf.String(), test.expected)
		}
	}
}

func TestWriteFields(t *testing.T) {
	tests := []struct {
		fields   interface{}
		expected string
	}{
		{nil, ``},
		{LogContext{}, ``},
		{LogContext{TransactionID: "txid", User: "uid"}, `"trans":"txid","user":"uid"`},
		{ReqLogContext{}, ``},
		{ReqLogContext{Method: "GET"}, `"method":"GET"`},
		{RespLogContext{}, ``},
		{RespLogContext{Latency: 4, Status: 200}, `"status":200,"latency":4`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		length := writeObject(&buf, test.fields)
		if buf.String() != test.expected {
			t.Errorf("Invalid writeObject. Actual: %s. Expected: %s", buf.String(), test.expected)
		}
		if length != len(buf.String()) {
			t.Errorf("Invalid writeObject length. Actual: %d. Expected: %d", length, len(buf.String()))
		}
	}
}

func extractFirstField(r string) string {
	i := strings.Index(r, ",")
	return r[i:]
}

var ctxtA = &LogContext{TransactionID: "txid", Operation: "op1"}
var ctxtB = &ReqLogContext{Method: "GET", Path: "/users"}

func TestDebug(t *testing.T) {
	tests := []struct {
		ctxtA    interface{}
		ctxtB    interface{}
		msg      string
		args     []interface{}
		expected string
	}{
		{nil, nil, "This is a demo", []interface{}{}, `,"lvl":"DEBUG","msg":"This is a demo"}`},
		{nil, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"DEBUG","msg":"This is a test"}`},
		{ctxtA, nil, "This is a demo", []interface{}{}, `,"lvl":"DEBUG","trans":"txid","op":"op1","msg":"This is a demo"}`},
		{ctxtA, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"DEBUG","trans":"txid","op":"op1","msg":"This is a test"}`},
		{ctxtA, ctxtB, "This is a demo", []interface{}{}, `,"lvl":"DEBUG","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a demo"}`},
		{ctxtA, ctxtB, "This is a %s", []interface{}{"test"}, `,"lvl":"DEBUG","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a test"}`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		logger := &Logger{out: &buf, logLevel: debugLevel}
		logger.SetLogContext(test.ctxtA)
		if test.ctxtB == nil {
			logger.Debug(test.msg, test.args...)
		} else {
			logger.DebugC(test.ctxtB, test.msg, test.args...)
		}
		expected := test.expected + "\n"
		if extractFirstField(buf.String()) != expected {
			t.Errorf("Invalid log. Actual: %s. Expected to end with: %s", buf.String(), expected)
		}
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		ctxtA    interface{}
		ctxtB    interface{}
		msg      string
		args     []interface{}
		expected string
	}{
		{nil, nil, "This is a demo", []interface{}{}, `,"lvl":"INFO","msg":"This is a demo"}`},
		{nil, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"INFO","msg":"This is a test"}`},
		{ctxtA, nil, "This is a demo", []interface{}{}, `,"lvl":"INFO","trans":"txid","op":"op1","msg":"This is a demo"}`},
		{ctxtA, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"INFO","trans":"txid","op":"op1","msg":"This is a test"}`},
		{ctxtA, ctxtB, "This is a demo", []interface{}{}, `,"lvl":"INFO","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a demo"}`},
		{ctxtA, ctxtB, "This is a %s", []interface{}{"test"}, `,"lvl":"INFO","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a test"}`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		logger := &Logger{out: &buf, logLevel: debugLevel}
		logger.SetLogContext(test.ctxtA)
		if test.ctxtB == nil {
			logger.Info(test.msg, test.args...)
		} else {
			logger.InfoC(test.ctxtB, test.msg, test.args...)
		}
		expected := test.expected + "\n"
		if extractFirstField(buf.String()) != expected {
			t.Errorf("Invalid log. Actual: %s. Expected to end with: %s", buf.String(), expected)
		}
	}
}

func TestWarn(t *testing.T) {
	tests := []struct {
		ctxtA    interface{}
		ctxtB    interface{}
		msg      string
		args     []interface{}
		expected string
	}{
		{nil, nil, "This is a demo", []interface{}{}, `,"lvl":"WARN","msg":"This is a demo"}`},
		{nil, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"WARN","msg":"This is a test"}`},
		{ctxtA, nil, "This is a demo", []interface{}{}, `,"lvl":"WARN","trans":"txid","op":"op1","msg":"This is a demo"}`},
		{ctxtA, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"WARN","trans":"txid","op":"op1","msg":"This is a test"}`},
		{ctxtA, ctxtB, "This is a demo", []interface{}{}, `,"lvl":"WARN","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a demo"}`},
		{ctxtA, ctxtB, "This is a %s", []interface{}{"test"}, `,"lvl":"WARN","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a test"}`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		logger := &Logger{out: &buf, logLevel: debugLevel}
		logger.SetLogContext(test.ctxtA)
		if test.ctxtB == nil {
			logger.Warn(test.msg, test.args...)
		} else {
			logger.WarnC(test.ctxtB, test.msg, test.args...)
		}
		expected := test.expected + "\n"
		if extractFirstField(buf.String()) != expected {
			t.Errorf("Invalid log. Actual: %s. Expected to end with: %s", buf.String(), expected)
		}
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		ctxtA    interface{}
		ctxtB    interface{}
		msg      string
		args     []interface{}
		expected string
	}{
		{nil, nil, "This is a demo", []interface{}{}, `,"lvl":"ERROR","msg":"This is a demo"}`},
		{nil, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"ERROR","msg":"This is a test"}`},
		{ctxtA, nil, "This is a demo", []interface{}{}, `,"lvl":"ERROR","trans":"txid","op":"op1","msg":"This is a demo"}`},
		{ctxtA, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"ERROR","trans":"txid","op":"op1","msg":"This is a test"}`},
		{ctxtA, ctxtB, "This is a demo", []interface{}{}, `,"lvl":"ERROR","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a demo"}`},
		{ctxtA, ctxtB, "This is a %s", []interface{}{"test"}, `,"lvl":"ERROR","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a test"}`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		logger := &Logger{out: &buf, logLevel: debugLevel}
		logger.SetLogContext(test.ctxtA)
		if test.ctxtB == nil {
			logger.Error(test.msg, test.args...)
		} else {
			logger.ErrorC(test.ctxtB, test.msg, test.args...)
		}
		expected := test.expected + "\n"
		if extractFirstField(buf.String()) != expected {
			t.Errorf("Invalid log. Actual: %s. Expected to end with: %s", buf.String(), expected)
		}
	}
}

func TestFatal(t *testing.T) {
	tests := []struct {
		ctxtA    interface{}
		ctxtB    interface{}
		msg      string
		args     []interface{}
		expected string
	}{
		{nil, nil, "This is a demo", []interface{}{}, `,"lvl":"FATAL","msg":"This is a demo"}`},
		{nil, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"FATAL","msg":"This is a test"}`},
		{ctxtA, nil, "This is a demo", []interface{}{}, `,"lvl":"FATAL","trans":"txid","op":"op1","msg":"This is a demo"}`},
		{ctxtA, nil, "This is a %s", []interface{}{"test"}, `,"lvl":"FATAL","trans":"txid","op":"op1","msg":"This is a test"}`},
		{ctxtA, ctxtB, "This is a demo", []interface{}{}, `,"lvl":"FATAL","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a demo"}`},
		{ctxtA, ctxtB, "This is a %s", []interface{}{"test"}, `,"lvl":"FATAL","trans":"txid","op":"op1","method":"GET","path":"/users","msg":"This is a test"}`},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		logger := &Logger{out: &buf, logLevel: debugLevel}
		logger.SetLogContext(test.ctxtA)
		if test.ctxtB == nil {
			logger.Fatal(test.msg, test.args...)
		} else {
			logger.FatalC(test.ctxtB, test.msg, test.args...)
		}
		expected := test.expected + "\n"
		if extractFirstField(buf.String()) != expected {
			t.Errorf("Invalid log. Actual: %s. Expected to end with: %s", buf.String(), expected)
		}
	}
}

func TestStdLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{out: &buf, logLevel: infoLevel}
	l := NewStdLogger(logger)
	l.Printf("This is a demo")
	expected := `,"lvl":"ERROR","msg":"This is a demo"}` + "\n"
	if extractFirstField(buf.String()) != expected {
		t.Errorf("Invalid std log. Actual: %s. Expected to end with: %s", buf.String(), expected)
	}
}

// func TestStdCLogger(t *testing.T) {
// 	var buf bytes.Buffer
// 	l := NewStdLoggerC(ctxtA)
// 	l.Printf("This is a demo")
// 	expected := `,"lvl":"ERROR","msg":"This is a demo"}` + "\n"
// 	if extractFirstField(buf.String()) != expected {
// 		t.Errorf("Invalid std log. Actual: %s. Expected to end with: %s", buf.String(), expected)
// 	}
// }
