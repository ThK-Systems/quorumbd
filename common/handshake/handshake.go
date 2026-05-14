// Package handshake provides the minimal core connection handshake.
package handshake

import (
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

const (
	Magic   uint16 = 0x5142
	Version uint16 = 0x0001
	Size           = 32
)

const (
	TypeProbe   = "PROB"
	TypeControl = "CTRL"
)

type Handshake struct {
	Magic   uint16
	Version uint16
	Type    string
	UUID    uuid.UUID
	Reserved [8]byte
}

func New(handshakeType string, id uuid.UUID) []byte {
	result := make([]byte, Size)
	binary.BigEndian.PutUint16(result[0:2], Magic)
	binary.BigEndian.PutUint16(result[2:4], Version)
	copy(result[4:8], handshakeType)
	copy(result[8:24], id[:])
	return result
}

func Parse(data []byte) (Handshake, error) {
	if len(data) < Size {
		return Handshake{}, fmt.Errorf("handshake too short: %d < %d", len(data), Size)
	}

	magic := binary.BigEndian.Uint16(data[0:2])
	if magic != Magic {
		return Handshake{}, fmt.Errorf("invalid handshake magic: 0x%04x", magic)
	}

	var reserved [8]byte
	copy(reserved[:], data[24:32])

	return Handshake{
		Magic:    magic,
		Version:  binary.BigEndian.Uint16(data[2:4]),
		Type:     string(data[4:8]),
		UUID:     uuid.UUID(data[8:24]),
		Reserved: reserved,
	}, nil
}
