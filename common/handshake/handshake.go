// Package handshake provides the minimal core connection handshake.
package handshake

import (
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

const Version uint16 = 0x0001

const Size = 32

var Magic = [3]byte{'Q', 'B', 'D'}

const (
	TypeProbe   = "PROB"
	TypeControl = "CTRL"
)

type System uint8

const (
	SystemCore System = iota + 1
	SystemMiddlewareCommon
)

func (system System) String() string {
	switch system {
	case SystemCore:
		return "CORE"
	case SystemMiddlewareCommon:
		return "MIDDLEWARE-COMMON"
	default:
		return "UNKNOWN"
	}
}

func (system System) IsKnown() bool {
	switch system {
	case SystemCore, SystemMiddlewareCommon:
		return true
	default:
		return false
	}
}

type Handshake struct {
	Magic    [3]byte
	Version  uint16
	System   System
	Type     string
	UUID     uuid.UUID
	Reserved [6]byte
}

func New(system System, handshakeType string, id uuid.UUID) []byte {
	result := make([]byte, Size)
	copy(result[0:3], Magic[:])
	binary.BigEndian.PutUint16(result[3:5], Version)
	result[5] = byte(system)
	copy(result[6:10], handshakeType)
	copy(result[10:26], id[:])
	return result
}

func Parse(data []byte) (Handshake, error) {
	if len(data) < Size {
		return Handshake{}, fmt.Errorf("handshake too short: %d < %d", len(data), Size)
	}

	var magic [3]byte
	copy(magic[:], data[0:3])
	if magic != Magic {
		return Handshake{}, fmt.Errorf("invalid handshake magic: %q", string(magic[:]))
	}

	var reserved [6]byte
	copy(reserved[:], data[26:32])

	return Handshake{
		Magic:    magic,
		Version:  binary.BigEndian.Uint16(data[3:5]),
		System:   System(data[5]),
		Type:     string(data[6:10]),
		UUID:     uuid.UUID(data[10:26]),
		Reserved: reserved,
	}, nil
}
