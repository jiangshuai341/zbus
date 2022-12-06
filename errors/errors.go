package errors

import "errors"

var (
	// ErrServiceShutdown occurs when server is closing.
	ErrServiceShutdown = errors.New("service is going to be shutdown")
	// ErrServiceInShutdown occurs when attempting to shut the server down more than once.
	ErrServiceInShutdown = errors.New("service is already in shutdown")
	// ErrAcceptSocket occurs when acceptor does not accept the new connection properly.
	ErrAcceptSocket = errors.New("accept a new connection error")
	// ErrTooManyEventLoopThreads occurs when attempting to set up more than 10,000 event-loop goroutines under LockOSThread mode.
	ErrTooManyEventLoopThreads = errors.New("too many event-loops under LockOSThread mode")
	// ErrUnsupportedProtocol occurs when trying to use protocol that is not supported.
	ErrUnsupportedProtocol = errors.New("only unix, tcp/tcp4/tcp6, udp/udp4/udp6 are supported")
	// ErrUnsupportedTCPProtocol occurs when trying to use an unsupported TCP protocol.
	ErrUnsupportedTCPProtocol = errors.New("only tcp/tcp4/tcp6 are supported")
	// ErrUnsupportedUDPProtocol occurs when trying to use an unsupported UDP protocol.
	ErrUnsupportedUDPProtocol = errors.New("only udp/udp4/udp6 are supported")
	// ErrUnsupportedUDSProtocol occurs when trying to use an unsupported Unix protocol.
	ErrUnsupportedUDSProtocol = errors.New("only unix is supported")
	// ErrUnsupportedPlatform occurs when running gnet on an unsupported platform.
	ErrUnsupportedPlatform = errors.New("unsupported platform in gnet")
	// ErrUnsupportedOp occurs when calling some methods that has not been implemented yet.
	ErrUnsupportedOp = errors.New("unsupported operation")
	// ErrNegativeSize occurs when trying to pass a negative size to a buffer.
	ErrNegativeSize = errors.New("negative size is invalid")
)
