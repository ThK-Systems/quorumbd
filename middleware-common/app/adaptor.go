package app

type Adaptor interface {
	GetName() string
	Connect() error
	Disconnect() error
	HandleConnection() error
}
