package main

import (
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	BOOTSTRAP_PEERS = []ma.Multiaddr{
		convertPeers("/ip4/1.13.245.16/tcp/8080/p2p/QmWiG7ExhxNokqzghHrxC25m3W8gVEftgcrZsJKhPv1Y74"),
	}
)

func convertPeers(address string) ma.Multiaddr {
	multiAddr, err := ma.NewMultiaddr(address)
	if err != nil {
		panic(fmt.Sprintf("malformed multiaddr string, err = %v", err))
	}

	return multiAddr
}
