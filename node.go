package cluster

import (
	//"fmt"
	"encoding/json"
	"io"
	"os"
	"strconv"
	//	"flag"
	catalog "github.com/abhishekg16/cluster/peerCatalog"
	zmq "github.com/pebbe/zmq4"
	"log"
	//"time"
	//"bufio"
)

const (
	PROTOCOL = "tcp://"
	INPUT_BUFFER_LENGTH  = 1000 // length of the input buffer which keeps the incoming messages
	OUTPUT_BUFFER_LENGTH = 1000
)

var LOG int

// TODO : Make server singleton

// Server stucture
type server struct {
	pid int  // Local Pid
	
	pCatalog *catalog.PeerCatalog  // Maping of the other servers Pid to their socket address
	
	out chan *Envelope // a channel which puts the message to the out channel

	in chan *Envelope // a channel which gets the message from out channel

	out_msg_socket *zmq.Socket // Sending msg Socket
	
	out_ctrl_socket *zmq.Socket // Sending control Socket

	in_msg_socket *zmq.Socket // Recieving Msg Socket
	
	in_ctrl_socket * zmq.Socket  // Recieving Ctrl Soket

	msgId int64 // Unique massage id
	
	
	
	closeServerOutbox chan bool
	
	closeServerInbox chan bool
	
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



//==== Function related to server interface  
func (s *server) Pid() int {
	return s.pid
}

func (s *server) Peers() []int {
	return s.pCatalog.GetPeerList(s.pid)
}

// This method parse the json file returns a map from peer id to socket address
func (s * server) parse(ownId int, path string) (map[int]([]string), error) {
	if (LOG >= HIGH) {
		s.logger.Printf("Server %v :Parsing The Configuration File\n",ownId)
	}
	addr := make(map[int]([]string))
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		if (LOG >= INFO ) {
			s.logger.Fatalf("Server %v :Parsing Failed %v", ownId , err)
		}	
		return nil, err
	}
	dec := json.NewDecoder(file)
	for {
		var v map[string]string
		if err := dec.Decode(&v); err == io.EOF || len(v) == 0 {
			if (LOG >= FINE) {
				s.logger.Println("Parsing Done !!!")
			}
			return addr, nil
		} else if err != nil {
				if (LOG >= INFO) {
					s.logger.Fatalf("Server %v : Decoding Failed!!! \n",ownId, err)
				}
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

// Outbox method returns a channel used by client inorder send the method to the other peers on the cluster 
func (s *server) Outbox() chan *Envelope {
	return s.out
}

// Inbox Method returns a channel used by client inorder receive the incoming channel
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
func (s *server) connectToAllPeers() (bool, error){
	if (LOG >= FINE) { 
		s.logger.Printf("Server %v :Start Connecting to all Peers",s.pid)
	}
	listOfPeers := s.Peers()
	// Need to connect ot self in order to get shut down message 
	listOfPeers = append(listOfPeers,s.pid)
	for _, pid := range listOfPeers {
		//if pid != s.pid {
			ok, err := s.pCatalog.Connect(pid)
			if ok == false {s.closeServerOutbox<-true
				if (LOG >= INFO) { 
					s.logger.Printf("Server %v :Error occuerd", s.pid)
				}			
				return false, err
			}
		//}
	}
	if (LOG >= FINE) { 
		s.logger.Printf("Server %v :Successfully Connected to peers", s.pid)
	}
	return true, nil
}

func (s *server) Shutdown() bool {
	if (LOG >= HIGH) {
		s.logger.Printf("Server %v : Shutdown is called \n",s.pid)
	}
	env := Envelope{Pid: s.pid, MsgType : SHUTDOWN_CLUSTER, PeerList: nil , Msg : Message{}}
	s.SendMessage(&env)
	//s.shutdown
	if (LOG >= HIGH) {
		s.logger.Printf("Server %v : Waiting for Inbox to closed \n",s.pid)
	}
	
	<-s.closeServerInbox
	
	s.closeServerOutbox<-true
	if (LOG >= INFO) {
		s.logger.Printf("Server %v : Shutdown \n",s.pid)
	}
	return true
}

// This method iniitlaizes the server and stablish the connection with other instances.
func (s *server) initialize() error {
	if (LOG >= HIGH) {
		s.logger.Printf("Server %v :Initializing Server\n",s.pid)
	}
	s.msgId = 0
	
	// allocate all channels
	s.in = make(chan *Envelope, INPUT_BUFFER_LENGTH)
	s.out = make(chan *Envelope, OUTPUT_BUFFER_LENGTH)
	
	s.closeServerOutbox = make (chan bool , 1)
	s.closeServerInbox = make (chan bool , 1)
	
	// Initiate connection with all other peers
	s.connectToAllPeers()
	
	// start Binding
	ok, err := s.pCatalog.Bind(s.pid)
	if ok == false {
		s.logger.Printf("Server %v :Binding of Socket address Failed\n", s.pid)
		return err
	}
	
	s.in_msg_socket = s.pCatalog.GetMsgSocket(s.pid)
	s.in_ctrl_socket = s.pCatalog.GetCtrlSocket(s.pid)
	
	
	s.S_encodingFacility, _ = GetEncoder() 
	/*
	if ok != true {
		log.Println("Cound not get Encoder")
		return fmt.Errorf("Could not ")
	}
	*/
	
	//s.logger.Printf("Server %v :Initialization is Done\n",s.pid)
	return nil
}



// New method take the pid of a new server and path to the Config.json file 
// Config.json file contains the enrty of all other peers in the system
// This method return a server object
// logger is true if want you use the file logger 
// else false (print on Stderr)

func New(pid int, path string, logger *log.Logger, logLevel int ) (* server, error) {
	var s server
	s.pid = pid
	
	if logger == nil {
		s.logger = log.New(os.Stderr, "Log: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		s.logger = logger
	}
	
	m, err := s.parse(pid, path)
	if err != nil {
		s.logger.Println(err)
		return &s, err
	}
	
	LOG = logLevel
	
	// Allocate a peer catalog for server
	s.pCatalog = catalog.AllocateNewCatalog(s.logger, LOG, s.pid)
	s.pCatalog.SetPeers(&m)
	//log.Println(m)
	err = (&s).initialize()
	if err != nil {
		s.pCatalog.CloseCatalog()
		return &s, err
	}

	if err == nil {
		go s.serverOutbox()
		go s.serverInbox()
	}
	return &s, nil
}

// This method will listen Listing the server sockets
// The server will listen on all the input ports 

func (s *server) serverInbox() {
		defer s.pCatalog.CloseCatalog()
		poller := zmq.NewPoller()
		poller.Add(s.in_msg_socket, zmq.POLLIN )
		poller.Add(s.in_ctrl_socket, zmq.POLLIN )
	
		for {
			 sockets ,_ := poller.Poll(-1)
		     for _ ,  socket := range sockets {
	         	switch soc := socket.Socket; soc {
	            	case s.in_ctrl_socket : 
		                env, err := s.in_ctrl_socket.RecvBytes(0)
		                if (LOG >= HIGH) {
							s.logger.Printf("Server %v: Recieved In Control Socket \n", s.pid)
						}
						if err != nil {
							if (LOG == INFO) {
								s.logger.Printf("Server %v: Error while recieving\n", s.pid)
								s.logger.Println(err)
							}
						}
						if err == nil {
							if (LOG == FINE) {
								s.logger.Printf("Server %v: Unmarshaling....\n",s.pid)
							}
							msg, err := s.S_encodingFacility.Decode(env)
							if err != nil {
								if (LOG == INFO) {
									s.logger.Printf("Server %v, error: %v\n", s.pid, err)
								}
								continue
							}
							if (LOG >= FINE) {
								s.logger.Printf("Server %v: Recieved Message : %+v \n", s.pid,msg)
							}	
							if msg.MsgType == SHUTDOWN_CLUSTER {
								if (LOG >= HIGH) {
									s.logger.Printf("Sever %v : InboxClosed \n",s.pid)
								}
								s.closeServerInbox<-true
								return 
							}
							
							// TODO : based on the size of the input queue start droping the message
							if len(s.in) < 1000 { 
								s.in <- msg
							} else  {
								if (LOG >= INFO) {
									s.logger.Printf("Server %v: Input Chennel is full server dropping message : %+v \n", s.pid)
								}	
							}
						}
					case s.in_msg_socket :
						env, err := s.in_msg_socket.RecvBytes(0)
						s.logger.Printf("Server %v :Recieved on In Message Socket on Peer %v \n", s.pid)
						if err != nil {
							s.logger.Printf("Server %v :Error in Reieving Bytes on the In Control socket\n")
							s.logger.Println(err)
						}
						
						s.logger.Fatalf("Server %v :Message Received %v\n", s.pid,env)
						s.logger.Fatalf("Server %v :Not Expected to receive on this socket %v \n", s.pid)
						// take code if needed from version 9
				}	
		}
	}
}	




// serverOutbox Read the message dumped on the out channel by client and sent the to peers    
func (s *server) serverOutbox() {
	for {
		select {
		case env := <-s.out:
			if (LOG == HIGH) {
				s.logger.Printf("Server %v: Recived Outbox.. Message:  +%v \n", s.pid, env)
			}
			s.SendMessage(env)
		case <- s.closeServerOutbox:
			if (LOG == HIGH) {
				s.logger.Printf("Server %v: Outbox is closed ", s.pid)
			}
			return
		}
	}
}


// At present the sending sevice only support the point to point and broadcast services
// Depending On the type of message the Message is send either on the control or message socket
func (s *server) SendMessage(env *Envelope) {
	if (LOG >= HIGH) {
		s.logger.Printf("Server %v: Sending Message", s.pid)
	}
	dest := make([]int, 0)
	if env.Pid == -1 {
		if (LOG >= HIGH) {
			s.logger.Printf("Server %v: BroadCasting...Message %+v :",s.pid, env )
		}
		dest = s.Peers()
	} else if env.Pid == -2{
		dest = env.PeerList
		if (LOG >= HIGH) {
			s.logger.Printf("Server %v: Multicasting..PeerList : %+v Message : %+v",s.pid,dest,env)
		}
	} else {
		if (LOG >= HIGH) { 
			s.logger.Printf("Server %v: Peer(%v) to Peer(%v)...Message %+v", s.pid, s.pid ,env.Pid, env)
		}
		dest = append( dest, env.Pid)
	}
	if len(dest) == 0 {
		s.logger.Printf("Server %v: SendMessage failed as the detination List is empty\n",s.pid)
		return 
	}
	env.Pid = s.pid
	env.MsgId = s.getMsgId()
	env.PeerList = nil 	//set PeerList Nil as it not needed any more
	if (LOG == FINE ) {
		s.logger.Printf("Server %v: Message being send  %+v:",s.pid , env )
	}
	
	for _, pid := range dest {
		if (LOG == FINE ) {
			s.logger.Printf("Server %v: Sending to  %+v:", s.pid, pid)
		}
		msg, err := s.S_encodingFacility.Encode(env)
		if err != nil {
			if (LOG >= INFO) {
				s.logger.Printf("Server %v : Encoding failed..Skiped Message\n",s.pid)
				s.logger.Println(err)
			}	
			continue
		}
		var socket *zmq.Socket 
		if (env.MsgType == SHUTDOWN_CLUSTER) {
			socket = s.pCatalog.GetShutdownSocket()
		} else if (env.MsgType == CTRL) {
			socket = s.pCatalog.GetCtrlSocket(pid)
		} else if (env.MsgType == MSG){
			if (LOG == INFO) {
				s.logger.Fatalf("Server %v:  Sending on message Socket, which must not be used", s.pid)
			}
			socket = s.pCatalog.GetMsgSocket(pid)			
		}
		if socket == nil {
			if (LOG == HIGH) {
				s.logger.Printf("Server %v:No Socket avail for pid = %v\n...Skipping Message", pid)
			}
			continue
		}
		if (LOG == FINE ) {
			s.logger.Printf("Server %v:  Just besfore sending to Sending to  %+v:", s.pid, pid)
		}
		_, err = socket.SendBytes(msg, zmq.DONTWAIT)
		if ( err != nil && LOG == FINE ) {
			s.logger.Printf("Server %v:  Sending failed :   %+v:", s.pid, err)
		}
	}
}