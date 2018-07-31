package log

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const (
	kCallDepth = 2
)

var (
	logger      = log.New(os.Stderr, "", log.LstdFlags)
	debug       = false
	exitOnFatal = false
)

func Logger() *log.Logger {
	return logger
}

func EnableDebug() {
	debug = true
}

func ExitOnFatal() {
	exitOnFatal = true
}

func DebugV(v interface{}) {
	Debugf("%+v", v)
}

func DebugJson(v interface{}) {
	if debug {
		data, _ := json.MarshalIndent(v, "", "  ")
		Debugf("\n%s", string(data))
	}
}

func Debug(v ...interface{}) {
	if debug {
		logger.Output(kCallDepth, header("DEBUG", fmt.Sprint(v...)))
	}
}

func Debugf(format string, v ...interface{}) {
	if debug {
		logger.Output(kCallDepth, header("DEBUG", fmt.Sprintf(format, v...)))
	}
}

func Info(v ...interface{}) {
	logger.Output(kCallDepth, header("INFO", fmt.Sprint(v...)))
}

func Infof(format string, v ...interface{}) {
	logger.Output(kCallDepth, header("INFO", fmt.Sprintf(format, v...)))
}

func Warn(v ...interface{}) {
	logger.Output(kCallDepth, header("WARN", fmt.Sprint(v...)))
}

func Warnf(format string, v ...interface{}) {
	logger.Output(kCallDepth, header("WARN", fmt.Sprintf(format, v...)))
}

func Error(v ...interface{}) {
	logger.Output(kCallDepth, header("ERROR", fmt.Sprint(v...)))
}

func Errorf(format string, v ...interface{}) {
	logger.Output(kCallDepth, header("ERROR", fmt.Sprintf(format, v...)))
}

func Fatal(v ...interface{}) {
	msg := header("FATAL", fmt.Sprint(v...))
	logger.Output(kCallDepth, msg)
	if exitOnFatal {
		os.Exit(1)
	} else {
		panic(msg)
	}
}

func Fatalf(format string, v ...interface{}) {
	msg := header("FATAL", fmt.Sprintf(format, v...))
	logger.Output(kCallDepth, msg)
	if exitOnFatal {
		os.Exit(1)
	} else {
		panic(msg)
	}
}

func header(level, msg string) string {
	_, file, line, ok := runtime.Caller(kCallDepth)
	if ok {
		file = filepath.Base(file)
	}
	if len(file) == 0 {
		file = "???"
	}
	if line < 0 {
		line = 0
	}

	return fmt.Sprintf("%s %s:%d: %s", level, file, line, msg)
}
