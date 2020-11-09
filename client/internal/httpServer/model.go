package httpServer

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

type clientDeviceInfoResponce struct {
	Id    int64  `json:"Id"`
	Name  string `json:"Name"`
	Token string `json:"Token,omitempty"`
}

type successResponce struct {
	Message string `json:"Message"`
}

type errorResponce struct {
	Message string `json:"Message"`
	ErrCode int    `json:"ErrCode"`
}

func makeErrResponce(userMessage string, errCode int, w http.ResponseWriter, logger logrus.FieldLogger) {
	res, err := json.Marshal(errorResponce{Message: userMessage, ErrCode: errCode})
	if err != nil {
		logger.Error(fmt.Errorf("httpServer.makeErrResponce: %v", err))
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(errCode)
	_, err = w.Write(res)
	if err != nil {
		logger.Error(fmt.Errorf("httpServer.makeErrResponce: %v", err))
	}
}
