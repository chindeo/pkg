package logging

import (
	"fmt"
	"path/filepath"

	"github.com/chindeo/pkg/file"
)

func GetMyLogger(name string) *Logger {
	var logger *Logger
	logger = NewLogger(&Options{
		Rolling:     DAILY,
		TimesFormat: TIMESECOND,
	}, filepath.Join(file.SelfDir(), fmt.Sprintf("./logs/%s.log", name)))
	logger.SetLogPrefix("log_prefix")
	return logger
}
