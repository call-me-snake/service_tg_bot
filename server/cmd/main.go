package main

import (
	"errors"
	"fmt"
	logger2 "github.com/call-me-snake/service_tg_bot/internal/logger"
	"github.com/call-me-snake/service_tg_bot/server/internal/model"
	"github.com/call-me-snake/service_tg_bot/server/internal/storage"
	"github.com/call-me-snake/service_tg_bot/server/internal/telegram"
	"github.com/call-me-snake/service_tg_bot/server/internal/trackerServer"
	"github.com/jessevdk/go-flags"
	"log"
)

type commandFlags struct {
	GrpcServerPort string `long:"grpc" env:"GRPC" description:"grpc server port" default:":50051"`
	StorageAdress  string `long:"storage" env:"STORAGE" description:"storage address" default:"user=postgres password=example dbname=devices sslmode=disable port=5432 host=localhost"`
	TgBotToken     string `long:"bot-token" env:"BOT_TOKEN" description:"telegram bot token" default:""`
	LogMode        string `long:"logmode" env:"LOGMODE" description:"log mode. can be 'console' or 'file'" default:"console"`
	DebugMode      bool   `long:"debugmode" env:"DEBUGMODE" description:"debug mode trigger. In debug mode we log info messages"`
}

func initConfig() (model.Config, error) {
	f := commandFlags{}
	c := model.Config{}
	var err error
	parser := flags.NewParser(&f, flags.Default)
	if _, err = parser.Parse(); err != nil {
		return c, fmt.Errorf("Init: %v", err)
	}
	c.GrpcServerPort = f.GrpcServerPort
	c.StorageAdress = f.StorageAdress
	if f.TgBotToken == "" {
		return c, errors.New("Init: не задан TgBotToken")
	}
	c.TgBotToken = f.TgBotToken
	c.LogMode = f.LogMode
	c.DebugMode = f.DebugMode
	return c, nil
}

func main() {

	fmt.Println("Started")
	//получение входных параметров
	config, err := initConfig()
	if err != nil {
		log.Print(err.Error())
		return
	}
	//инициализация логгера
	logger := logger2.InitLogger(config.LogMode, config.DebugMode)
	logger.WithField("config", config).Info("service_tg_bot: tracker_server started")

	//подключение к базе устройств
	s, err := storage.New(config.StorageAdress)
	if err != nil {
		logger.Fatal(err.Error())
	}

	//установка статусов всех устройств в оффлайн
	err = s.ResetStatuses()
	if err!=nil{
		logger.Fatal(err.Error())
	}

	//запуск grpc сервера
	go func() {
		err = trackerServer.StartServer(config.GrpcServerPort, s)
		if err != nil {
			logger.Fatal(err.Error())
		}
	}()

	//подключение к боту
	err = telegram.StartBot(config.TgBotToken, s, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
