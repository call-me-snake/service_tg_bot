package httpServer

import (
	"fmt"
	"github.com/call-me-snake/service_tg_bot/client/internal/trackerClient"
	"github.com/gorilla/mux"
	"net/http"
)

type Connector struct {
	router  *mux.Router
	address string
}

//New - Конструктор *Connector
func New(addr string) *Connector {
	c := &Connector{}
	c.router = mux.NewRouter()
	c.address = addr
	return c
}

func (c *Connector) executeHandlers(t *trackerClient.TrackerClient) {
	c.router.HandleFunc("/device/info", clientDeviceInfo(t)).Methods("GET")
	c.router.HandleFunc("/device/register", registerClientDevice(t)).Methods("POST")
	c.router.HandleFunc("/device/token", getToken(t)).Methods("GET")
	c.router.HandleFunc("/device/connect", connect(t)).Methods("PUT")
	c.router.HandleFunc("/device/disconnect", disconnect(t)).Methods("PUT")
}

func (c *Connector) Start(t *trackerClient.TrackerClient) error {
	c.executeHandlers(t)
	err := http.ListenAndServe(c.address, c.router)
	return fmt.Errorf("httpServer.Start: %v", err)
}
