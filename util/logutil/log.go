package logutil

import (
	"github.com/sirupsen/logrus"
	"log"
)

func Init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceQuote:      true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	logrus.SetReportCaller(true)
}

func SetLogLevel(level uint32) {
	logrus.SetLevel(logrus.Level(level))
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func LogOnError(err error, msg string) {
	if err != nil {
		logrus.Errorf("msg= %s, err= %s", msg, err)
	}
}
