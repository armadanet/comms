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

// These values are used by other Go packages.
const (
  WriteDeadline   = 10 * time.Second
  ReadDeadline    = WriteDeadline
  PongWait        = 60 * time.Second
  PingPeriod      = (PongWait * 9) / 10
  MessageSize     = 4096
  ReadBufferSize  = 1024
  WriteBufferSize = 1024
  InteruptPeriod  = 4 * time.Second
)

// Socket interface defines how code can
// interact with the sockets.
type Socket interface {
  Reader()  chan(interface{})
  Writer()  chan(interface{})
  Start(read interface{})
  Close()
}

// Hidden structure that holds the actual
// socket information.
type socket struct {
  conn        *websocket.Conn
  reader      chan interface{}
  writer      chan interface{}
  quit        chan interface{}
}

// Mechanism to convert a connection into a Socket.
// This can be used by outside packages, but operates
// as a common final action of the EstablishSocket
// and AcceptSocket functions.
func NewSocket(conn *websocket.Conn) Socket {
  s := &socket{
    conn: conn,
    reader: make(chan interface{}),
    writer: make(chan interface{}),
    quit: make(chan interface{}),
  }
  return s
}

// Connect to a websocket by the url.
func EstablishSocket(dialurl string) (Socket, error) {
  conn, _, err := websocket.DefaultDialer.Dial(dialurl, nil)
  if err != nil {
    return nil, err
  }
  return NewSocket(conn), nil
}

// Use the server types (http) to upgrade a connect
// to socket. This is the golang paradigm with the
// gorilla package.
func AcceptSocket(w http.ResponseWriter, r *http.Request) (Socket, error) {
  upgrader := websocket.Upgrader{
    ReadBufferSize: ReadBufferSize,
    WriteBufferSize: WriteBufferSize,
  }
  conn, err := upgrader.Upgrade(w,r,nil)
  if err != nil {
    return nil, err
  }
  return NewSocket(conn), nil
}

// Gorilla websockets require reading from a websocket
// to occur on a single goroutine, the same with write.
// Dealing with JSON types directly is difficult, so
// by supplying a variable of the type that will be read,
// the already established JSON function can handle that.
// NOTE: reflect package has some performance concerns,
// but is used in JSON package anyway, so of no concern.
func (s *socket) Start(read interface{}) {
  go s.readroutine(reflect.TypeOf(read))
  go s.writeroutine()
}

// Return the interface channel to read from or write to.
// This allows the application to do whatever custom work,
// with the only interaction being a channel that can take
// in any type.
func (s *socket) Reader() chan interface{} {return s.reader}
func (s *socket) Writer() chan interface{} {return s.writer}

// Closes the socket. After which the pointer should not be
// used again.
func (s *socket) Close() {close(s.quit)}

// Basic connection setup. Read in data of the
// given type from the socket to pass to the
// channel. See the comment on the Start function
// for input type explanation. Closes the reader
// if any error is detected. Quitting does not close
// this function, since a channel blocks operation and
// would halt the for loop, but the connection closing
// from the writeroutine *should* cause the ReadJSON
// operation to fail and end this goroutine.
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

// Basic connection setup. Anything sent to the
// writer channel is sent through the socket.
// Closes the writer if any error is detected.
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
    case <- s.quit:
      log.Println("Quitting write channel.")
      err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
      if err != nil {
        log.Println(err)
        return
      }
      return
    case <- interrupt:
      err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
      if err != nil {
        log.Println(err)
        return
      }
      select {
      case <- time.After(InteruptPeriod):
      }
      return
    }
  }
}
