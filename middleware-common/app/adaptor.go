package app

type Adaptor interface {
	GetImplementationName() string
	Listen() error
	Disconnect() error
}

// TODO: Receive
// TODO: Reply
// TODO: Common structs and mapper for receive and reply

// TODO: Maybe HandleConnection
