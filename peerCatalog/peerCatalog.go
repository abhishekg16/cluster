package PeerCatalog

// Each peer stuct conatains all the information about a peer node

import (
		"fmt"
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

const (
	INFO = iota
	WARNING = iota
	FINE = iota
	FINEST = iota
)

var LOG int

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
	logger * log.Logger
	serverId int
	shutdownSoc *zmq.Socket // only used to send message to itself 
}

// Send logger if do not want logging
func AllocateNewCatalog(l * log.Logger , logLevel int , serverId int) *PeerCatalog {
	t_mapOfPeers := make(map[int]peer)
	catalog := PeerCatalog{mapOfPeers: t_mapOfPeers}
	catalog.logger = l
	catalog.serverId = serverId
	LOG = logLevel
	if ( LOG >= INFO  ) {
		l.Printf("PeerCalalog :%v, Allocated new peer catalog for server\n", serverId)
	}
	return &catalog
}




// TODO: Add the return error
func (p *PeerCatalog) SetPeers(allPeers *map[int]([]string)) {
	//fmt.Println(allPeers)
	for tpid, taddr := range *allPeers {
		if len(taddr) != 2 {
			if ( LOG >= INFO  ) {
				p.logger.Printf("PeerCalalog %v: peer %v does not provide two socket address\n", p.serverId,tpid)
			}
		}
		newPeer := peer{tpid, taddr[0],taddr[1], nil,nil}
		p.mapOfPeers[tpid] = newPeer
	}
	if ( LOG >= INFO  ) {
		p.logger.Printf("PeerCalalog %v : All peers are set in peer catalog\n",p.serverId)
		//p.PrintCatalog()
	}
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
// return false if could not connect
func (p *PeerCatalog) Connect(pid int) (bool, error) {
	if !p.isExist(pid) {
		if ( LOG >= INFO ){ 
			p.logger.Println("Pid does not exist")
		}
		return false, fmt.Errorf("Pid does not exist\n")
	}
	// make shut down socket 
	
	// Connect to Msg Socket
	socket, err := zmq.NewSocket(zmq.PUSH)
	if err != nil {
		p.logger.Printf("Error: %v\n", err)
		return false,err
	}
	addr, ok := p.GetMsgAddr(pid)
	if ok == false {
		return ok, nil 
	}
	err = socket.Connect(PROTOCOL + addr)
	if err != nil {
		if ( LOG >= INFO ){
			p.logger.Printf("Error Generated : %v , Argument Sent were  : %v", err, PROTOCOL + addr)
		}
		return false , err
	}
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return false, fmt.Errorf("Pid does not exist\n")
	}
	if ( p.serverId != pid ) {
		peer.msgSoc = socket
	}
	
	
	// Connect to Control Socket
	socket, err = zmq.NewSocket(zmq.PUSH)
	addr, ok = p.GetCtrlAddr(pid)
	if ok == false {
		return ok , err
	}
	err = socket.Connect(PROTOCOL + addr)
	if err != nil {
		if  p.logger != nil {
			p.logger.Printf("PeerCalalog %v : Error %v , Argument Sent were : %v ", p.serverId ,err, PROTOCOL + addr)
		}
		return false, err
	}
	if ( p.serverId != pid ) {
		peer.ctrlSoc = socket
	} else {
		p.shutdownSoc = socket
	}
	
	p.mapOfPeers[pid] = *peer
	if ( LOG >= INFO  ) {
		p.logger.Printf("Connecting to %v : %v", pid, addr)
	}
	// Connect to ctrl Socket 
	return true, nil
}

// This method will take the socket and
func (p *PeerCatalog) Bind(pid int) ( bool , error ){
	if !p.isExist(pid) {
		return false , fmt.Errorf("Pid does not exist\n")
	}
	socket, err := zmq.NewSocket(zmq.PULL)
	addr, ok := p.GetMsgAddr(pid)
	if ok == false {
		return false, err 
	}
	socket.Bind(PROTOCOL + addr)
	peer, ok := p.getPeerData(pid)
	if ok == false {
		return false , fmt.Errorf("Pid does not exist\n")
	}
	peer.msgSoc = socket
	if ( LOG >= INFO  ) {
		p.logger.Printf("Binding Server %d MSG POST To Listen on %v", pid, addr)
	}
	
	socket, err = zmq.NewSocket(zmq.PULL)
	addr, ok = p.GetCtrlAddr(pid)
	if ok == false {
		return false , err
	}
	socket.Bind(PROTOCOL + addr)
	peer.ctrlSoc = socket
	
	//Fix this later
	p.mapOfPeers[pid] = *peer
	if ( LOG >= INFO  ) {
		p.logger.Printf("Binding Server %v CTRL POrt To Listen on %v", pid, addr)
	}
	return true , nil
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
		p.logger.Printf("Pid : %d doen not exist", pid)
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
 
func (p *PeerCatalog) GetShutdownSocket() *zmq.Socket {
	return p.shutdownSoc
}


func (p* PeerCatalog) CloseCatalog() {
		if ( LOG >= INFO ){
				p.logger.Printf("Closing Sockets\n")
		}
		for pid , peer := range p.mapOfPeers {
			if ( LOG >= FINE ){
				p.logger.Printf("Closing Socket to pid %v", pid)
			}
			if peer.ctrlSoc != nil { 
				peer.ctrlSoc.Close()
			} else if (	peer.msgSoc != nil ) {
				peer.msgSoc.Close()
			}
		}
}

