// Package transport provides interfaces for forwarding connections.
package transport

// Interface Transport should be a valid object, which is ready to open connection.
type Interface interface {
	// Connect initializes the connection with transport method. For example, if transport method
	// requires initial authentication, it should happen at this point, so further forward errors
	// are more specific.
	Connect() (Connected, error)
}

// Connected interface describes universal way of communicating with remote hosts
// using different transport protocols.
type Connected interface {
	// ForwardUnixSocket forwards unix socket to local machine to make it available for the process.
	ForwardUnixSocket(remotePath string) (localPath string, err error)

	// ForwardTCP listens on random local port and forwards incoming connections to given remote address.
	ForwardTCP(remoteAddr string) (localAddr string, err error)
}

// Config describes how Transport interface should be created.
type Config interface {
	// New returns new instance of Transport object.
	New() (Interface, error)

	// Validate should validate Transport configuration.
	Validate() error
}
