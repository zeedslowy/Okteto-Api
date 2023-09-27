// Copyright 2023 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logexperimental

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
)

// PlainWriter writes into a plain terminal
type PlainWriter struct {
	out   *logrus.Logger
	file  *logrus.Entry
	stage string
	buf   *bytes.Buffer
}

// newPlainWriter creates a new plainWriter
func newPlainWriter(out *logrus.Logger, file *logrus.Entry) *PlainWriter {
	return &PlainWriter{
		out:  out,
		file: file,
	}
}

func (w *PlainWriter) SetStage(stage string) {
	w.stage = stage
}

// Debug writes a debug-level log
func (w *PlainWriter) Debug(args ...interface{}) {
	w.out.Debug(args...)
	if w.file != nil {
		w.file.Debug(args...)
	}
}

// Debugf writes a debug-level log with a format
func (w *PlainWriter) Debugf(format string, args ...interface{}) {
	w.out.Debugf(format, args...)
	if w.file != nil {
		w.file.Debugf(format, args...)
	}
}

// Info writes a info-level log
func (w *PlainWriter) Info(args ...interface{}) {
	w.out.Info(args...)
	if w.file != nil {
		w.file.Info(args...)
	}
}

// Infof writes a info-level log with a format
func (w *PlainWriter) Infof(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	if w.file != nil {
		w.file.Infof(format, args...)
	}
}

// Error writes a error-level log
func (w *PlainWriter) Error(args ...interface{}) {
	w.out.Error(args...)
	if w.file != nil {
		w.file.Error(args...)
	}
}

// Errorf writes a error-level log with a format
func (w *PlainWriter) Errorf(format string, args ...interface{}) {
	w.out.Errorf(format, args...)
	if w.file != nil {
		w.file.Errorf(format, args...)
	}
}

// Fatalf writes a error-level log with a format
func (w *PlainWriter) Fatalf(format string, args ...interface{}) {
	if w.file != nil {
		w.file.Errorf(format, args...)
	}

	w.out.Fatalf(format, args...)
}

// Green writes a line in green
func (w *PlainWriter) Green(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.FPrintln(w.out.Out, fmt.Sprintf(format, args...))
}

// Yellow writes a line in yellow
func (w *PlainWriter) Yellow(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.FPrintln(w.out.Out, fmt.Sprintf(format, args...))
}

// Success prints a message with the success symbol first, and the text in green
func (w *PlainWriter) Success(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.Fprintf(w.out.Out, "SUCCESS: %s\n", fmt.Sprintf(format, args...))
}

// Information prints a message with the information symbol first, and the text in blue
func (w *PlainWriter) Information(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.Fprintf(w.out.Out, "INFO: %s\n", fmt.Sprintf(format, args...))
}

// Question prints a message with the question symbol first, and the text in magenta
func (w *PlainWriter) Question(format string, args ...interface{}) error {
	w.out.Infof(format, args...)
	w.Fprintf(w.out.Out, "%s %s", questionSymbol, fmt.Sprintf(format, args...))
	return nil
}

// Warning prints a message with the warning symbol first, and the text in yellow
func (w *PlainWriter) Warning(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.Fprintf(w.out.Out, "WARNING: %s\n", fmt.Sprintf(format, args...))
}

// FWarning prints a message with the warning symbol first, and the text in yellow into an specific writer
func (w *PlainWriter) FWarning(writer io.Writer, format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.Fprintf(writer, "WARNING: %s\n", fmt.Sprintf(format, args...))
}

// Hint prints a message with the text in blue
func (w *PlainWriter) Hint(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.Fprintf(w.out.Out, "%s\n", fmt.Sprintf(format, args...))
}

// Fail prints a message with the error symbol first, and the text in red
func (w *PlainWriter) Fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	w.out.Info(msg)
	w.Fprintf(w.out.Out, "ERROR: %s\n", fmt.Sprintf(format, args...))
	if msg != "" {
		msg = w.convertToJSON(ErrorLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}
}

// Println writes a line with colors
func (w *PlainWriter) Println(args ...interface{}) {
	msg := fmt.Sprint(args...)
	w.out.Info(msg)
	w.FPrintln(w.out.Out, args...)
	if msg != "" {
		msg = w.convertToJSON(InfoLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}
}

// Fprintf prints a line with format
func (w *PlainWriter) Fprintf(writer io.Writer, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprint(writer, msg)
	if msg != "" && writer == w.out.Out {
		msg = w.convertToJSON(InfoLevel, w.stage, msg)
		w.buf.WriteString(msg)
		w.buf.WriteString("\n")
	}
}

// FPrintln prints a line with format
func (w *PlainWriter) FPrintln(writer io.Writer, args ...interface{}) {
	msg := fmt.Sprint(args...)
	fmt.Fprintln(writer, args...)
	if msg != "" && writer == w.out.Out {
		msg = w.convertToJSON(InfoLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}
}

// Print writes a line with colors
func (w *PlainWriter) Print(args ...interface{}) {
	msg := fmt.Sprint(args...)
	fmt.Fprint(w.out.Out, args...)
	if msg != "" {
		msg = w.convertToJSON(InfoLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}
}

// Printf writes a line with format
func (w *PlainWriter) Printf(format string, a ...interface{}) {
	w.Fprintf(w.out.Out, format, a...)
}

// IsInteractive checks if the writer is interactive
func (*PlainWriter) IsInteractive() bool {
	return false
}

// AddToBuffer logs into the buffer but does not print anything
func (w *PlainWriter) AddToBuffer(level, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if msg != "" {
		msg = w.convertToJSON(level, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}
}

// Write logs into the buffer but does not print anything
func (w *PlainWriter) Write(p []byte) (n int, err error) {
	return w.out.Out.Write(p)
}

func (w *PlainWriter) convertToJSON(level, stage, message string) string {
	message = strings.TrimRightFunc(message, unicode.IsSpace)
	if stage == "" || message == "" {
		return ""
	}
	messageStruct := jsonMessage{
		Level:     level,
		Message:   ansiRegex.ReplaceAllString(message, ""),
		Stage:     stage,
		Timestamp: time.Now().Unix(),
	}
	messageJSON, err := json.Marshal(messageStruct)
	if err != nil {
		w.Infof("error marshalling message: %s", err)
		return ""
	}
	return string(messageJSON)
}
