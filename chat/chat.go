package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/czh0526/libp2p-examples/utils"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
)

func main() {
	sourcePort := flag.Int("sp", 0, "Source port number")
	dest := flag.String("d", "", "Destination multiaddr string")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		fmt.Printf("This program demonstrates a simple p2p chat application using libp2p\n\n")
		fmt.Println("Usage: Run './chat -sp <SOURCE_PORT>' where <SOURCE_PORT> can be any port number.")
		fmt.Println("Now run './chat -d <MULTIADDR>' where <MULTIADDR> is multiaddress of previous listener host.")

		os.Exit(0)
	}

	if *dest == "" {
		basicHost, err := makeNode(1, *sourcePort)
		if err != nil {
			log.Println(err)
			return
		}

		startPeer(basicHost, handleStream)

	} else {
		basicHost, err := makeNode(2, *sourcePort)
		if err != nil {
			log.Println(err)
			return
		}

		rw, err := startPeerAndConnect(basicHost, *dest)
		if err != nil {
			log.Println(err)
			return
		}

		go writeData(rw)
		go readData(rw)
	}

	// wait forever
	select {}
}

func handleStream(s network.Stream) {
	fmt.Println("Got a new stream !")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

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

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			return
		}
		err = rw.Flush()
		if err != nil {
			return
		}
	}
}

func makeNode(id int, port int) (host.Host, error) {
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

	var port string
	for _, la := range h.Network().ListenAddresses() {
		if p, err := la.ValueForProtocol(multiaddr.P_TCP); err == nil {
			port = p
			break
		}
	}

	if port == "" {
		log.Println("was not able to find actual local port")
		return
	}

	log.Printf("Run './chat -d /ip4/127.0.0.1/tcp/%v/p2p/%s' on another console.\n", port, h.ID())
	log.Println("You can replace 127.0.0.1 with public IP as well.")
	log.Println("Waiting for incoming connection")
	log.Println()
}

func startPeerAndConnect(h host.Host, dest string) (*bufio.ReadWriter, error) {
	log.Printf("This node's multiaddresses:")
	for _, la := range h.Addrs() {
		log.Printf("- %s \n", la)
	}

	maddr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		log.Printf("new multiaddr failed, err = %v", err)
		return nil, err
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Printf("new peer info failed, err = %v", err)
		return nil, err
	}

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	s, err := h.NewStream(context.Background(), info.ID, "/chat/1.0.0")
	if err != nil {
		log.Printf("new stream failed, err = %v", err)
		return nil, err
	}
	log.Println("Established connection to destination")

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	return rw, nil
}
