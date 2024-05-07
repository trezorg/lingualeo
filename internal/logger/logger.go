package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var log *logrus.Entry
var onceLog sync.Once

type utcFormatter struct {
	logrus.Formatter
}

func (u utcFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

func initLogger(logLevel string, logPrettyPrint bool) {

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Fatalf("Cannot parse log level: %s", logLevel)
	}

	logrus.SetFormatter(utcFormatter{&logrus.JSONFormatter{
		PrettyPrint:     logPrettyPrint,
		TimestampFormat: "2006-01-02 15:04:05 -0700",
	}})
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(level)
	log = logrus.WithFields(logrus.Fields{"service": "lingualeo"})
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func InitLogger(logLevel string, logPrettyPrint bool) *logrus.Entry {
	onceLog.Do(func() {
		initLogger(logLevel, logPrettyPrint)
	})
	return log
}
