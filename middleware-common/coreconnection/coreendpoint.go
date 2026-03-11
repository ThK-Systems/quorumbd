package coreconnection

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type Protocol string

const (
	ProtocolUnix Protocol = "unix"
	ProtocolTCP  Protocol = "tcp"
)

type CoreEndpoint struct {
	protocol Protocol
	address  string
}

func fromURI(uri string) (*CoreEndpoint, error) {
	unixPrefix := string(ProtocolUnix) + "://"
	tcpPrefix := string(ProtocolTCP) + "://"

	switch {
	case strings.HasPrefix(uri, unixPrefix):
		return &CoreEndpoint{
			protocol: ProtocolUnix,
			address:  strings.TrimPrefix(uri, unixPrefix),
		}, nil

	case strings.HasPrefix(uri, tcpPrefix):
		return &CoreEndpoint{
			protocol: ProtocolTCP,
			address:  strings.TrimPrefix(uri, tcpPrefix),
		}, nil
	}

	return nil, fmt.Errorf("invalid URI: %s", uri)
}

func (ce *CoreEndpoint) toURI() string {
	return string(ce.protocol) + "://" + ce.address
}

func (ce *CoreEndpoint) tryDial(ctx context.Context) (*CoreEndpoint, error) {

	dialer := net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, string(ce.protocol), ce.address)
	if err != nil {
		return nil, err
	}

	conn.Close()

	return ce, nil
}
