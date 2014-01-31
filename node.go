package cluster

import (
	//	"fmt"
	"encoding/json"
	"io"
	"os"
	"strconv"
	//	"flag"
	catalog "github.com/abhishekg16/cluster/peerCatalog"
	zmq "github.com/pebbe/zmq4"
	"log"
	"time"
	//	"bufio"
)

const (
	PROTOCOL = "tcp://"
)


// Server stucture
type server struct {
	// Local Pid
	pid int

	// Maping of the other servers Pid to their socket address
	pCatalog *catalog.PeerCatalog

	// a channel which puts the message to the out channel
	out chan *Envelope

	// a channel which gets the message from out channel
	in chan *Envelope

	// Each Server will have an unique context
	//context zmq.Context

	// Sending Socket
	out_socket *zmq.Socket

	// Recieving Socket
	in_socket *zmq.Socket

	// Unique massage id
	msgId int64

	//
	recieve chan *Envelope
}

// getMsgId method return the unique global message id
// At present implementation allows only 10000 unique message at a moment for server
// TODO: Replace multiply operation with the Bit-Shift

func (s *server) getMsgId() int64 {
	s.msgId++
	if s.msgId > 10000 {
		s.msgId = 0
	}
	return int64(s.pid*100*1000) + s.msgId
}

func (s *server) Pid() int {
	return s.pid
}

func (s *server) Peers() []int {
	return s.pCatalog.GetPeerList(s.pid)
}


// This method parse the json file returns a map from peer id to socket address
func parse(ownId int, path string) (map[int]string, error) {
	//log.Println("Parsing The Configuration File")
	addr := make(map[int]string)
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Parsing Failed %q", err)
		return nil, err
	}
	dec := json.NewDecoder(file)
	for {
		var v map[string]string
		if err := dec.Decode(&v); err == io.EOF || len(v) == 0 {
			//log.Println("Parsing Done !!!")
			file.Close()
			return addr, nil
		} else if err != nil {
			file.Close()
			return nil, err
		}

		spid := v["ID"]
		pid, _ := strconv.Atoi(spid)
		ip := v["IP"]
		port := v["Port"]
		addr[pid] = ip + ":" + port
	}
}

func (s *server) Outbox() chan *Envelope {
	return s.out
}

func (s *server) Inbox() chan *Envelope {
	return s.in
}

// This method accept a pid and return the true, socket address associated with that pid
// In case the pid not present in peer list it will return the false, nil

func (s *server) GetPeerAddress(pid int) (string, bool) {
	addr, ok := s.pCatalog.GetAddr(pid)
	return addr, ok
}

// connectToallPeers method setup the connection with all other peers
func (s *server) connectToAllPeers() {
	//log.Println("Start Connecting to all Peers")
	listOfPeers := s.Peers()
	for _, pid := range listOfPeers {
		if pid != s.pid {
			_ = s.pCatalog.Connect(pid)
		}
	}
	//log.Println("Successfully Connected to peers: Connection depends on availability of peer")
}

// This method iniitlaizes the server and stablish the connection with other instances.
func (s *server) initialize() error {

	//log.Println("Initializing Server")
	s.msgId = 0

	// allocate all channels
	s.in = make(chan *Envelope, 100)
	s.out = make(chan *Envelope, 100)
	s.recieve = make(chan *Envelope, 100)

	// Initiate connection with all other peers
	s.connectToAllPeers()

	// start Binding
	s.in_socket = s.pCatalog.Bind(s.pid)

	log.Println("Initialization is Done")
	return nil
}


// New method take the pid of a new server and path to the Config.json file 
// Config.json file contains the enrty of all other peers in the system
// This method return a server object

func New(pid int, path string) (server, error) {
	var s server
	s.pid = pid

	m, err := parse(pid, path)
	if err != nil {
		log.Println(err)
		return s, err
	}

	// Allocate a peer catalog for server
	s.pCatalog = catalog.AllocateNewCatalog()
	s.pCatalog.SetPeers(&m)

	err = (&s).initialize()

	if err != nil {
		return s, err
	}

	if err == nil {
		go s.startClientOutBox()
		go s.startClientInBox()
		go s.startServerInBox()
	}
	return s, nil
}

// This method will listen Listing the server socket

func (s *server) startServerInBox() {
	for {

		env, err := s.in_socket.RecvBytes(0)
		//log.Println("Recieved on InSocket on Peer ", s.pid)
		if err != nil {
			log.Println("Error in Reieving Bytes on the Insocket")
			log.Println(err)
		}
		//os.Stdout.Write(env)
		if err == nil {
			var msg Envelope
			err := json.Unmarshal(env, &msg)
			//log.Println("Unmarshaling....")
			if err != nil {
				log.Println("error:", err)
				continue
			}
			s.recieve <- &msg
		}
		// time.After(20 * time.Second):
		//return

	}
}

// This method take an envelope for reply
// It checks the pid set in the envelope and
// forward the message to the peer.
// Before sending the messge to the other peer it set the Envelope.Pid
// servers pid, In case the Pid is not present in the severs peer list the message is sighlently dropped
// This method is not used in current implementation but might be useful in future

func (s *server) ReplyMessage(env Envelope) {
	log.Println("Sending Relply...")
	_, ok := s.GetPeerAddress(env.Pid)
	if ok == false {
		return
	}
	env.Pid = s.pid
	env.Msg = "Ack"
	msg, err := json.Marshal(env)
	if err != nil {
		log.Println(err)
		return
	}
	s.in_socket.SendBytes(msg, 0)
}

// At present the sending sevice only support the point to point and broadcast services
func (s *server) SendMessage(env *Envelope) {
	//log.Println("Sending Message")
	length := len(s.Peers())
	dest := make([]int, length)
	//log.Println(*env)
	if env.Pid == -1 {
		//log.Println("BroadCasting...")
		dest = s.Peers()
	} else {
		dest = dest[0:1]
		dest[0] = env.Pid
	}
	env.Pid = s.pid
	env.MsgId = s.getMsgId()
	for _, pid := range dest {
		msg, err := json.Marshal(*env)
		if err != nil {
			log.Println(err)
			continue
		}
		socket := s.pCatalog.GetSocket(pid)
		if socket == nil {
			log.Println("Socket in not instantiated for pid = %d", pid)
			continue
		}
		socket.SendBytes(msg, 0)
	}
}

// This method will transfer the incoming message to client
func (s *server) startClientInBox() {
	for {
		select {
		case env := <-s.recieve:
			//log.Println("Forwading from server to client %q on peer %d", env, s.pid)
			s.in <- env
		case <-time.After(20 * time.Second):
			return
		}
	}
}


// This Method will forward the client request to server for send
func (s *server) startClientOutBox() {
	for {
		select {
		case env := <-s.out:
			//log.Println("Recieved from client of peer %d to send to Peers %d ", s.pid, env.Pid)
			s.SendMessage(env)
		case <-time.After(200 * time.Second):
			return
		}
	}
}

