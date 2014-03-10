package main

import (
        cluster "github.com/abhishekg16/cluster"
		//"log"
        "fmt"
        "encoding/gob"
)

type VoteRequestToken struct {
	Term int64 	  
	CandidateId int
	//	LastLogIndex int64
	//	LstLogTerm int64
}

func main() {
	//gob.Register(Message{})	
	gob.Register(VoteRequestToken{})
	
	s1,err := cluster.New(0,"../Config.json")
	if err != nil {
               fmt.Println("cound not initialize server")
               return
    }
    
    s2, err := cluster.New(1,"../Config.json")
    if err != nil {
               fmt.Println("cound not initialize server")
               return
    }
    
 	v1 := VoteRequestToken{1,1}    
    m1  := cluster.Message{ MsgCode: 0, Msg: v1 }
    env := cluster.Envelope {Pid: 1, MsgType : cluster.CTRL, Msg : m1 }
    
    s1.Outbox() <- &env
    env2 := <-s2.Inbox()
    
    fmt.Println(env2)
    
    
    
      
	
	
	/*

        s, err := cluster.New (0,"../Config.json")
        if err != nil {
                fmt.Println("cound not initialize server")
                return
        }
        for i := 1 ; i < 5 ; i++ {
        m := cluster.Message {cluster.CTRL,"HI"}
        env := cluster.Envelope{Pid:0,MsgType:cluster.CTRL,Msg:m}
        data, err  := s.S_encodingFacility.Encode(&env)
        fmt.Println("Send ", env)
        if err != nil  {
                log.Println ("Encoding Error")
                log.Println(err)
                return
        }

        d_msg, err := s.S_encodingFacility.Decode(data)
        if err != nil {
                log. Println("Decoding Error")
                log.Println(err)
        }
        log.Println(" RECEIVED",d_msg)
        }
        */
}

