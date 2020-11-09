package trackerClient

import (
	"context"
	"fmt"
	"github.com/call-me-snake/service_tg_bot/internal/tracker"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

const (
	heartbeatIntervalInSec = 5
	connectionErrorMessage = "Ошибка соединения"
)

//TrackerClient - структура grpc клиента
type TrackerClient struct {
	Logger               logrus.FieldLogger
	serverAdress         string
	conn                 *grpc.ClientConn
	client               tracker.TrackerClient
	ctx                  context.Context
	ClientId             int64
	ClientName           string
	ClientToken          string
	isConnected          bool
	LastConnectionStatus string
	LastConnectionError  error
}

//StartNewClient - конструктор grpc клиента
func StartNewClient(logger logrus.FieldLogger, serverAddress string, clientId int64, clientName, clientToken string) (c *TrackerClient, connCloseFunc func(c *TrackerClient), err error) {
	c = &TrackerClient{
		Logger:       logger,
		serverAdress: serverAddress,
		ClientId:     clientId,
		ClientName:   clientName,
		ClientToken:  clientToken}
	c.conn, err = grpc.Dial(c.serverAdress, grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("trackerClient.StartNewClient: %v", err)
	}
	c.client = tracker.NewTrackerClient(c.conn)
	c.ctx = context.Background()
	return c,
		func(c *TrackerClient) {
			err = c.conn.Close()
			if err != nil {
				c.Logger.Error(fmt.Errorf("trackerClient.connCloseFunc: %v", err.Error()))
			}
		},
		nil
}

//RegisterNewDevice - регистрация устройства
func (c *TrackerClient) RegisterNewDevice() (clientErrorMessage string, connectionError error) {
	resp, err := c.client.Register(c.ctx, &tracker.TokenRequest{Name: c.ClientName, Id: c.ClientId})
	if err != nil {
		return "", fmt.Errorf("trackerClient.RegisterNewDevice: %v", err)
	}
	if resp.Token != "" {
		c.ClientToken = resp.Token
	}
	return resp.ErrorMessage, nil
}

//GetToken - получение токена
func (c *TrackerClient) GetToken() (clientErrorMessage string, connectionError error) {
	resp, err := c.client.GetToken(c.ctx, &tracker.TokenRequest{Name: c.ClientName, Id: c.ClientId})
	if err != nil {
		return "", fmt.Errorf("trackerClient.RegisterNewDevice: %v", err)
	}
	if resp.Token != "" {
		c.ClientToken = resp.Token
	}
	return resp.ErrorMessage, nil
}

//SynchDevice - синхронизация клиента с сервером (клиент онлайн)
func (c *TrackerClient) SynchDevice() (clientErrorMessage string, connectionError error) {
	if c.isConnected {
		return "Device already connected", nil
	}
	stream, err := c.client.Synch(c.ctx)
	if err != nil {
		return connectionErrorMessage, err
	}
	//Отправка первого сообщения на сервер
	if err = stream.Send(&tracker.Heartbeat{Id: c.ClientId, Token: c.ClientToken}); err != nil {
		c.LastConnectionError = err
		return connectionErrorMessage, err
	}

	//получение первого сообщения от сервера
	serverAnswer, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			//дописать
			return "", err
		}
		return "", err
	}
	if serverAnswer.Synched == false {
		err = stream.CloseSend()
		if err != nil {
			log.Print(err)
		}
		return serverAnswer.ErrorMessage, nil
	}
	go func() {
		c.isConnected = true
		defer func() {
			err = stream.CloseSend()
			if err != nil {
				log.Print(err)
			}
			c.isConnected = false
		}()

		for c.isConnected {
			time.Sleep(heartbeatIntervalInSec * time.Second)
			if err = stream.Send(&tracker.Heartbeat{Id: c.ClientId, Token: c.ClientToken}); err != nil {
				c.LastConnectionError = err
				return
			}
		}
	}()
	return serverAnswer.ErrorMessage, nil
}

//UnSynchDevice - отключить клиента от сервеа (оффлайн)
func (c *TrackerClient) UnSynchDevice() (clientErrorMessage string) {
	if !c.isConnected {
		return "Device was not connected"
	}
	c.isConnected = false
	return ""
}
