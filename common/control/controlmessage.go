// Package control provides messages for the control plane
package control

import (
	"fmt"
)

const (
	CMDummy = iota
)

type ControlMessage interface {
	Type() uint32
	RequestID() uint64
	isResponse() bool
	GeneratedAt() uint64 // Milliseconds since 1970
	GeneratedBy() uint32 // UUID of instance
}

type BaseControlMessage struct {
	Type        uint32 `json:"type"`
	RequestID   uint64 `json:"request_id"`
	IsResponse  bool   `json:"is_response"`
	GeneratedAt uint64 `json:"generated_at"`
	GeneratedBy uint32 `json:"generated_by"`
}

func NewMessage(messageType uint32) (ControlMessage, error) {
	switch messageType {
	default:
		return nil, fmt.Errorf("unknown type %d", messageType)
	}
}
