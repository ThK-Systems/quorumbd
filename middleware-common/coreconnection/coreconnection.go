package coreconnection

import (
	"fmt"
	"strings"
)

type Protocol string

const (
	ProtocolUnix Protocol = "unix"
	ProtocolTCP  Protocol = "tcp"
)

type CoreConnection struct {
	protocol Protocol
	address  string
}

func FromURI(uri string) (*CoreConnection, error) {
	unixPrefix := string(ProtocolUnix) + "://"
	tcpPrefix := string(ProtocolTCP) + "://"

	switch {
	case strings.HasPrefix(uri, unixPrefix):
		return &CoreConnection{
			protocol: ProtocolUnix,
			address:  strings.TrimPrefix(uri, unixPrefix),
		}, nil

	case strings.HasPrefix(uri, tcpPrefix):
		return &CoreConnection{
			protocol: ProtocolTCP,
			address:  strings.TrimPrefix(uri, tcpPrefix),
		}, nil
	}

	return nil, fmt.Errorf("invalid URI: %s", uri)
}
