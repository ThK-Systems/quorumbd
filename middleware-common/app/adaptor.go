package app

type Adaptor interface {
	GetImplementationName() string
	IsServer() bool
	Listen() error
}

// TODO: Receive
// TODO: Reply
// TODO: Common structs and commands and mapper for receive and reply

// TODO: Maybe HandleConnection
