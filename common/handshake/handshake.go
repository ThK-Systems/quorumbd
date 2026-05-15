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

const (
	SourceCore uint8 = iota + 1
	SourceMiddlewareCommon
)

func SourceName(source uint8) string {
	switch source {
	case SourceCore:
		return "CORE"
	case SourceMiddlewareCommon:
		return "MIDDLEWARE-COMMON"
	default:
		return "UNKNOWN"
	}
}

type Handshake struct {
	Magic    [3]byte
	Version  uint16
	Source   uint8
	Type     string
	UUID     uuid.UUID
	Reserved [6]byte
}

func New(source uint8, handshakeType string, id uuid.UUID) []byte {
	result := make([]byte, Size)
	copy(result[0:3], Magic[:])
	binary.BigEndian.PutUint16(result[3:5], Version)
	result[5] = source
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
		Source:   data[5],
		Type:     string(data[6:10]),
		UUID:     uuid.UUID(data[10:26]),
		Reserved: reserved,
	}, nil
}
