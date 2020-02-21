package comms

import "time"

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
