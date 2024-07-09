package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"io"
	"net/http"
	"strings"
)

const Protocol = "/proxy-example/0.0.1"

func makeRandomHost(port int) host.Host {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)))
	if err != nil {
		panic(fmt.Sprintf("Failed to create libp2p random host: %v", err))
	}

	return host
}

func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		panic(fmt.Sprintf("Failed to make multiaddr: %v", err))
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		panic(fmt.Sprintf("Failed to get IPFS pid: %v", err))
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode peer ID: %v", err))
	}

	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", pid))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

func streamHandler(stream network.Stream) {
	defer stream.Close()

	buf := bufio.NewReader(stream)
	req, err := http.ReadRequest(buf)
	if err != nil {
		stream.Reset()
		return
	}
	defer req.Body.Close()

	client := &http.Client{}
	req.URL.Scheme = "http"
	hp := strings.Split(req.Host, ":")
	if len(hp) > 1 && hp[1] == "443" {
		req.URL.Scheme = "https"
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	req.URL.Host = req.Host

	outreq := new(http.Request)
	*outreq = *req

	fmt.Printf("Making request to %s \n", req.URL)

	resp, err := client.Do(outreq)
	if err != nil {
		stream.Reset()
		return
	}

	resp.Write(stream)
}

type ProxyService struct {
	host      host.Host
	dest      peer.ID
	proxyAddr ma.Multiaddr
}

func NewProxyService(h host.Host, proxyAddr ma.Multiaddr, dest peer.ID) *ProxyService {
	h.SetStreamHandler(Protocol, streamHandler)

	fmt.Println("Proxy service is ready")
	fmt.Println("libp2p-peer addresses: ")
	for _, a := range h.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", a, h.ID())
	}

	return &ProxyService{
		host:      h,
		dest:      dest,
		proxyAddr: proxyAddr,
	}
}

func (p *ProxyService) Serve() {
	_, serveArgs, _ := manet.DialArgs(p.proxyAddr)
	fmt.Println("Proxy listening on ", serveArgs)
	if p.dest != "" {
		http.ListenAndServe(serveArgs, p)
	}
}

func (p *ProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("proxying request for %s to peer %s \n", r.URL, p.dest)
	stream, err := p.host.NewStream(context.Background(), p.dest, Protocol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()
	fmt.Println("NewStream finished.")

	err = r.Write(stream)
	if err != nil {
		stream.Reset()
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	fmt.Println("Write request finished.")

	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, r)
	if err != nil {
		stream.Reset()
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	fmt.Println("Read response finished.")

	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
	resp.Body.Close()
}

const help = `
This example creates a simple HTTP Proxy using two libp2p peers. The first peer
provides an HTTP server locally which tunnels the HTTP requests with libp2p
to a remote peer. The remote peer performs the requests and 
send the sends the response back.

Usage: Start remote peer first with:   ./proxy
       Then start the local peer with: ./proxy -d <remote-peer-multiaddress>

Then you can do something like: curl -x "localhost:9900" "http://ipfs.io".
This proxies sends the request through the local peer, which proxies it to
the remote peer, which makes it and sends the response back.
`

func main() {
	flag.Usage = func() {
		fmt.Println(help)
		flag.PrintDefaults()
	}

	destPeer := flag.String("d", "", "destination peer address")
	port := flag.Int("p", 9900, "proxy port")
	p2pport := flag.Int("l", 12000, "libp2p listen port")
	flag.Parse()

	if *destPeer != "" {
		host := makeRandomHost(*p2pport + 1)
		destPeerID := addAddrToPeerstore(host, *destPeer)
		proxyAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *port))
		if err != nil {
			panic(err)
		}
		proxy := NewProxyService(host, proxyAddr, destPeerID)
		proxy.Serve()

	} else {
		host := makeRandomHost(*p2pport)

		_ = NewProxyService(host, nil, "")
		<-make(chan struct{})
	}
}
