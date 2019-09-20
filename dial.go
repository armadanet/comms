package comms

import (
  "github.com/gorilla/websocket"
)

type Dailer interface {
  Dail(string) error
  Reader()
  Writer()
  Close()
  Conn() *websocket.Conn
}

type dailer struct {
  conn    *websocket.Conn
  
}

func (d *dailer) Dail(urlstr string) error {
  conn, _, err := websocket.DefaultDialer.Dial(urlstr, nil)
  if err != nil {
    return err
  }

}
