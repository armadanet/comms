package comms

import (
  "github.com/gorilla/websocket"
  "net/http"
  "time"
  "log"
  "os"
  "os/signal"
  "reflect"
)

const (
  WriteDeadline = 10 * time.Second
  ReadDeadline  = WriteDeadline
  PongWait      = 60 * time.Second
  PingPeriod    = (PongWait * 9) / 10
  MessageSize   = 4096
)

type Socket interface {
  Reader()  chan(interface{})
  Writer()  chan(interface{})
  Start(read interface{})
  Close()
}

type socket struct {
  conn        *websocket.Conn
  reader      chan interface{}
  writer      chan interface{}
  quit        chan interface{}
}

func NewSocket(conn *websocket.Conn) Socket {
  s := &socket{
    conn: conn,
    reader: make(chan interface{}),
    writer: make(chan interface{}),
    quit: make(chan interface{}),
  }
  return s
}

func EstablishSocket(dialurl string) (Socket, error) {
  conn, _, err := websocket.DefaultDialer.Dial(dialurl, nil)
  if err != nil {
    return nil, err
  }
  return NewSocket(conn), nil
}

func AcceptSocket(w http.ResponseWriter, r *http.Request) (Socket, error) {
  upgrader := websocket.Upgrader{
    ReadBufferSize: 1024,
    WriteBufferSize: 1024,
  }
  conn, err := upgrader.Upgrade(w,r,nil)
  if err != nil {
    return nil, err
  }
  return NewSocket(conn), nil
}

func (s *socket) Start(read interface{}) {
  go s.readroutine(reflect.TypeOf(read))
  go s.writeroutine()
}
func (s *socket) Reader() chan interface{} {return s.reader}
func (s *socket) Writer() chan interface{} {return s.writer}
func (s *socket) Close() {close(s.quit)}

func (s *socket) readroutine(t reflect.Type) {
  defer func() {
    s.conn.Close()
  }()
  s.conn.SetReadLimit(MessageSize)
  s.conn.SetReadDeadline(time.Now().Add(PongWait))
  s.conn.SetPongHandler(func(string) error {
    s.conn.SetReadDeadline(time.Now().Add(PongWait))
    return nil
  })
  for {
    v := reflect.New(t).Interface()
    err := s.conn.ReadJSON(v)
    if err != nil {
      log.Println(err)
      return
    }
    s.reader <- v
  }
}

func (s *socket) writeroutine() {
  ticker := time.NewTicker(PingPeriod)
  interrupt := make(chan os.Signal, 1)
  signal.Notify(interrupt, os.Interrupt)
  defer func() {
    ticker.Stop()
    s.conn.Close()
    signal.Stop(interrupt)
  }()
  for {
    select {
    case data, ok := <- s.writer:
      s.conn.SetWriteDeadline(time.Now().Add(WriteDeadline))
      if !ok {
        s.conn.WriteMessage(websocket.CloseMessage, []byte{})
        return
      }
      err := s.conn.WriteJSON(data)
      if err != nil {
        log.Println(err)
        s.conn.WriteMessage(websocket.CloseMessage, []byte{})
        return
      }
    case <- ticker.C:
      s.conn.SetWriteDeadline(time.Now().Add(WriteDeadline))
      if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
        log.Println(err)
        return
      }
    case <- interrupt:
      err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
      if err != nil {
        log.Println(err)
        return
      }
      select {
      case <- time.After(4*time.Second):
      }
      return
    }
  }
}
