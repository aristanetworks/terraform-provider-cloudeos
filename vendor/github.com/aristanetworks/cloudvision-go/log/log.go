// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

// Package log is a generic logging system that does multiplexing of logs according to the caller
// interface of the log functions. InitLogging() should be called to set up multiplexing for
// an interface and Log(intf interface{}) should be used for all subsequent loggings.
// A logging hook is used that gets triggered on every log function calls, and this is where
// multiplexing is done. The hook keeps a map from raw pointers of the interfaces passed in
// by InitLogging() to its output io.Writer. A call to Log(intf interface{}) sets the caller
// key, which is then retrieved in the hook to determine which io.Writer to use for the log entry.
package log

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"
)

type callerKeyType string

const (
	// callerKey is the key used for the logger to retrive the struct instace from the context
	callerKey callerKeyType = "caller"
)

var (
	globalLogger = logger{}
)

func init() {
	logrus.AddHook(&globalLogger)
}

// logger implements the logrus.Hook interface, where Levels() returns what logging levels
// should this hook be triggered, and Fire() is called on every call global logging function
// like logrus.Info() just before the logs are written to the logger's output io.Writer.
type logger struct {
	logDir string

	// map from uintptr of struct instances to io.Writer
	logDestMap sync.Map
}

func (l *logger) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *logger) Fire(entry *logrus.Entry) error {
	ctx := entry.Context
	if ctx == nil {
		return nil
	}
	caller := ctx.Value(callerKey)
	if !isPointer(caller) {
		return nil
	}
	key := mapKey(caller)
	out, ok := l.logDestMap.Load(key)
	if !ok {
		// It is possible that log is called before InitLogging is called to an interface.
		// This only happens in the first call to device.DeviceID() function, because
		// to multiplex a log it in turn needs its device ID, so it's a chicken-or-egg problem.
		// As this only happens once per device we don't return an error.
		return nil
	}
	// inject information about the caller in the logs
	entry.Data[string(callerKey)] = reflect.TypeOf(caller)
	entry.Logger.Out = out.(io.Writer)
	return nil
}

func logContext(intf interface{}) context.Context {
	return context.WithValue(context.Background(), callerKey, intf)
}

// The map can't use the interface directly as a key because it does a deep equal, and if
// some field inside the interface changes, we won't be able to retrieve the original key. For
// that reason, a pointer to the interface is used because in this case we want a shallow equal.
func mapKey(intf interface{}) interface{} {
	return reflect.ValueOf(intf).Pointer()
}

// mapKey() will panic if its input interface{} is not a pointer.
// (Passing in a non-pointer doesn't make sense because its address will change.)
// isPointer() filters out anything that's not a pointer.
func isPointer(intf interface{}) bool {
	return intf != nil && reflect.ValueOf(intf).Kind() == reflect.Ptr
}

// InitLogging sets up logging for an input interface.
func InitLogging(filename string, intf interface{}) error {
	if !isPointer(intf) {
		logrus.Errorf("Cannot init logging for interface %#v because it's not a pointer type", intf)
		return nil
	}
	var out io.Writer
	if globalLogger.logDir == "" {
		out = os.Stderr
	} else {
		f, err := os.OpenFile(filepath.Join(globalLogger.logDir, filename),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		out = f
	}
	key := mapKey(intf)
	globalLogger.logDestMap.Store(key, out)
	return nil
}

// Log is a wrapper for logging information related to an interface.
func Log(intf interface{}) *logrus.Entry {
	return logrus.WithContext(logContext(intf))
}

// SetLogDir sets the output of the log files to a directory.
func SetLogDir(logDir string) {
	globalLogger.logDir = logDir
}
