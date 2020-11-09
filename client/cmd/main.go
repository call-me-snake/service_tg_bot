package main

import (
	"fmt"
	"github.com/call-me-snake/service_tg_bot/client/internal/httpServer"
	"github.com/call-me-snake/service_tg_bot/client/internal/model"
	"github.com/call-me-snake/service_tg_bot/client/internal/trackerClient"
	logger2 "github.com/call-me-snake/service_tg_bot/internal/logger"
	"github.com/jessevdk/go-flags"
	"log"
)

//envs - получает флаги командной строки или переменные окружения
type commandFlags struct {
	ClientHttpAddress string `long:"http" env:"HTTP" description:"address of client" default:":8000"`
	GrpcServerAddress string `long:"grpc" env:"GRPC" description:"grpc server address" default:"localhost:50051"`
	ClientId          int64  `long:"id" env:"ID" description:"client id" default:"0"`
	ClientName        string `long:"name" env:"NAME" description:"client name" default:""`
	ClientToken       string `long:"token" env:"TOKEN" description:"client token" default:""`
	LogMode           string `long:"logmode" env:"LOGMODE" description:"log mode. can be 'console' or 'file'" default:"console"`
	DebugMode         bool   `long:"debugmode" env:"DEBUGMODE" description:"debug mode trigger. In debug mode we log info messages"`
}

func initConfig() (model.Config, error) {
	f := commandFlags{}
	c := model.Config{}
	var err error
	parser := flags.NewParser(&f, flags.Default)
	if _, err = parser.Parse(); err != nil {
		return c, fmt.Errorf("Init: %v", err)
	}
	c.ClientHttpAddress = f.ClientHttpAddress
	c.GrpcServerAddress = f.GrpcServerAddress
	var errString string
	if f.ClientId == 0 {
		errString += "Не задан id"
	}
	c.ClientId = f.ClientId
	if f.ClientName == "" {
		errString += "; Не задано name"
	}
	c.ClientName = f.ClientName
	c.ClientToken = f.ClientToken
	c.LogMode = f.LogMode
	c.DebugMode = f.DebugMode
	if errString != "" {
		return c, fmt.Errorf("Init: %s", errString)
	} else {
		return c, nil
	}
}

/*
func initLogger(logMode string, debugMode bool) logrus.FieldLogger {
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
		log.Printf("initLogger: неизвестный режим логирования: %s, вывод в консоль", config.LogMode)
		logger.SetOutput(os.Stdout)
	}

	if !debugMode {
		logger.SetLevel(logrus.ErrorLevel)
	}
	return logger
}
*/
func main() {
	log.Print("Started")
	//получение входных параметров
	config, err := initConfig()
	if err != nil {
		log.Print(err.Error())
		return
	}
	//инициализация логгера
	logger := logger2.InitLogger(config.LogMode, config.DebugMode)
	logger.WithField("config", config).Info("service_tg_bot: tracker_client started")

	//инициализация grpc клиента
	client, closeFunc, err := trackerClient.StartNewClient(logger, config.GrpcServerAddress, config.ClientId, config.ClientName, config.ClientToken)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer closeFunc(client)

	//инициализация http сервера
	s := httpServer.New(config.ClientHttpAddress)
	err = s.Start(client)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
