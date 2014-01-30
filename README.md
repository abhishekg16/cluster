#CLUSTER#
Cluster is provides an interface and associated library by which user can communicate between a group of node. By using cluser API the client of the cluster can send message to any other node in cluster and vice-versa. 

####Features:#### 
1. Allows BroadCasting
2. Peer to Peer Communication
3. Uses the ZeroMQ as underlying communication library
4. Tested for different kind of senarios

####Limitations####
1. Does not support  Multicast
2. Does not guaranteed message delivery
3. Message will be silently droped in case of destination is not reachable is down
4. The cluster does not support dynamic addition of new servers

#PeerCatalog#
The cluster package also contains a sub package call peerCatalog. PeerCatalog package provide the metastore service for the cluster. The cluster uses the peerCatalog to store the information of all the peer instances and using the course of operation it can store and retrive the peer's information efficiently 

##Usages##
Cluster Provide Following API using which user can get a server instance and can further communicate with the other peer machine in the cluster.

#####Config.json#####
Config.json the configuration file which contains the detail all other peers in the cluster. In order to add more server in the cluster add new entry in this file

####API####
New method take the pid of a new server and path to the Config.json file. It allocates a new server and return it after initailization. User can use this method to create an server instance by which he can access the clsuster

New(pid int, path string) (server, error) 


#####Messages#####
The communication unit is an Envelope which contains the pid of destination, Globally unique messageId and Message. User can create a new message using this
    

type Envelope struct {
	Pid int // Peer id of destination
        MsgId int64 // Globally unique messageId
        Msg interface{} // Actual message 
}


#####Server Interface#####
Server instance Exposes the some method which allows user to communicate with cluster. 
type Server interface {
        Pid() int	// Id of this server
        Peers() []int   // Returns the List of Peers
        Outbox() chan *Envelope // the channel to send the message to othe peers
        Inbox() chan *Envelopw // the channel to receive messages from other peers.
}

###Building###
To build the project you need to install the zmq4 golang binding.







