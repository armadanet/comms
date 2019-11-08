package comms

type Messenger interface {
  // Open the messenger service. Is open by default.
  Open()
  // Operate the messenger service
  Run()
  // Close and quit the messenger service. Can be re-opened and restarted.
  Close()
}
