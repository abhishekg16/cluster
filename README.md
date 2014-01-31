#CLUSTER#
Cluster provides an interface and associated library by which user can communicate with a group of node. By using cluster API the client of the cluster can send message to any other node in cluster and can receive the message send by them. The cluster uses the asynchronous message passing mechanism for communication. 

####Features 
1. Allows BroadCasting and Peer to Peer Communication
3. Uses the ZeroMQ as underlying communication library.
4. Tested for different kind of scenarios

####Limitations####
1. Does not support Multicast, but that can be easily extended.
2. Does not guaranteed message delivery.
3. Message will be silently dropped in case of destination is not reachable.
4. The cluster does not support dynamic addition of new servers.

#PeerCatalog#
The cluster package also contains a sub package call peerCatalog. PeerCatalog package provide the metastore service for the cluster. The cluster uses the peerCatalog to store the information of all the peer instances and during the course of operation it can store and retrieve the peer's information efficiently. 

##Usages##
Cluster provides rich API using which user can get a reference to server instance and can further communicate with the other peer machine in the cluster.

#####Config.json#####
Config.json the configuration file which contains the detail all other peers in the cluster. In order to add more server in the cluster add new entry in this file. The present implementation does not allow the dynamic configuration changes. 

####API####
New methods take the peerId : pid ( unique id for each server in the cluster ) of a server and path to the Config.json file. It allocates a new server if not already exist and return it after initialization. In case of error the method returns the error message.


func New(pid int, path string) (server, error) 


#####Messages#####
The communication unit is an Envelope struct which contains the peerId of destination, globally unique messageId and actual message. User can create a new message 

env := Envelope{Pid: 0, Msg: "hello there"}   

type Envelope struct {

	Pid int // Peer id of destination

        MsgId int64 // Globally unique messageId

        Msg interface{} // Actual message 

}


#####Server Interface#####
Server instance provides the some method which allows user to communicate with cluster. 

type Server interface {

        Pid() int	// Id of this server
	
        Peers() []int   // Returns the List of Peers
	
        Outbox() chan *Envelope // the channel to send the message to othe peers
	
        Inbox() chan *Envelopw // the channel to receive messages from other peers.
}

###Building###
1. To build the project you need to install the zmq4 golang binding. Binding are present at https://github.com/pebbe/zmq4
2. Get the cluster package from git hub
	go get github.com/abhishekg16/cluster
3. Then enter in cluster directory and run test
	go test

####An Example Client
In following code snippet the client1 sends 10 message from server with pid : 0 to the client which is connected to server1. These are main functions which take the id server id as the command line argument

**"$ go run client1.go -id 0"**

**"$ go run client2.go -id 1"**

###client 1
func main() {

	var f int
	flag.IntVar(&f,"id",-1,"enter pid")
	flag.Parse()
	if f == -1 {
		fmt.Println("Invalid Arguments")
		os.Exit(0)
	} 
	s , err := New(f,"Config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i := 1; i < 10 ; i++ {
		s.Outbox() <- &Envelope{Pid: BROADCAST, Msg: "hello there"}
		time.Sleep(time.Second)
	}	

} 



####client 2

func main() {

	var f int
	flag.IntVar(&f,"id",-1,"enter pid")
	flag.Parse()
	if f == -1 {
		fmt.Println("Invalid Arguments")
		os.Exit(0)
	} 
	s , err := New(f,"Config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	for {	
		select {
	    	case env := <- s.Inbox() :
	        	log.Println("Msg Recived to client %q", env)
	  	 case <- time.After(20 * time.Second): 
	         	println("Waited and waited. Ab thak gaya\n")
	  	}
	 }
	
} 





