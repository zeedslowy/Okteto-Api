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
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var ansiRegex = regexp.MustCompile(ansi)

// TTYWriter writes into a tty terminal
type TTYWriter struct {
	out     *logrus.Logger
	file    *logrus.Entry
	stage   string
	buf     *bytes.Buffer
	spinner *spinnerLogger
}

// newTTYWriter creates a new ttyWriter
func newTTYWriter(out *logrus.Logger, file *logrus.Entry, spinner *spinnerLogger) *TTYWriter {
	return &TTYWriter{
		out:     out,
		file:    file,
		spinner: spinner,
	}
}

func (w *TTYWriter) SetStage(stage string) {
	w.stage = stage
}

// Debug writes a debug-level log
func (w *TTYWriter) Debug(args ...interface{}) {
	w.out.Debug(args...)
	if w.file != nil {
		w.file.Debug(args...)
	}
}

// Debugf writes a debug-level log with a format
func (w *TTYWriter) Debugf(format string, args ...interface{}) {
	w.out.Debugf(format, args...)
	if w.file != nil {
		w.file.Debugf(format, args...)
	}
}

// Info writes a info-level log
func (w *TTYWriter) Info(args ...interface{}) {
	w.out.Info(args...)
	if w.file != nil {
		w.file.Info(args...)
	}
}

// Infof writes a info-level log with a format
func (w *TTYWriter) Infof(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	if w.file != nil {
		w.file.Infof(format, args...)
	}
}

// Error writes a error-level log
func (w *TTYWriter) Error(args ...interface{}) {
	w.out.Error(args...)
	if w.file != nil {
		w.file.Error(args...)
	}
}

// Errorf writes a error-level log with a format
func (w *TTYWriter) Errorf(format string, args ...interface{}) {
	w.out.Errorf(format, args...)
	if w.file != nil {
		w.file.Errorf(format, args...)
	}
}

// Fatalf writes a error-level log with a format
func (w *TTYWriter) Fatalf(format string, args ...interface{}) {
	if w.file != nil {
		w.file.Errorf(format, args...)
	}

	w.out.Fatalf(format, args...)
}

// Green writes a line in green
func (w *TTYWriter) Green(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.FPrintln(w.out.Out, greenString(format, args...))
	w.spinner.unhold()
}

// Yellow writes a line in yellow
func (w *TTYWriter) Yellow(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.FPrintln(w.out.Out, yellowString(format, args...))
	w.spinner.unhold()
}

// Success prints a message with the success symbol first, and the text in green
func (w *TTYWriter) Success(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.Fprintf(w.out.Out, "%s %s\n", coloredSuccessSymbol, greenString(format, args...))
	w.spinner.unhold()
}

// Information prints a message with the information symbol first, and the text in blue
func (w *TTYWriter) Information(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.Fprintf(w.out.Out, "%s %s\n", coloredInformationSymbol, blueString(format, args...))
	w.spinner.unhold()
}

// Question prints a message with the question symbol first, and the text in magenta
func (w *TTYWriter) Question(format string, args ...interface{}) error {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.Fprintf(w.out.Out, "%s %s", coloredQuestionSymbol, color.MagentaString(format, args...))
	w.spinner.unhold()
	return nil
}

// Warning prints a message with the warning symbol first, and the text in yellow
func (w *TTYWriter) Warning(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.Fprintf(w.out.Out, "%s %s\n", coloredWarningSymbol, yellowString(format, args...))
	w.spinner.unhold()
}

// FWarning prints a message with the warning symbol first, and the text in yellow into an specific writer
func (w *TTYWriter) FWarning(writer io.Writer, format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.Fprintf(writer, "%s %s\n", coloredWarningSymbol, yellowString(format, args...))
	w.spinner.unhold()
}

// Hint prints a message with the text in blue
func (w *TTYWriter) Hint(format string, args ...interface{}) {
	w.out.Infof(format, args...)
	w.spinner.hold()
	w.Fprintf(w.out.Out, "%s\n", blueString(format, args...))
	w.spinner.unhold()
}

// Fail prints a message with the error symbol first, and the text in red
func (w *TTYWriter) Fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	w.out.Info(msg)
	w.spinner.hold()
	w.Fprintf(w.out.Out, "%s %s\n", coloredErrorSymbol, redString(format, args...))
	w.spinner.unhold()
	if msg != "" {
		msg = w.convertToJSON(ErrorLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}
}

// Println writes a line with colors
func (w *TTYWriter) Println(args ...interface{}) {
	w.out.Info(args...)
	w.spinner.hold()
	w.FPrintln(w.out.Out, args...)
	w.spinner.unhold()
}

// Fprintf prints a line with format
func (w *TTYWriter) Fprintf(writer io.Writer, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprint(writer, msg)
	if msg != "" && writer == w.out.Out {
		msg = w.convertToJSON(InfoLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}

}

// FPrintln prints a line with format
func (w *TTYWriter) FPrintln(writer io.Writer, args ...interface{}) {
	msg := fmt.Sprint(args...)
	fmt.Fprintln(writer, msg)
	if msg != "" && writer == w.out.Out {
		msg = w.convertToJSON(InfoLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}

}

// Print writes a line with colors
func (w *TTYWriter) Print(args ...interface{}) {
	msg := fmt.Sprint(args...)
	fmt.Fprint(w.out.Out, args...)
	if msg != "" {
		msg = w.convertToJSON(ErrorLevel, w.stage, msg)
		if msg != "" {
			w.buf.WriteString(msg)
			w.buf.WriteString("\n")
		}
	}

}

// Printf writes a line with format
func (w *TTYWriter) Printf(format string, a ...interface{}) {
	w.spinner.hold()
	w.Fprintf(w.out.Out, format, a...)
	w.spinner.unhold()
}

// IsInteractive checks if the writer is interactive
func (*TTYWriter) IsInteractive() bool {
	return true
}

// AddToBuffer logs into the buffer but does not print anything
func (w *TTYWriter) AddToBuffer(level, format string, a ...interface{}) {
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
func (w *TTYWriter) Write(p []byte) (n int, err error) {
	return w.out.Out.Write(p)
}

func (w *TTYWriter) convertToJSON(level, stage, message string) string {
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
