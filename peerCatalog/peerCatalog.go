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

// peer struct contains the details about each peer 
// peer id , ctrl_addr : port where to sent the control message 
// msg_addr : socket adress where to send the message  

type peer struct {
	pid  int
	ctrlAddr string
	msgAddr string
	ctrlSoc  *zmq.Socket
	msgSoc	*zmq.Socket
	
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


// TODO: Add the return error
func (p *PeerCatalog) SetPeers(allPeers *map[int]([]string)) {
	for tpid, taddr := range *allPeers {
		if len(taddr) != 2 {
			log.Printf("The Peer : %q does not provide two socket address")
			
		}
		newPeer := peer{tpid, taddr[0],taddr[1], nil,nil}
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

func (p *PeerCatalog) GetMsgAddr(pid int) (string, bool) {
	peer, ok := p.mapOfPeers[pid]
	if ok == false {
		return "", ok
	}
	addr := peer.msgAddr
	return addr, true
}

func (p *PeerCatalog) GetCtrlAddr(pid int) (string, bool) {
	peer, ok := p.mapOfPeers[pid]
	if ok == false {
		return "", ok
	}
	addr := peer.ctrlAddr
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
	
	// Connect to Msg Socket
	socket, _ := zmq.NewSocket(zmq.PUSH)
	addr, ok := p.GetMsgAddr(pid)
	if ok == false {
		return ok
	}
	err := socket.Connect(PROTOCOL + addr)
	if err != nil {
		log.Printf("Error Generated : %q", err)
		return false
	}
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return false
	}
	peer.msgSoc = socket
	
	// Connect to Control Socket
	socket, _ = zmq.NewSocket(zmq.PUSH)
	addr, ok = p.GetCtrlAddr(pid)
	if ok == false {
		return ok
	}
	err = socket.Connect(PROTOCOL + addr)
	if err != nil {
		log.Printf("Error : %q", err)
		return false
	}
	peer.ctrlSoc = socket
	
	p.mapOfPeers[pid] = *peer
	log.Printf("Connecting server to %d : %q", pid, addr)
	// Connect to ctrl Socket 
	return true
}

// This method will take the socket and
func (p *PeerCatalog) Bind(pid int) bool {
	if !p.isExist(pid) {
		return false
	}
	socket, _ := zmq.NewSocket(zmq.PULL)
	addr, ok := p.GetMsgAddr(pid)
	if ok == false {
		return false
	}
	socket.Bind(PROTOCOL + addr)
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return false
	}
	peer.msgSoc = socket
	log.Printf("Binding Server %d To Listen on %q", pid, addr)
	
	socket, _ = zmq.NewSocket(zmq.PULL)
	addr, ok = p.GetCtrlAddr(pid)
	if ok == false {
		return false
	}
	socket.Bind(PROTOCOL + addr)
	peer.ctrlSoc = socket
	
	//Fix this later
	p.mapOfPeers[pid] = *peer
	log.Printf("Binding Server %d To Listen on %q", pid, addr)
	return true
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
func (p *PeerCatalog) GetMsgSocket(pid int) *zmq.Socket {
	if !p.isExist(pid) {
		return nil
	}
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return nil
	}
	return peer.msgSoc
}

func (p *PeerCatalog) GetCtrlSocket(pid int) *zmq.Socket {
	if !p.isExist(pid) {
		return nil
	}
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return nil
	}
	return peer.ctrlSoc
}
 

/*
func (P * PeerCatalog) addPeers( pid int, addr string ) {
	// check for duplicate id
	p.mapOfPeers[pid] = addr
}
*/
