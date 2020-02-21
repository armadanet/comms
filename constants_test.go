package comms_test

import (
  "testing"
  "github.com/armadanet/comms"
)

func TestConstantConsitency(t *testing.T) {
  if !(comms.PongWait > comms.PingPeriod) {
    t.Errorf("Pong should wait longer than ping.\n")
  }
}
