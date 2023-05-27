package watch

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	//"minik8s/pkg/kubeapiserver/storage"
	"net/http"
)

//var Storage = storage.NewEtcdStorageNoParam()

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


func ListWatch(watchKey string, value []byte) error {

	list, ok := WatchStorage.Load(watchKey)
	if !ok {
		log.Error("[ListWatch] key: ", watchKey, " not found")
		return nil
	}
	if threadList, ok := list.(*ThreadSafeList); ok {
		for e := threadList.List.Front(); e != nil; e = e.Next() {
			if server, ok := e.Value.(*WatchServer); ok {
				err := server.Write(value)
				if err != nil {
					log.Warn("[ListWatch] Write message error: ", err)
				}
			}
		}
	}
	return nil
}
