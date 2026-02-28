// Package implementation implements the adaptor interface of the middleware-common
package implementation

type implementation struct {
}

func New() *implementation {
	return &implementation{}
}

// GetImplementationName is an interface method
func (impl *implementation) GetImplementationName() string {
	return "qemu-nbd"
}

// Listen is an interface method
func (impl *implementation) Listen() error {
	return nil
}

// Connect is an interface method
func (impl *implementation) Connect() error {
	return nil
}

// Disconnect is an interface method
func (impl *implementation) Disconnect() error {
	return nil
}
