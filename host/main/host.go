package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"os"
	"time"
)

func main() {

	var idht *dht.IpfsDHT
	priv, err := generatePrivateKey("privkey.pem")

	connmgr, err := connmgr.NewConnManager(
		100,
		400,
		connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		panic(err)
	}

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",
			"/ip4/0.0.0.0/udp/9000/quic"),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.ConnectionManager(connmgr),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(context.Background(), h)
			return idht, err
		}),
		libp2p.EnableNATService(),
	)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	peerInfo := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("libp2p peer.ID = %v, peerAddrs = %v, node address: %v",
		h.ID(), h.Addrs(), addrs)

}

func generatePrivateKey(filename string) (crypto.PrivKey, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			privateKey, _, err := crypto.GenerateKeyPair(
				crypto.Ed25519, -1)
			if err != nil {
				return nil, fmt.Errorf("create priv key failed, err = %v", err)
			}
			privateBytes, err := crypto.MarshalPrivateKey(privateKey)
			if err != nil {
				return nil, fmt.Errorf("marshal priv key failed, err = %v", err)
			}
			err = os.WriteFile(filename, privateBytes, 0600)
			if err != nil {
				return nil, fmt.Errorf("write priv key failed, err = %v", err)
			}

			return privateKey, nil
		}
		return nil, err
	}

	privateBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read priv key failed, err = %v", err)
	}

	privateKey, err := crypto.UnmarshalPrivateKey(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("unmarshal priv key failed, err = %v", err)
	}

	return privateKey, nil
}
