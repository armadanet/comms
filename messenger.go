package comms

import (
  "github.com/google/uuid"
)

type Instance struct {
  Id        *uuid.UUID
  Reciever  chan interface{}
}
type InstancePool map[uuid.UUID]*Instance
type Message struct {
  Success   chan bool
  Reciever  *uuid.UUID
  Data      interface{}
}

type Messenger struct {
  instances     InstancePool
  Register      chan *Instance
  Unregister    chan *Instance
  UnregisterId  chan *uuid.UUID
  Message       chan *Message
}

func NewMessenger() *Messenger{
  return &Messenger{
    instances: make(InstancePool),
    Register: make(chan *Instance),
    Unregister: make(chan *Instance),
    UnregisterId: make(chan *uuid.UUID),
    Message: make(chan *Message),
  }
}

func (m *Messenger) Start() {go m.run()}
func (m *Messenger) MakeInstance(rec chan interface{}) *Instance {
  id := uuid.New()
  return &Instance{
    Id: &id,
    Reciever: rec,
  }
}
func (m *Messenger) SendMessage(to *uuid.UUID, mes interface{}) bool {
  success := make(chan bool)
  m.Message <- &Message{
    Success: success,
    Reciever: to,
    Data: mes,
  }
  result := <-success
  return result
}

func (m *Messenger) run() {
  for {
    select {
    case instance := <- m.Register:
      m.instances[*instance.Id] = instance
    case instance := <- m.Unregister:
      // delete idempotent
      delete(m.instances, *instance.Id)
    case id := <- m.UnregisterId:
      delete(m.instances, *id)
    case mes := <- m.Message:
      if instance, ok := m.instances[*mes.Reciever]; ok {
        go func() {
          instance.Reciever <- mes.Data
          mes.Success <- true
        }()
      } else {go func() {mes.Success <- false}()}
    }
  }
}
