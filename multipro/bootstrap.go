package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"sync"
)

var (
	BOOTSTRAP_PEERS = convertPeers([]string{
		"/ip4/9.134.4.207/tcp/8080/p2p/12D3KooWCSvzkfKPH5SgtBb2m5Sk5MSqwmDsVcMS5ieYhAKtaq5x",
	})
)

func convertPeers(peers []string) []peer.AddrInfo {
	pinfos := make([]peer.AddrInfo, len(peers))
	for i, addr := range peers {
		maddr := ma.StringCast(addr)
		p, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			panic(fmt.Sprintf("parse ipfs bootstrap peers failed: err = %v", err))
		}
		pinfos[i] = *p
	}

	return pinfos
}

func bootstrapConnect(ctx context.Context, ph host.Host, peers []peer.AddrInfo) error {
	if len(peers) < 1 {
		return errors.New("not enough bootstrap peers")
	}

	errs := make(chan error, len(peers))
	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		go func(p peer.AddrInfo) {
			defer func() {
				wg.Done()
			}()

			fmt.Printf("local host connect to => `%s` \n", p)
			ph.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
			if err := ph.Connect(ctx, p); err != nil {
				fmt.Printf("bootstrapDialFailed %s, err = %v \n", p.ID, err)
				errs <- err
				return
			}
		}(p)
	}
	wg.Wait()

	close(errs)
	count := 0
	var err error
	for err = range errs {
		if err != nil {
			count++
		}
	}
	if count == len(peers) {
		return fmt.Errorf("failed to bootstrap. %s", err)
	}

	return nil
}
