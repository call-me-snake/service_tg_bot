package model

import "time"

const DevicesIdConstraint = "id_pk"

//IClientDeviceStorage - интерфейс для работы с устройствами
type IClientDeviceStorage interface {
	//RegisterNewDevice - регистрация нового устройства
	RegisterNewDevice(info *DeviceInfo) (deviceToken string, err error)
	//GetDeviceInfo - получение информации по устройству
	GetDeviceInfo(deviceId int64) (*DeviceInfo, error)
	//GetDeviceInfoList - получение информации по всем устройствам
	GetDeviceInfoList() ([]DeviceInfo, error)
	//SetOnlineStatus - установить онлайн статус устройства
	SetOnlineStatus(deviceId int64, isOnline bool) (isSet bool, err error)
	//Сброс всех статусов до оффлайн (выполняется в начале работы приложения. исправление непредвиденных ошибок, например отключение бд или падение приложения)
	ResetStatuses()(err error)
}

//DeviceInfo - структура для хранения информации по устройству
type DeviceInfo struct {
	Id        int64     `gorm:"primary_key;column:id"`
	Name      string    `gorm:"column:name"`
	Token     string    `gorm:"column:token"`
	CreatedAt time.Time `gorm:"column:created_at"`
	Online    bool      `gorm:"column:online"`
}

// TableName - declare table name for GORM
func (DeviceInfo) TableName() string {
	return "devices"
}

type Config struct {
	GrpcServerPort string
	StorageAdress  string
	TgBotToken     string
	LogMode        string
	DebugMode      bool
}
