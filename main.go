package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ChainSafe/gossamer/p2p"
	"github.com/ChainSafe/gossamer/common"
	"github.com/ChainSafe/gossamer/common/optional"
	ps "github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

func startNewService(cfg *p2p.Config) *p2p.Service {
	node, err := p2p.NewService(cfg, nil)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	e := node.Start()
	err = <-e
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	return node
}

func main() {
	testServiceConfigA := &p2p.Config{
		NoBootstrap: true,
		Port:        7004,
	}

	sa, err := p2p.NewService(testServiceConfigA, nil)
	if err != nil {
		fmt.Printf("NewService error: %s", err)
		os.Exit(1)
	}

	defer sa.Stop()

	e := sa.Start()
	err = <-e
	if err != nil {
		fmt.Printf("Start error: %s", err)
		os.Exit(1)
	}

	testServiceConfigB := &p2p.Config{
		NoBootstrap: true,
		Port:        7005,
	}

	msgChan := make(chan p2p.Message)
	sb, err := p2p.NewService(testServiceConfigB, msgChan)
	if err != nil {
		fmt.Print("NewService error:", err)
		os.Exit(1)
	}

	defer sb.Stop()

	sb.Host().Peerstore().AddAddrs(sa.Host().ID(), sa.Host().Addrs(), ps.PermanentAddrTTL)
	addr, err := ma.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", sa.Host().Addrs()[0].String(), sa.Host().ID()))
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	err = sb.Host().Connect(sb.Ctx(), *addrInfo)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	e = sb.Start()
	err = <-e
	if err != nil {
		fmt.Printf("Start error: %s", err)
		os.Exit(1)
	}

	p, err := sa.DHT().FindPeer(sa.Ctx(), sb.Host().ID())
	if err != nil {
		fmt.Printf("could not find peer: %s", err)
		os.Exit(1)
	}

	endBlock, err := common.HexToHash("0xfd19d9ebac759c993fd2e05a1cff9e757d8741c2704c8682c15b5503496b6aa1")
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	bm := &p2p.BlockRequestMessage{
		Id:            7,
		RequestedData: 1,
		StartingBlock: []byte{1, 1},
		EndBlockHash:  optional.NewHash(true, endBlock),
		Direction:     1,
		Max:           optional.NewUint32(true, 1),
	}

	encMsg, err := bm.Encode()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	err = sa.Send(p, encMsg)
	if err != nil {
		fmt.Printf("Send error: %s", err)
	}

	select {
	case <-msgChan:
	case <-time.After(5 * time.Second):
		fmt.Printf("Did not receive message from %s", sa.FullAddrs()[0])
		os.Exit(1)
	}
}