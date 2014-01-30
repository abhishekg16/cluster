#ClUSTER#
Cluster is provides an interface and associated library by which user can communicate between a group of node. By using cluser API the client of the cluster can send message to any other node in cluster and vice-versa. 

Features: 
1. Allows BroadCasting
2. Peer to Peer Communication
3. Uses the ZeroMQ as underlying communication library
4. Tested for different kind of senarios

Limitations
1. Does not support  Multicast
2. Does not guaranteed message delivery

#PeerCatalog#
The cluster package also contains a sub package call peerCatalog. PeerCatalog package provide the metastore service for the cluster. The cluster uses the peerCatalog to store the information of all the peer instances and using the course of operation it can store and retrive the peer's information efficiently 

##Usages##





