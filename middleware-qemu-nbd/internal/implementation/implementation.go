// Package implementation implements the adaptor interface of the middleware-common
package implementation

type implementation struct {
}

func New() *implementation {
	return &implementation{}
}

// GetName is an interface method
func (impl *implementation) GetName() string {
	return "qemu-nbd"
}

// Connect is an interface method
func (impl *implementation) Connect() error {
	return nil
}

// Disconnect is an interface method
func (impl *implementation) Disconnect() error {
	return nil
}

// HandleConnection is an interface method
func (impl *implementation) HandleConnection() error {
	return nil
}
