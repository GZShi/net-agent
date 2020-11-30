package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

var gLogger = logrus.New()

func init() {
	gLogger.Formatter = &errFormatter{&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02T15:04:05.9999",
	}}

	gLogger.Out = ansicolor.NewAnsiColorWriter(os.Stdout)
	gLogger.Level = logrus.DebugLevel
}

// Get 获取实例
func Get() *logrus.Logger {
	return gLogger
}

type errFormatter struct {
	Formatter *logrus.TextFormatter
}

// Format 将error字段进行格式化
func (f *errFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 对错误信息进行简化
	val, ok := entry.Data[logrus.ErrorKey]
	if ok {
		str := fmt.Sprint(val)
		const errEstablishFailed = "A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond."
		const errUseClosedConn = ": use of closed network connection"
		switch {
		case strings.HasSuffix(str, errEstablishFailed):
			entry.Data[logrus.ErrorKey] = "errEstablishFailed"
		case strings.HasSuffix(str, errUseClosedConn):
			entry.Data[logrus.ErrorKey] = "errUseClosedConn"
		}
	}

	return f.Formatter.Format(entry)
}
