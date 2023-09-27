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
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	redString = color.New(color.FgHiRed).SprintfFunc()

	greenString = color.New(color.FgGreen).SprintfFunc()

	yellowString = color.New(color.FgHiYellow).SprintfFunc()

	blueString = color.New(color.FgHiBlue).SprintfFunc()

	errorSymbol        = " x "
	coloredErrorSymbol = color.New(color.BgHiRed, color.FgBlack).Sprint(errorSymbol)

	successSymbol        = " ✓ "
	coloredSuccessSymbol = color.New(color.BgGreen, color.FgBlack).Sprint(successSymbol)

	informationSymbol        = " i "
	coloredInformationSymbol = color.New(color.BgHiBlue, color.FgBlack).Sprint(informationSymbol)

	warningSymbol        = " ! "
	coloredWarningSymbol = color.New(color.BgHiYellow, color.FgBlack).Sprint(warningSymbol)

	questionSymbol        = " ? "
	coloredQuestionSymbol = color.New(color.BgHiMagenta, color.FgBlack).Sprint(questionSymbol)

	// InfoLevel is the json level for information
	InfoLevel = "info"
	// WarningLevel is the json level for warning
	WarningLevel = "warn"
	// ErrorLevel is the json level for error
	ErrorLevel = "error"
	// DebugLevel is the json level for debug
	DebugLevel = "debug"
)

type Logger struct {
	out    *logrus.Logger
	file   *logrus.Entry
	writer OktetoWriter

	outputMode string

	buf *bytes.Buffer

	maskedWords []string
	isMasked    bool
	replacer    *strings.Replacer

	spinner *spinnerLogger
}

// Init configures the logger for the package to use.
func NewLogger(level logrus.Level) *Logger {
	log := &Logger{
		out: logrus.New(),
	}
	log.out.SetOutput(os.Stdout)
	log.out.SetLevel(level)
	log.writer = log.getWriter(TTYFormat)
	log.maskedWords = []string{}
	log.buf = &bytes.Buffer{}
	log.spinner = &spinnerLogger{
		sp:             newSpinner(log.writer),
		spinnerSupport: !log.loadBool(OktetoDisableSpinnerEnvVar) && log.IsInteractive(),
	}

	return log
}

func getRollingLog(path string) io.Writer {
	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    1, // megabytes
		MaxBackups: 10,
		MaxAge:     28, // days
		Compress:   true,
	}
}

// SetLevel sets the level of the main logger
func (log *logger) SetLevel(level string) {
	l, err := logrus.ParseLevel(level)
	if err == nil {
		log.out.SetLevel(l)
	}
}

// GetLevel gets the level of the main logger
func (log *logger) GetLevel() string {
	l := log.out.Level
	return l.String()
}

// GetOutputFormat returns the output format of the command
func (log *logger) GetOutputFormat() string {
	return log.outputMode
}

// GetOutput returns the log output
func (log *logger) GetOutput() io.Writer {
	return log.out.Out
}

// SetOutput sets the log output
func (log *logger) SetOutput(output io.Writer) {
	log.out.SetOutput(output)
}

// SetOutputFormat sets the output format
func (log *logger) SetOutputFormat(format string) {
	log.writer = log.getWriter(format)
}

// GetOutputWriter sets the output format
func (log *logger) GetOutputWriter() OktetoWriter {
	return log.writer
}

// SetStage sets the stage of the logger
func (log *logger) SetStage(stage string) {
	log.writer.SetStage(stage)
}

// IsDebug checks if the level of the main logger is DEBUG or TRACE
func (log *logger) IsDebug() bool {
	return log.out.GetLevel() >= logrus.DebugLevel
}

// Debug writes a debug-level log
func (log *logger) Debug(args ...interface{}) {
	log.writer.Debug(args...)
}

// Debugf writes a debug-level log with a format
func (log *logger) Debugf(format string, args ...interface{}) {
	log.writer.Debugf(format, args...)
}

// Info writes a info-level log
func (log *logger) Info(args ...interface{}) {
	log.writer.Info(args...)
}

// Infof writes a info-level log with a format
func (log *logger) Infof(format string, args ...interface{}) {
	log.writer.Infof(format, args...)
}

// Error writes a error-level log
func (log *logger) Error(args ...interface{}) {
	log.writer.Error(args...)
}

// Errorf writes a error-level log with a format
func (log *logger) Errorf(format string, args ...interface{}) {
	log.writer.Errorf(format, args...)
}

