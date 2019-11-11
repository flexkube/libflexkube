package transport

// Transport interface describes universal way of communicating with remote hosts
// using different transport protocols.
type Transport interface {
	// ForwardUnixSocket forwards unix socket to local machine to make it available for the process
	ForwardUnixSocket(string) (string, error)
}

// Config describes how Transport interface should be created.
type Config interface {
	// New returns new instance of Transport object.
	New() (Transport, error)
	// Validate should validate Transport configuration.
	Validate() error
}
