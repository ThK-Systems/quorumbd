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

type CoreConnection struct {
	protocol Protocol
	address  string
}

func fromURI(uri string) (*CoreConnection, error) {
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

func (cc *CoreConnection) toURI() string {
	return string(cc.protocol) + "://" + cc.address
}

func (cc *CoreConnection) tryDial(ctx context.Context) (*CoreConnection, error) {

	dialer := net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, string(cc.protocol), cc.address)
	if err != nil {
		return nil, err
	}

	conn.Close()

	return cc, nil
}
