// Package implementation implements the adaptor interface of middleware-common
package implementation

type Implementation struct {
}

func New() *Implementation {
	return &Implementation{}
}

// GetImplementationName is an interface method
func (impl *Implementation) GetImplementationName() string {
	return "qemu-nbd"
}

// Listen is an interface method
func (impl *Implementation) Listen() error {
	return nil
}

// Connect is an interface method
func (impl *Implementation) Connect() error {
	return nil
}

// Disconnect is an interface method
func (impl *Implementation) Disconnect() error {
	return nil
}
