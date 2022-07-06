package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tor "berty.tech/go-libp2p-tor-transport"
	config "berty.tech/go-libp2p-tor-transport/config"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli"
)

var log = logging.Logger("net")

var (
	flagNoTor = cli.BoolFlag{
		Name:  "no-tor",
		Usage: "run a normal libp2p node without tor",
	}
	flagBootnodes = cli.StringFlag{
		Name:  "bootnodes",
		Usage: "comma-separated list of bootnodes",
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "tor-libp2p"
	app.Usage = "run a libp2p tor node"
	app.Action = runNode
	app.Flags = []cli.Flag{
		flagNoTor,
		flagBootnodes,
	}

	_ = logging.SetLogLevel("net", "debug")

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func runNode(c *cli.Context) (err error) {
	var bootnodes []peer.AddrInfo
	if c.String("bootnodes") != "" {
		bootnodeStrs := strings.Split(c.String("bootnodes"), ",")
		bootnodes, err = stringsToAddrInfos(bootnodeStrs)
		if err != nil {
			return err
		}
	}

	port, err := rand.Int(rand.Reader, big.NewInt(1<<16))
	if err != nil {
		return err
	}

	h, err := newHost(c.Bool("no-tor"), bootnodes, uint16(port.Int64()))
	if err != nil {
		return err
	}

	fmt.Println(h.Addresses())

	err = h.bootstrap()
	if err != nil {
		return err
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	select {
	case <-sigc:
		fmt.Println("signal interrupt, shutting down...")
	}
	return nil
}

type thost struct {
	ctx       context.Context
	h         host.Host
	bootnodes []peer.AddrInfo
}

func newHost(noTor bool, bootnodes []peer.AddrInfo, port uint16) (*thost, error) {
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(addr),
	}

	if noTor {
		h, err := libp2p.New(opts...)
		if err != nil {
			return nil, err
		}

		return &thost{
			ctx:       context.Background(),
			h:         h,
			bootnodes: bootnodes,
		}, nil
	}

	builder, err := tor.NewBuilder(
		config.EnableEmbeded,
	)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(libp2p.Transport(builder))
	if err != nil {
		return nil, err
	}

	return &thost{
		ctx:       context.Background(),
		h:         h,
		bootnodes: bootnodes,
	}, nil
}

func (h *thost) bootstrap() error {
	failed := 0
	for _, addrInfo := range h.bootnodes {
		log.Infof("bootstrapping to peer: peer=%s", addrInfo.ID)
		err := h.h.Connect(h.ctx, addrInfo)
		if err != nil {
			log.Infof("failed to bootstrap to peer: err=%s", err)
			failed++
		}
	}

	if failed == len(h.bootnodes) && len(h.bootnodes) != 0 {
		return fmt.Errorf("failed to bootstrap")
	}

	return nil
}

// multiaddrs returns the multiaddresses of the host
func (h *thost) multiaddrs() (multiaddrs []ma.Multiaddr) {
	addrs := h.h.Addrs()
	for _, addr := range addrs {
		multiaddr, err := ma.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", addr, h.h.ID()))
		if err != nil {
			continue
		}
		multiaddrs = append(multiaddrs, multiaddr)
	}
	return multiaddrs
}

func (h *thost) Addresses() []string {
	var addrs []string
	for _, ma := range h.multiaddrs() {
		addrs = append(addrs, ma.String())
	}
	return addrs
}

// stringToAddrInfo converts a single string peer id to AddrInfo
func stringToAddrInfo(s string) (peer.AddrInfo, error) {
	maddr, err := ma.NewMultiaddr(s)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	p, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	return *p, err
}

// stringsToAddrInfos converts a string of peer ids to AddrInfo
func stringsToAddrInfos(peers []string) ([]peer.AddrInfo, error) {
	pinfos := make([]peer.AddrInfo, len(peers))
	for i, p := range peers {
		p, err := stringToAddrInfo(p)
		if err != nil {
			return nil, err
		}
		pinfos[i] = p
	}
	return pinfos, nil
}
