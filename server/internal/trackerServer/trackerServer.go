package trackerServer

import (
	"context"
	"errors"
	"fmt"
	"github.com/call-me-snake/service_tg_bot/internal/tracker"
	"github.com/call-me-snake/service_tg_bot/server/internal/model"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	maxWaitTimeInSec     = 15
	incorrectInformation = "Некорректная информация об устройстве"
	connectionError      = "Ошибка подключения к базе устройств"
)

type trackerServer struct {
	tracker.UnimplementedTrackerServer
	deviceInfoStorage model.IClientDeviceStorage
}

func (s trackerServer) Register(ctx context.Context, registerMessage *tracker.TokenRequest) (*tracker.ServerResponse, error) {
	var response *tracker.ServerResponse
	if registerMessage.Id == 0 || registerMessage.Name == "" {
		response = &tracker.ServerResponse{ErrorMessage: incorrectInformation, Token: ""}
		return response, nil
	}
	token, err := s.deviceInfoStorage.RegisterNewDevice(&model.DeviceInfo{Id: registerMessage.Id, Name: registerMessage.Name})
	if err != nil {
		if strings.Contains(err.Error(), model.DevicesIdConstraint) {
			response = &tracker.ServerResponse{ErrorMessage: "Попытка повторной регистрации устройства", Token: ""}
		} else {
			response = &tracker.ServerResponse{ErrorMessage: connectionError, Token: ""}
			log.Print(err.Error())
		}
		return response, nil
	}
	response = &tracker.ServerResponse{ErrorMessage: "", Token: token}
	return response, nil
}

func (s trackerServer) GetToken(ctx context.Context, askTokenMessage *tracker.TokenRequest) (*tracker.ServerResponse, error) {
	var response *tracker.ServerResponse
	if askTokenMessage.Id == 0 || askTokenMessage.Name == "" {
		response = &tracker.ServerResponse{ErrorMessage: incorrectInformation, Token: ""}
		return response, nil
	}
	device, err := s.deviceInfoStorage.GetDeviceInfo(askTokenMessage.Id)
	if err != nil {
		response = &tracker.ServerResponse{ErrorMessage: connectionError, Token: ""}
	} else if device == nil {
		response = &tracker.ServerResponse{ErrorMessage: fmt.Sprintf("Устройство с id: %d не найдено", askTokenMessage.Id), Token: ""}
	} else if device.Name != askTokenMessage.Name {
		response = &tracker.ServerResponse{ErrorMessage: "Некорректные данные устройства", Token: ""}
	} else {
		response = &tracker.ServerResponse{ErrorMessage: "", Token: device.Token}
	}
	return response, nil
}

func (s trackerServer) Synch(stream tracker.Tracker_SynchServer) error {

	ctx := stream.Context()

	clientMsg, err := stream.Recv()
	if err != nil {
		log.Print(err)
		if err == io.EOF {
			return nil
		}
	}
	serverResponce := tracker.HeartbeatResponse{Synched: false}
	deviceData, err := s.deviceInfoStorage.GetDeviceInfo(clientMsg.Id)
	switch {
	case err != nil:
		log.Print(err)
		serverResponce.ErrorMessage = connectionError
	case deviceData == nil:
		serverResponce.ErrorMessage = fmt.Sprintf("Клиента с id: %d не найдено", clientMsg.Id)
	case clientMsg.Token != deviceData.Token:
		serverResponce.ErrorMessage = "Не совпадают токены"
	case deviceData.Online:
		serverResponce.ErrorMessage = fmt.Sprintf("Клиент уже онлайн")
	default:
		isSet, err := s.deviceInfoStorage.SetOnlineStatus(clientMsg.Id, true)
		if err != nil {
			serverResponce.ErrorMessage = connectionError
		} else if !isSet {
			serverResponce.ErrorMessage = fmt.Sprintf("Клиента с id: %d не найдено", clientMsg.Id)
		} else {
			serverResponce.Synched = true
		}
	}
	if serverResponce.Synched {
		defer func() {
			isSet, err := s.deviceInfoStorage.SetOnlineStatus(clientMsg.Id, false)
			if err != nil {
				log.Print(err)
			}
			if !isSet {
				log.Print(fmt.Sprintf("Оффлайн статус для клиента с id: %d не установлен", clientMsg.Id))
			}
		}()
	}
	err = stream.Send(&serverResponce)
	if err != nil {
		log.Print(err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("Synch: %v", ctx.Err())
		case <-time.After(maxWaitTimeInSec * time.Second):
			return errors.New("Synch: connection aborted by timeout")
		default:
			_, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				} else {
					log.Print(err)
					return err
				}
			}
		}
	}
}

func StartServer(port string, storage model.IClientDeviceStorage) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("StartServer: %v", err)
	}
	s := trackerServer{deviceInfoStorage: storage}
	grpcServer := grpc.NewServer()
	tracker.RegisterTrackerServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("StartServer: %v", err)
	}
	return nil
}
