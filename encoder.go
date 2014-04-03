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
}

// TODO : Increase the encoding facility 
type EncodingFacility struct{
}

// TODO : Make is singloton
func GetEncoder() ( *EncodingFacility, bool ) {
	var ecf EncodingFacility
	return &ecf , true
}

func (e *EncodingFacility)  Encode (env *Envelope) ([]byte,error){
	
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
	buf_dec := new(bytes.Buffer)
	dec := gob.NewDecoder(buf_dec)
	buf_dec.Write(data)
	var env Envelope
	err := dec.Decode(&env)
	if err != nil {
		log.Println(err)
	}
	//log.Printf("Server: Decoded Data....\n",env)
	return &env,nil
}
