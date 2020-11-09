package httpServer

import (
	"encoding/json"
	"fmt"
	"github.com/call-me-snake/service_tg_bot/client/internal/trackerClient"
	"net/http"
)

const (
	serverConnectionErrorMessage = "Ошибка соединения с сервером"
	internalErrorMessage         = "Внутренняя ошибка"
)

func clientDeviceInfo(c *trackerClient.TrackerClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Info("httpServer.clientDeviceInfo: Started")
		clientInfo := clientDeviceInfoResponce{}
		clientInfo.Id = c.ClientId
		clientInfo.Name = c.ClientName
		if c.ClientToken != "" {
			clientInfo.Token = c.ClientToken
		}
		resp, err := json.Marshal(clientInfo)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.clientDeviceInfo: %v", err))
			makeErrResponce(internalErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(resp)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.clientDeviceInfo: %v", err))
		}
	}
}

func registerClientDevice(c *trackerClient.TrackerClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Info("httpServer.registerClientDevice: Started")
		clientErrorMessage, err := c.RegisterNewDevice()
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.registerClientDevice: %v", err))
			makeErrResponce(serverConnectionErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		if clientErrorMessage != "" {
			c.Logger.WithField("clientErrorMessage", clientErrorMessage).Warn("httpServer.registerClientDevice: clientErrorMessage != nil")
			makeErrResponce(clientErrorMessage, http.StatusBadRequest, w, c.Logger)
			return
		}
		respMessage := successResponce{Message: "Клиент успешно зарегистрирован"}
		resp, err := json.Marshal(respMessage)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.registerClientDevice: %v", err))
			makeErrResponce(internalErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(resp)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.registerClientDevice: %v", err))
		}
	}
}

func getToken(c *trackerClient.TrackerClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Info("httpServer.getToken: Started")
		clientErrorMessage, err := c.GetToken()
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.getToken: %v", err))
			makeErrResponce(serverConnectionErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		if clientErrorMessage != "" {
			c.Logger.WithField("clientErrorMessage", clientErrorMessage).Warn("httpServer.getToken: clientErrorMessage != nil")
			makeErrResponce(clientErrorMessage, http.StatusBadRequest, w, c.Logger)
			return
		}
		respMessage := successResponce{Message: "Токен успешно получен"}
		resp, err := json.Marshal(respMessage)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.getToken: %v", err))
			makeErrResponce(internalErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(resp)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.getToken: %v", err))
		}
	}
}

func connect(c *trackerClient.TrackerClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Info("httpServer.connect: Started")
		clientErrorMessage, err := c.SynchDevice()
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.connect: %v", err))
			makeErrResponce(serverConnectionErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		if clientErrorMessage != "" {
			c.Logger.WithField("clientErrorMessage", clientErrorMessage).Warn("httpServer.connect: clientErrorMessage != nil")
			makeErrResponce(clientErrorMessage, http.StatusBadRequest, w, c.Logger)
			return
		}
		respMessage := successResponce{Message: "Устройство подключено"}
		resp, err := json.Marshal(respMessage)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.connect: %v", err))
			makeErrResponce(internalErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(resp)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.connect: %v", err))
		}
	}
}

func disconnect(c *trackerClient.TrackerClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Logger.Info("httpServer.disconnect: Started")
		clientErrorMessage := c.UnSynchDevice()
		if clientErrorMessage != "" {
			c.Logger.WithField("clientErrorMessage", clientErrorMessage).Warn("httpServer.disconnect: clientErrorMessage != nil")
			makeErrResponce(clientErrorMessage, http.StatusBadRequest, w, c.Logger)
			return
		}
		respMessage := successResponce{Message: "Устройство отключено"}
		resp, err := json.Marshal(respMessage)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.disconnect: %v", err))
			makeErrResponce(internalErrorMessage, http.StatusInternalServerError, w, c.Logger)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(resp)
		if err != nil {
			c.Logger.Error(fmt.Errorf("httpServer.disconnect: %v", err))
		}
	}
}
