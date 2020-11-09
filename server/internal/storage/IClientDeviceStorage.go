package storage

import (
	"fmt"
	"github.com/call-me-snake/service_tg_bot/server/internal/helpers"
	"github.com/call-me-snake/service_tg_bot/server/internal/model"
	"github.com/jinzhu/gorm"
	"time"
)

//методы, реализующие интерфейс model.IClientDeviceStorage

//RegisterNewDevice - реализует метод интерфейса model.IClientDeviceStorage
func (db *storage) RegisterNewDevice(info *model.DeviceInfo) (deviceToken string, err error) {
	info.CreatedAt = time.Now()
	info.Token = helpers.GenerateToken()
	result := db.database.Create(info)
	if result.Error != nil {
		return "", fmt.Errorf("storage.RegisterNewDevice: %v", result.Error)
	}
	return info.Token, nil
}

//GetDeviceInfo - реализует метод интерфейса model.IClientDeviceStorage
func (db *storage) GetDeviceInfo(deviceId int64) (*model.DeviceInfo, error) {
	result := &model.DeviceInfo{}
	query := db.database.First(result, deviceId)
	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("storage.GetDeviceInfo: %v", query.Error)
	}
	return result, nil
}

//GetDeviceInfoList - реализует метод интерфейса model.IClientDeviceStorage
func (db *storage) GetDeviceInfoList() ([]model.DeviceInfo, error) {
	result := make([]model.DeviceInfo, 0)
	query := db.database.Order("id ASC").Find(&result)
	if query.Error != nil {
		return nil, fmt.Errorf("storage.GetDeviceInfoList: %v", query.Error)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

//SetOnlineStatus - реализует метод интерфейса model.IClientDeviceStorage
func (db *storage) SetOnlineStatus(deviceId int64, isOnline bool) (isSet bool, err error) {
	update := db.database.Model(&model.DeviceInfo{}).Where("id = ?", deviceId).Update("online", isOnline)
	if update.Error != nil {
		return false, fmt.Errorf("storage.SetOnlineStatus: %v", err)
	}
	if update.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

//ResetStatuses - реализует метод интерфейса model.IClientDeviceStorage
func (db *storage) ResetStatuses() (err error) {
	update:=db.database.Model(&model.DeviceInfo{}).Update("online", false)
	if update.Error != nil {
		return fmt.Errorf("storage.ResetStatuses: %v", err)
	}
	return nil
}
