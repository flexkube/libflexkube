package container

// Container interface describes universal way of starting containers
type Container interface {
	Start(*Config) error
	Stop(string) error
	Status(string) (*Status, error)
}

// Config describes how container should be started
type Config struct {
	Name  string
	Image string
}

// Status describes what informations are returned about container
type Status struct {
	Image string
}
