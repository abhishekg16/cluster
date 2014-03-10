package cluster

import "log"
// Cluster Provide following Interface

const (
	BROADCAST = -1
	MULTICAST = -2
	CTRL = iota
	MSG  = iota
	SHUTDOWN = iota
)

type Message struct {
	MsgCode int
	Msg interface{}
}




type Envelope struct {
	// On the sender side, Pid identifies the receiving peer. If instead, Pid is
	// set to cluster.BROADCAST, the message is sent to all peers. On the receiver side, the
	// Id is always set to the original sender. If the Id is not found, the message is silently dropped
	Pid int

	// An id that globally and uniquely identifies the message, meant for duplicate detection at
	// higher levels. It is opaque to this package.
	MsgId int64
	
	// Message Type Tells about the type of messge : Control or Normal
	MsgType int

	PeerList []int

	// the actual message.
	Msg Message
}

type Server interface {
	// Id of this server
	Pid() int
	

	// array of other servers' ids in the same cluster
	Peers() []int

	// the channel to use to send messages to other peers
	// Note that there are no guarantees of message delivery, and messages
	// are silently dropped
	Outbox() chan *Envelope

	// the channel to receive messages from other peers.
	Inbox() chan *Envelope
	
	SetLogger(*log.Logger)
	
	Shutdown() bool
}
