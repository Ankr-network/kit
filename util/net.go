package util

import (
	"net"
)

func MustTcpListen(address string) net.Listener {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	return lis
}
