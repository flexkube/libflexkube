package transport

// Transport interface describes universal way of copying files and communicating with remote hosts
// using different transport protocols.
type Transport interface {
	// ForwardUnixSocket forwards unix socket to local machine to make it available for the process
	ForwardUnixSocket(string) (string, error)
}

type TransportConfig interface {
	New() (Transport, error)
}
