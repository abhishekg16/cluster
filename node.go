package cluster

import (
		"fmt"
	"encoding/json"
	"io"
	"os"
	"strconv"
	//	"flag"
	catalog "github.com/abhishekg16/cluster/peerCatalog"
	zmq "github.com/pebbe/zmq4"
	"log"
//	"time"
	
	//	"bufio"
	
)

const (
	PROTOCOL = "tcp://"
)


// TODO : Make server singleton
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

	// Sending msg Socket
	out_msg_socket *zmq.Socket
	
	// Sending control Socket
	out_ctrl_socket *zmq.Socket

	// Recieving Msg Socket
	in_msg_socket *zmq.Socket
	
	// Recieving Ctrl Soket
	in_ctrl_socket * zmq.Socket

	// Unique massage id
	msgId int64

	//
	recieve chan *Envelope
	
	// shutdown channel
	closeClientInbox chan bool
	
	closeClientOutbox chan bool
	
	S_encodingFacility *EncodingFacility
	
	logger *log.Logger
	
	isShutdown bool // isShutdown Requested 
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

func (s* server) SetLogger(newlogger *log.Logger){
	
}
// This method parse the json file returns a map from peer id to socket address
func parse(ownId int, path string) (map[int]([]string), error) {
	log.Println("Server %v :Parsing The Configuration File\n",ownId)
	addr := make(map[int]([]string))
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Server %v :Parsing Failed %v", ownId , err)
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
		port1 := v["Port1"]
		tSocAddr := make([]string,0)
		tSocAddr = append(tSocAddr, ip + ":" + port1)
		port2 :=v["Port2"]
		tSocAddr = append(tSocAddr, ip + ":" + port2)
		addr[pid]= tSocAddr
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

func (s *server) GetPeerMsgAddress(pid int) (string, bool) {
	addr, ok := s.pCatalog.GetMsgAddr(pid)
	return addr, ok
}

func (s *server) GetPeerCtrlAddress(pid int) (string, bool) {
	addr, ok := s.pCatalog.GetCtrlAddr(pid)
	return addr, ok
}

// connectToallPeers method setup the connection with all other peers
func (s *server) connectToAllPeers() {
	log.Printf("Server %v :Start Connecting to all Peers",s.pid)
	listOfPeers := s.Peers()
	for _, pid := range listOfPeers {
		if pid != s.pid {
			_ = s.pCatalog.Connect(pid)
		}
	}
	log.Printf("Server %v :Successfully Connected to peers: Connection depends on availability of peer", s.pid)
}

func (s *server) Shutdown() bool {
	log.Printf("Server %v : Shutdown is called at pid %q \n",s.pid)
	s.isShutdown = true
	s.closeClientInbox<-true
	s.closeClientOutbox<-true
	env := Envelope{Pid: s.pid, MsgType : CTRL, PeerList: nil , Msg : Message{}}
	s.SendMessage(&env)
	return true
}

// This method iniitlaizes the server and stablish the connection with other instances.
func (s *server) initialize() error {

	log.Printf("Server %v :Initializing Server\n",s.pid)
	s.msgId = 0

	// allocate all channels
	s.in = make(chan *Envelope, 100)
	s.out = make(chan *Envelope, 100)
	s.recieve = make(chan *Envelope, 100)
	

	// Initiate connection with all other peers
	s.connectToAllPeers()

	// start Binding
	ok := s.pCatalog.Bind(s.pid)
	if ok == false {
		log.Printf("Server %v :Binding of Socket address Failed\n", s.pid)
		return fmt.Errorf("Server %v :Binding Failed \n", s.pid)
	}
	
	s.in_msg_socket = s.pCatalog.GetMsgSocket(s.pid)
	s.in_ctrl_socket = s.pCatalog.GetCtrlSocket(s.pid)
	
	s.closeClientInbox = make(chan bool,1)
	s.closeClientOutbox = make(chan bool,1)
	
	s.S_encodingFacility, _ = GetEncoder() 
	/*
	if ok != true {
		log.Println("Cound not get Encoder")
		return fmt.Errorf("Could not ")
	}
	*/
	
	log.Printf("Server %v :Initialization is Done\n",s.pid)
	return nil
}



// New method take the pid of a new server and path to the Config.json file 
// Config.json file contains the enrty of all other peers in the system
// This method return a server object

func New(pid int, path string) (* server, error) {
	var s server
	s.pid = pid

	m, err := parse(pid, path)
	if err != nil {
		log.Println(err)
		return &s, err
	}

	// Allocate a peer catalog for server
	s.pCatalog = catalog.AllocateNewCatalog()
	s.pCatalog.SetPeers(&m)

	err = (&s).initialize()

	if err != nil {
		return &s, err
	}

	if err == nil {
		go s.startClientOutBox()
		go s.startClientInBox()
		go s.startServerInBox()
	}
	return &s, nil
}

// This method will listen Listing the server sockets
// The server will listen on all the input ports 

func (s *server) startServerInBox() {
		poller := zmq.NewPoller()
		poller.Add(s.in_msg_socket, zmq.POLLIN )
		poller.Add(s.in_ctrl_socket, zmq.POLLIN )
	
		for {
			 sockets ,_ := poller.Poll(-1)
		     for _ ,  socket := range sockets {
	         	switch soc := socket.Socket; soc {
	            	case s.in_ctrl_socket :
	            		if s.isShutdown == true {
	            			return
	            		} 
		                env, err := s.in_ctrl_socket.RecvBytes(0)
						log.Printf("Server %v: Recieved In Control Socket \n", s.pid)
						if err != nil {
							log.Printf("Server %v: Error in Reieving Bytes on the InControl socket\n", s.pid)
							log.Println(err)
						}
						//log.Printf("Server %v: Message Before Unmarshell\n",s.pid)
						//os.Stdout.Write(env)
						if err == nil {
							log.Printf("Server %v: Unmarshaling....\n",s.pid)
							msg, err := s.S_encodingFacility.Decode(env)
							if err != nil {
								log.Printf("Server %v, error: %v\n", s.pid, err)
								continue
							}
							log.Printf("Server %v: Recieved Message : %+v \n", s.pid,msg)
							if msg.MsgType == SHUTDOWN {
								return 
							}
							s.recieve <- msg
						}
					case s.in_msg_socket :
						if s.isShutdown == true {
	            			return
	            		} 
						env, err := s.in_msg_socket.RecvBytes(0)
						log.Printf("Server %v :Recieved on In Message Socket on Peer %d \n", s.pid)
						if err != nil {
							log.Printf("Server %v :Error in Reieving Bytes on the In Control socket\n")
							log.Println(err)
						}
						//log.Println("Message Before Unmarshell")
						//os.Stdout.Write(env)
						if err == nil {
							//var msg Envelope
							//err := json.Unmarshal(env, &msg)
							msg, err := s.S_encodingFacility.Decode(env)
							log.Printf("Server %v Unmarshaling....\n",s.pid)
							if err != nil {
								log.Println("error:", err)
								continue
							}
							s.recieve <- msg
						}
				}	
		}
	}
}	

// This method take an envelope for reply
// It checks the pid set in the envelope and
// forward the message to the peer.
// Before sending the messge to the other peer it set the Envelope.Pid
// servers pid, In case the Pid is not present in the severs peer list the message is sighlently dropped
// This method is not used in current implementation but might be useful in future

/*
func (s *server) ReplyMessage(env Envelope) {
	log.Println("Sending Relply...")
	_, ok := s.GetPeerAddress(env.Pid)
	if ok == false {
		return
	}
	env.Pid = s.pid
	//env.Msg = "Ack"
	msg, err := json.Marshal(env)
	if err != nil {
		log.Println(err)
		return
	}
	s.in_socket.SendBytes(msg, 0)
}
*/

// At present the sending sevice only support the point to point and broadcast services
// Depending On the type of message the Message is send either on the control or message socket
func (s *server) SendMessage(env *Envelope) {
	//log.Printf("Server %v: Sending Message", s.pid)
	dest := make([]int, 0)
	log.Println(*env)
	if env.Pid == -1 {
		log.Printf("Server %v: BroadCasting...Message %+v :",s.pid, env )
		dest = s.Peers()
	} else if env.Pid == -2{
		log.Printf("Server %v: MultiCasting...%+v",s.pid , env)
		dest = env.PeerList
		log.Printf("Server %v: Peer List %v",s.pid,dest)
	} else { 
		log.Printf("Server %v: Peer %v sending to Peer %v Message %+v", s.pid, s.pid ,env.Pid, env)
		dest = append( dest, env.Pid)
	}
	env.Pid = s.pid
	env.MsgId = s.getMsgId()
	env.PeerList = nil
	//log.Printf("Server %v: Message Being send  :",s.pid , env )
	log.Printf("Peer %v",dest)
	for _, pid := range dest {
		msg, err := s.S_encodingFacility.Encode(env)
		if err != nil {
			log.Println(err)
			continue
		}
		var socket *zmq.Socket 
		if (env.MsgType == CTRL) {
			socket = s.pCatalog.GetCtrlSocket(pid)
		} else if (env.MsgType == MSG){
			socket = s.pCatalog.GetMsgSocket(pid)
		}
		if socket == nil {
			//log.Printf("Server %v: Socket in not instantiated for pid = %v\n", pid)
			continue
		}
		//os.Stdout.Write(msg)
		//log.Printf("Server %v: Send On socket %v :",s.pid , socket )
		socket.SendBytes(msg, zmq.DONTWAIT)
	}
}

// This method will transfer the incoming message to client
func (s *server) startClientInBox() {
	for {
		select {
		case env := <-s.recieve:
			log.Printf("Server %v: Forwarding to client %+v ",s.pid, env)
			s.in <- env
		case  <- s.closeClientInbox:
			//log.Println("Server %v: Client Inbox closed",s.pid)
			return
		}
	}
}


// startClientOutBox start the client outbox 
func (s *server) startClientOutBox() {
	for {
		select {
		case env := <-s.out:
			//log.Printf("Server %v: Recived on Client Outbox =>  +%v \n", s.pid, env)
			s.SendMessage(env)
		case <- s.closeClientOutbox:
			//log.Printf("Server %v: Client outbox is closed ", s.pid)
			return
		}
	}
}


