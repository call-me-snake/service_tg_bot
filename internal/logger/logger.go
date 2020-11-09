package logger

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

const (
	logFileName        = "./logfile.txt"
	logTimeStampFormat = "02-01-2006 15:04:05"
)

func InitLogger(logMode string, debugMode bool) logrus.FieldLogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: logTimeStampFormat,
		FullTimestamp:   true,
	})
	switch logMode {
	case "file":
		file, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Printf("initLogger: ошибка открытия файла для логирования: %v, вывод в консоль", err)
			logger.SetOutput(os.Stdout)
		} else {
			logger.SetOutput(file)
		}
	case "console":
		logger.SetOutput(os.Stdout)
	default:
		log.Printf("initLogger: неизвестный режим логирования: %s, вывод в консоль", logMode)
		logger.SetOutput(os.Stdout)
	}

	if !debugMode {
		logger.SetLevel(logrus.ErrorLevel)
	}
	return logger
}
