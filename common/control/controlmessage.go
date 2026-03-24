// Package control provides messages for the control plane
package control

import "fmt"

const (
	CMDummy = iota
)

type ControlMessage interface {
	Type() uint32
	RequestID() uint64
}

func NewMessage(messageType uint32) (ControlMessage, error) {
	switch messageType {
	default:
		return nil, fmt.Errorf("unknown type %d", messageType)
	}
}
