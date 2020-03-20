package log

import (
	"os"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

var gLogger = logrus.New()

func init() {
	gLogger.Formatter = &logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02T15:04:05.9999",
	}

	gLogger.Out = ansicolor.NewAnsiColorWriter(os.Stdout)
	gLogger.Level = logrus.DebugLevel
}

// Get 获取实例
func Get() *logrus.Logger {
	return gLogger
}
