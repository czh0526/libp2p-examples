package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	"net/http"
	"sync"
)

var (
	IPFS_PEERS = convertPeers([]string{
		"/ip4/9.134.4.207/tcp/4001/p2p/12D3KooWT2eBrh3wfuUmNMsyeqmcqRGg5weAspKVmeKzrmHWjNQ6",
	})
	LOCAL_PEER_ENDPOINT = "http://localhost:5001/api/v0/id"
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

type IdOutput struct {
	ID              string
	PublicKey       string
	Addresses       []string
	AgentVersion    string
	ProtocolVersion string
}

func getLocalPeerInfo() []peer.AddrInfo {
	resp, err := http.PostForm(LOCAL_PEER_ENDPOINT, nil)
	if err != nil {
		panic(fmt.Sprintf("get local peer info failed: err = %v", err))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("read local peer info failed: err = %v", err))
	}

	var js IdOutput
	err = json.Unmarshal(body, &js)
	if err != nil {
		panic(fmt.Sprintf("parse local peer info failed: err = %v", err))
	}

	for _, addr := range js.Addresses {
		if addr[0:8] == "/ip4/127" {
			return convertPeers([]string{addr})
		}
	}
	return make([]peer.AddrInfo, 1)
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
				fmt.Printf("bootstrap dial: host = %v, peer = %v \n", ph.ID(), p.ID)
			}()

			fmt.Printf("%s bootstrapping to %s \n", ph.ID(), p.ID)

			ph.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
			if err := ph.Connect(ctx, p); err != nil {
				fmt.Printf("bootstrapDialFailed %s\n, err = %v \n", p.ID, err)
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
