package model

//Config хранит входные параметры программы
type Config struct {
	ClientHttpAddress string
	GrpcServerAddress string
	ClientId          int64
	ClientName        string
	ClientToken       string
	LogMode           string
	DebugMode         bool
}
