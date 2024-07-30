package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
	"time"
)

const NameSpace = "chat-test"
const PORT = 3001

func main() {
	ctx := context.Background()

	id := flag.Int("id", 0, "Source port number")
	flag.Parse()

	// 构造 Host
	basicHost, err := makeHost(*id, PORT)
	if err != nil {
		log.Println(err)
		return
	}

	// 监听协议
	startPeer(basicHost, handleStream)

	// 启动 DHT 服务，连接 bootstrap peers
	kadDHT, err := dht.New(ctx, basicHost, dht.BootstrapPeers(BOOTSTRAP_PEERS...))
	if err != nil {
		fmt.Printf("new DHT failed, err =%v\n", err)
		return
	}
	if err = kadDHT.Bootstrap(ctx); err != nil {
		fmt.Printf("KadDHT bootstrap failed, err =%v\n", err)
		return
	}
	time.Sleep(time.Second * 2)

	// 建立名空间，公布自己
	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)
	dutil.Advertise(ctx, routingDiscovery, NameSpace)

	// 查找其它的节点
	peerChan, err := routingDiscovery.FindPeers(ctx, NameSpace)
	if err != nil {
		fmt.Printf("FindPeers failed, err =%v\n", err)
		return
	}
	for peer := range peerChan {
		if peer.ID == basicHost.ID() {
			continue
		}
		fmt.Printf("Found peer: %s \n", peer.ID)

		stream, err := basicHost.NewStream(ctx, peer.ID, NameSpace)
		if err != nil {
			fmt.Printf("NewStream failed, err =%v\n", err)
			continue
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		go writeData(rw)
		go readData(rw)

		fmt.Printf("Connected to: %s\n", peer.ID)
	}

	select {}
}

func makeHost(id int, port int) (host.Host, error) {
	privKey, err := utils.GeneratePrivateKey(
		fmt.Sprintf("host%v.pem", id))
	if err != nil {
		log.Printf("Failed to generate private key, err = %v", err)
		return nil, err
	}

	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		log.Printf("Failed to generate multiaddr, err = %v", err)
		return nil, err
	}

	basicHost, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(sourceMultiAddr),
	)
	if err != nil {
		log.Printf("Failed to create libp2p host, err = %v", err)
		return nil, err
	}

	return basicHost, nil
}

func startPeer(h host.Host, streamHandler network.StreamHandler) {
	h.SetStreamHandler("/chat/1.0.0", streamHandler)
}

func handleStream(stream network.Stream) {
	fmt.Println("Got a new stream")

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')
		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("read data from console => %s\n", sendData)

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			log.Printf("write data to remote failed, err = %v", err)
			return
		}
		err = rw.Flush()
		if err != nil {
			log.Printf("flush data to remote failed, err = %v", err)
			return
		}
	}
}
