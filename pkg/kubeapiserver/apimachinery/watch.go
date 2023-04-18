package apimachinery

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"minik8s/pkg/kubeapiserver/storage"
	"net/http"
)

var Storage = storage.NewEtcdStorageNoParam()

// WatchServer WebSocket server
type WatchServer struct {
	Conn *websocket.Conn
}

// NewWatchServer create a new WebSocket server
func NewWatchServer(c *gin.Context) (*WatchServer, error) {
	// update HTTP connection to WebSocket connection
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return nil, err
	}

	return &WatchServer{Conn: conn}, nil
}

// Read websocket message
func (s *WatchServer) Read() ([]byte, error) {
	_, message, err := s.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Write websocket message
func (s *WatchServer) Write(message []byte) error {
	err := s.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return err
	}

	return nil
}

// Close websocket connection
func (s *WatchServer) Close() error {
	return s.Conn.Close()
}

// Watch a etcd key
func (s *WatchServer) Watch(key string) error {
	// TODO: concorrency problem?
	err := Storage.Watch(context.Background(), key, func(key string, value []byte) error {
		innerErr := s.Write(value)
		if innerErr != nil {
			log.Error("[Watch] write message error: ", innerErr)
			return innerErr
		}
		return nil
	})
	return err
}
