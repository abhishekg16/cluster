package PeerCatalog

// Each peer stuct conatains all the information about a peer node

import (
	//	"fmt"
	//	"encoding/json"
	//	"os"
	//	"io"
	//	"strconv"
	//	"flag"
	zmq "github.com/pebbe/zmq4"
	"log"

	//	"time"
	//	"bufio"
)

const (
	PROTOCOL = "tcp://"
)

type peer struct {
	pid  int
	addr string
	soc  *zmq.Socket
}

// Thsi is a meta data store of all the peers
type PeerCatalog struct {
	mapOfPeers map[int]peer
}

func AllocateNewCatalog() *PeerCatalog {
	t_mapOfPeers := make(map[int]peer)
	catalog := PeerCatalog{mapOfPeers: t_mapOfPeers}
	log.Println("Allocated new peer catalog for server")
	return &catalog
}

func (p *PeerCatalog) SetPeers(allPeers *map[int]string) {
	for tpid, taddr := range *allPeers {
		newPeer := peer{tpid, taddr, nil}
		p.mapOfPeers[tpid] = newPeer
	}
	log.Println("All peers are set in peer catalog")
}

func (p *PeerCatalog) GetNOfPeers() int {
	return len(p.mapOfPeers)
}

// checks whether peer exist or not
func (p *PeerCatalog) isExist(pid int) bool {
	_, ok := p.mapOfPeers[pid]
	return ok
}

func (p *PeerCatalog) GetAddr(pid int) (string, bool) {
	peer, ok := p.mapOfPeers[pid]
	if ok == false {
		return "", ok
	}
	addr := peer.addr
	return addr, true
}

func (p *PeerCatalog) GetPeerList(pid int) []int {
	peers := make([]int, 0)
	for key, _ := range p.mapOfPeers {
		if key != pid {
			peers = append(peers, key)
		}
	}
	return peers
}

// At present its only providing the Send function as the requirement spec says that the cluster can be lossy so
// it silently send the packet to the peer and does not wait for reply msg
func (p *PeerCatalog) send(msg []byte) {

}

// This creates a socket and connect the socket to addr associated with peer
func (p *PeerCatalog) Connect(pid int) bool {
	if !p.isExist(pid) {
		log.Println("Pid does not exist")
		return false
	}
	socket, _ := zmq.NewSocket(zmq.DEALER)
	addr, ok := p.GetAddr(pid)
	if ok == false {
		return ok
	}
	err := socket.Connect(PROTOCOL + addr)
	if err != nil {
		log.Printf("Error : %q", err)
		return false
	}
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return false
	}
	peer.soc = socket
	p.mapOfPeers[pid] = *peer
	log.Printf("Connecting server to %d : %q", pid, addr)
	return true
}

// This method will take the socket and
func (p *PeerCatalog) Bind(pid int) *zmq.Socket {
	if !p.isExist(pid) {
		return nil
	}
	socket, _ := zmq.NewSocket(zmq.DEALER)
	addr, ok := p.GetAddr(pid)
	if ok == false {
		return nil
	}
	socket.Bind(PROTOCOL + addr)
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return nil
	}
	peer.soc = socket
	//Fix this later
	p.mapOfPeers[pid] = *peer
	log.Printf("Binding Server %d To Listen on %q", pid, addr)
	return socket
}

func (p *PeerCatalog) getPeerData(pid int) (*peer, bool) {
	peer, ok := p.mapOfPeers[pid]
	if ok == false {
		return nil, ok
	}
	return &peer, true
}

/*
func (p* PeerCatalog)setSocket( pid int, socket * Socket)
{
	if (!isExist(pid)) {
		log.Printf("Pid : %d doen not exist", pid)
		return false
	}
	mapToPeers[pid].soc = socket;
}
*/
func (p *PeerCatalog) GetSocket(pid int) *zmq.Socket {
	if !p.isExist(pid) {
		return nil
	}
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return nil
	}
	return peer.soc
}

/*
func (P * PeerCatalog) addPeers( pid int, addr string ) {
	// check for duplicate id
	p.mapOfPeers[pid] = addr
}
*/
