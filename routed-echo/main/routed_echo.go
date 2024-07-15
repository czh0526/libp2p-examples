package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"io"
	"time"
)

func main() {
	target := flag.String("d", "", "target peer to dial")
	global := flag.Bool("global", false, "use global ipfs peers for bootstrapping")
	flag.Parse()

	var bootstrapPeers []peer.AddrInfo
	var globalFlag string
	if *global {
		fmt.Println("using global bootstrap")
		bootstrapPeers = IPFS_PEERS
		globalFlag = " -global"
	} else {
		fmt.Println("using local bootstrap")
		bootstrapPeers = LOCAL_PEERS
		globalFlag = ""
	}
	ha, _, err := makeRoutedHost(bootstrapPeers, globalFlag)
	if err != nil {
		panic(fmt.Sprintf("make routed host failed: err = %v", err))
	}

	if *target == "" {
		ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
			fmt.Printf("new echo stream from %s\n", s.Conn().RemotePeer())
			if err := doEcho(s); err != nil {
				fmt.Printf("echo failed: err = %v\n", err)
				s.Reset()
			} else {
				s.Close()
			}
		})

		fmt.Println("Listening for connections")
		fmt.Printf("Now run `./routed-echo -d %s%s` on a different terminal\n",
			ha.ID(), globalFlag)
		select {}
	} else {
		peerid, err := peer.Decode(*target)
		if err != nil {
			panic(fmt.Sprintf("peer(`%v`) decode failed: err = %s", *target, err))
		}
		fmt.Printf("\nI want connect to => `%s`\n", peerid)

		// 不停的尝试连接目标节点
		var s network.Stream
		for {
			fmt.Printf("opening stream(`/echo/1.0.0`) => %v\n", peerid.String())
			s, err = ha.NewStream(context.Background(), peerid, "/echo/1.0.0")
			if err != nil {
				fmt.Printf("peer(`%v`) create stream failed: err = %v \n", *target, err)
				time.Sleep(3 * time.Second)
				continue
			}
			break
		}

		_, err = s.Write([]byte("hello world!\n"))
		if err != nil {
			panic(fmt.Sprintf("peer(`%v`) write stream failed: err = %v", *target, err))
		}

		out, err := io.ReadAll(s)
		if err != nil {
			panic(fmt.Sprintf("peer(`%v`) read stream failed: err = %v", *target, err))
		}

		fmt.Printf("read reply: %q\n", out)
	}
}

func makeRoutedHost(bootstrapPeers []peer.AddrInfo,
	globalFlag string) (*rhost.RoutedHost, *kaddht.IpfsDHT, error) {

	r := rand.Reader
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
	if err != nil {
		return nil, nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9900"),
		libp2p.Identity(priv),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	ctx := context.Background()
	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, nil, err
	}

	// make the routed host
	//dstore := dsync.MutexWrap(ds.NewMapDatastore())
	//dht := kaddht.NewDHT(ctx, basicHost, dstore)
	dht, err := kaddht.New(ctx, basicHost)
	if err != nil {
		return nil, nil, fmt.Errorf("new dht failed: err = %v", err)
	}
	routedHost := rhost.Wrap(basicHost, dht)

	// bootstrap the host
	err = dht.Bootstrap(ctx)
	if err != nil {
		return nil, nil, err
	}

	time.Sleep(5 * time.Second)

	// connect to the ipfs nodes
	err = bootstrapConnect(ctx, routedHost, bootstrapPeers)
	if err != nil {
		return nil, nil, err
	}

	return routedHost, dht, nil
}

func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Printf("read: %s\n", str)
	_, err = s.Write([]byte(str))
	return nil
}
