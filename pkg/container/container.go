package container

// Container interface describes universal way of starting containers
type Container interface {
	// Create creates container and returns it's unique identifier
	Create(*Config) (string, error)
	// Delete removes the container
	Delete(string) error
	// Start starts created container
	Start(string) error
	// Status returns status of the container
	Status(string) (*Status, error)
	// Stop takes unique identifier as a parameter and stops the container
	Stop(string) error
}

// Config describes how container should be started
type Config struct {
	Name  string
	Image string
}

// Status describes what informations are returned about container
type Status struct {
	Image  string `json:"image"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
