package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"io"
	mrand "math/rand"
	"strings"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listenF := flag.Int("l", 0, "wait for incoming connections")
	targetF := flag.String("d", "", "target peer to dial")
	insecureF := flag.Bool("insecure", false, "use an unencrypted connection")
	seedF := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenF == 0 {
		panic("Please provide a port to bind on with -l")
	}

	ha, err := makeBasicHost(*listenF, *insecureF, *seedF)
	if err != nil {
		panic(fmt.Sprintf("make basic host failed: err = %v", err))
	}

	if *targetF == "" {
		startListener(ctx, ha, *listenF, *insecureF)
		<-ctx.Done()
	} else {
		runSender(ctx, ha, *targetF)
	}
}

func makeBasicHost(listenPort int, insecure bool, randSeed int64) (host.Host, error) {
	var r io.Reader
	if randSeed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randSeed))
	}

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
	}
	if insecure {
		opts = append(opts, libp2p.NoSecurity)
	}

	return libp2p.New(opts...)
}

func getHostAddress(ha host.Host) string {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", ha.ID()))

	var addr ma.Multiaddr
	for _, a := range ha.Addrs() {
		if !strings.Contains(a.String(), "127.0.0.1") {
			addr = a
			break
		}
	}
	return addr.Encapsulate(hostAddr).String()
}

func startListener(ctx context.Context, ha host.Host, listenPort int, insecure bool) {
	fullAddr := getHostAddress(ha)
	fmt.Printf("I am %s \n", fullAddr)

	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		fmt.Printf("Got a new stream from %s\n", s.Conn().RemotePeer())
		if err := doEcho(s); err != nil {
			fmt.Printf("Echo error: %v\n", err)
			s.Reset()
		} else {
			s.Close()
		}
	})
	fmt.Println("Listening for connections")

	if insecure {
		fmt.Printf("Now run './echo -l %d -d %s -insecure' on a different terminal\n", listenPort+1, fullAddr)
	} else {
		fmt.Printf("Now run './echo -l %d -d %s' on a different terminal\n", listenPort+1, fullAddr)
	}
}

func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Printf("read: %s\n", str)
	_, err = s.Write([]byte(str))
	return err
}

func runSender(ctx context.Context, ha host.Host, targetPeer string) {
	fullAddr := getHostAddress(ha)
	fmt.Printf("I am %s \n", fullAddr)

	maddr, err := ma.NewMultiaddr(targetPeer)
	if err != nil {
		fmt.Printf("parse targetPeer failed: err = %v\n", err)
		return
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		fmt.Printf("get addr info failed: err = %v\n", err)
		return
	}
	ha.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	fmt.Println("sender opening stream")
	s, err := ha.NewStream(ctx, info.ID, protocol.ID("/echo/1.0.0"))
	if err != nil {
		fmt.Printf("new stream failed: err = %v\n", err)
		return
	}

	fmt.Println("sender saying hello")
	_, err = s.Write([]byte("hello, world!\n"))
	if err != nil {
		fmt.Printf("say hello failed: err = %v\n", err)
		return
	}

	out, err := io.ReadAll(s)
	if err != nil {
		fmt.Printf("read echo failed: err = %v\n", err)
		return
	}

	fmt.Printf("read reply: %q\n", string(out))

}
