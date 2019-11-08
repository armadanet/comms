package comms

func Loop(f func()) {for {f()}}
func Routine(f func()) {go Loop(f)}

func Reader(f func(data interface{}, ok bool), c chan interface{}) {
  Routine(func() {
    select {
    case data, ok := <- c:
      f(data, ok)
    }
  })
}