// Fatalf writes a error-level log with a format
func (log *logger) Fatalf(format string, args ...interface{}) {
	log.writer.Fatalf(format, args...)
}

// Yellow writes a line in yellow
func (log *logger) Yellow(format string, args ...interface{}) {
	log.writer.Yellow(format, args...)
}

// Green writes a line in green
func (log *logger) Green(format string, args ...interface{}) {
	log.writer.Green(format, args...)
}

// BlueString returns a string in blue
func BlueString(format string, args ...interface{}) string {
	return blueString(format, args...)
}

// RedString returns a string in blue
func RedString(format string, args ...interface{}) string {
	return redString(format, args...)
}

// BlueBackgroundString returns a string in a blue background
func BlueBackgroundString(format string, args ...interface{}) string {
	return blueString(format, args...)
}

// Success prints a message with the success symbol first, and the text in green
func (log *logger) Success(format string, args ...interface{}) {
	log.writer.Success(format, args...)
}

// Information prints a message with the information symbol first, and the text in blue
func (log *logger) Information(format string, args ...interface{}) {
	log.writer.Information(format, args...)
}

// Question prints a message with the question symbol first, and the text in magenta
func (log *logger) Question(format string, args ...interface{}) error {
	return log.writer.Question(format, args...)
}

// Warning prints a message with the warning symbol first, and the text in yellow
func (log *logger) Warning(format string, args ...interface{}) {
	log.writer.Warning(format, args...)
}

// FWarning prints a message with the warning symbol first, and the text in yellow to a specific writer
func (log *logger) FWarning(w io.Writer, format string, args ...interface{}) {
	log.writer.FWarning(w, format, args...)
}

// Hint prints a message with the text in blue
func (log *logger) Hint(format string, args ...interface{}) {
	log.writer.Hint(format, args...)
}

// Fail prints a message with the error symbol first, and the text in red
func (log *logger) Fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg = log.redactMessage(msg)
	log.writer.Fail(msg)
}

// Println writes a line with colors
func (log *logger) Println(args ...interface{}) {
	msg := fmt.Sprint(args...)
	msg = log.redactMessage(msg)
	log.writer.Println(msg)
}

// FPrintln writes a line with colors to specific writer
func (log *logger) FPrintln(w io.Writer, args ...interface{}) {
	msg := fmt.Sprint(args...)
	msg = log.redactMessage(msg)
	log.writer.FPrintln(w, msg)
}

// Print writes a line with colors
func (log *logger) Print(args ...interface{}) {
	msg := fmt.Sprint(args...)
	msg = log.redactMessage(msg)
	log.writer.Print(msg)
}

// Printf writes a line with format
func (log *logger) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg = log.redactMessage(msg)
	log.writer.Print(msg)
}

// IsInteractive checks if the writer is interactive
func (log *logger) IsInteractive() bool {
	return log.writer.IsInteractive()
}

// AddMaskedWord adds a new word to be redacted
func (log *logger) AddMaskedWord(word string) {
	if strings.TrimSpace(word) != "" {
		log.maskedWords = append(log.maskedWords, word)
	}
}

// EnableMasking starts redacting all variables
func (log *logger) EnableMasking() {
	log.isMasked = true
	sort.Slice(log.maskedWords, func(i, j int) bool {
		return len(log.maskedWords[i]) > len(log.maskedWords[j])
	})
	oldnew := []string{}
	for _, maskWord := range log.maskedWords {
		oldnew = append(oldnew, maskWord)
		oldnew = append(oldnew, "***")
	}
	log.replacer = strings.NewReplacer(oldnew...)
}

// DisableMasking will stop showing secrets and vars
func (log *logger) DisableMasking() {
	log.isMasked = false
}

func (log *logger) redactMessage(message string) string {
	if log.isMasked {
		return log.replacer.Replace(message)
	}
	return message
}

// GetOutputBuffer returns the buffer of the running command
func (log *logger) GetOutputBuffer() *bytes.Buffer {
	return log.buf
}

// AddToBuffer logs into the buffer but does not print anything
func (log *logger) AddToBuffer(level, format string, args ...interface{}) {
	log.writer.AddToBuffer(level, format, args...)
}

func (log *logger) loadBool(env string) bool {
	value := os.Getenv(env)
	if value == "" {
		value = "false"
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Yellow("'%s' is not a valid value for environment variable %s", value, env)
	}
	return boolValue
}
