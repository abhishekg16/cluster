package cluster

import (
	"encoding/gob"
	"bytes"
	"log"
	//"sync"
	"fmt"
	"os"
	//cluster	"github.com/abhishekg16/cluster"

)
type VoteRequestToken struct {
	Term int64 	  
	CandidateId int
	//	LastLogIndex int64
	//	LstLogTerm int64
}

// TODO : Increase the encoding facility 
type EncodingFacility struct{
	//buf_enc *bytes.Buffer
	//buf_dec *bytes.Buffer
	//elock sync.Mutex // Mutex lock to provide synchronization
	//dlock sync.Mutex 
	//enc *gob.Encoder
	//dec *gob.Decoder
}

// TODO : Make is singloton
func GetEncoder() ( *EncodingFacility, bool ) {
	var ecf EncodingFacility
	//ecf.buf_enc = new(bytes.Buffer)
	//ecf.buf_dec = new(bytes.Buffer)
	//ecf.enc = gob.NewEncoder(ecf.buf_enc)
	//ecf.dec = gob.NewDecoder(ecf.buf_dec)
	
	return &ecf , true
}

func (e *EncodingFacility)  Encode (env *Envelope) ([]byte,error){
	//e.elock.Lock()
	//defer e.elock.Unlock()
	//e.buf_enc.Truncate(0)
	buf_enc := new(bytes.Buffer)
	enc := gob.NewEncoder(buf_enc)
	enc.Encode(env)
	tResult := buf_enc.Bytes()
	result := make([]byte, len(tResult))
	l := copy(result, tResult)
	var err error
	err = nil
	if l != len(tResult) {
		log.Println ("Encodoing Failed")
		err= fmt.Errorf("Encoding Failed")
		os.Exit(0)
	}
	return result, err
	 
}

func (e *EncodingFacility)  Decode (data []byte) (* Envelope,error){
	//log.Printf("Server : Started Decoding....\n")
	//e.dlock.Lock()
	//defer e.dlock.Unlock()
	buf_dec := new(bytes.Buffer)
	dec := gob.NewDecoder(buf_dec)
	//e.buf_dec.Truncate(0)
	buf_dec.Write(data)
	var env Envelope
	err := dec.Decode(&env)
	if err != nil {
		log.Println(err)
	}
	//log.Printf("Server: Decoded Data....\n",env)
	return &env,nil
}
