package cluster

import "testing"

//import "log"
import "time"
import "fmt"

const (
	NOFSERVER  = 4
	NOFMESSAGE = 10
)

func makeDummyServer(num int) ([]server, error) {
	s := make([]server, num)
	var err error
	for i := 0; i < num; i++ {
		s[i], err = New(i, "Config.json")
		if err != nil {
			return nil, err
		}

	}
	return s, nil
}

func sendMessage(s *server, delay int) {
	for i := 0; i < NOFMESSAGE; i++ {
		env := Envelope{BROADCAST, 0, "Hi"}
		s.Outbox() <- &env
		if delay == 1 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

/*
func TestClusterMessageCount(t *testing.T) {
	s, err := makeDummyServer(NOFSERVER)
	if err != nil {
		t.Errorf("Server Instantiation Failed %q", err)
	}
	go sendMessage(&s[0],1)
	go sendMessage(&s[1],1)
	go sendMessage(&s[2],1)
	go sendMessage(&s[3],1)
	count := make([]int, NOFSERVER)

	for {
		select {
			case val := <-s[1].Inbox():
				fmt.Println("case 1 ", val)
				count[1] += 1
			case val := <-s[1].Inbox():
				fmt.Println("case 1 ", val)
				count[1] += 1
			case  val := <-s[2].Inbox():
				fmt.Println("case 2 ", val)
				count[2] += 1
			case  val := <-s[3].Inbox():
				fmt.Println("case 3 ", val)
				count[3] += 1
			case <- time.After(3* time.Second):
				break
		}
	}
	fmt.Println("Reached")

	if ( count[1] != NOFMESSAGE|| count[2] != NOFMESSAGE || count[3] != NOFMESSAGE  ) {
		t.Errorf("Server Instantiation Failed %q", err)
	}
}

*/
func sendMessageTo(s *server, pid int, env *Envelope) {
	s.Outbox() <- env
}

func waitForMessage(s *server) *Envelope {
	select {
	case msg := <-s.Inbox():
		return msg
	case <-time.After(2 * time.Second):
		return nil
	}
}

/*
// This test case check the point to point communication by passing in a message in round Robin Mannaer
func TestClusterMessageRoundRobin(t *testing.T) {
	s, err := makeDummyServer(NOFSERVER)
	if err != nil {
		t.Errorf("Server Instantiation Failed %q", err)
	}
	env := Envelope{1, 0, "RoundRobin"}
	sendMessageTo(&s[0],1,&env)
	msg := waitForMessage(&s[1])

	msg.Pid = 2
	sendMessageTo(&s[1],2,msg)
	msg = waitForMessage(&s[2])

	msg.Pid = 3
	sendMessageTo(&s[2],3,msg)
	msg = waitForMessage(&s[3])

	if ( msg.Msg != "RoundRobin") {
		t.Errorf("Round Robin Test us failed %q", err)
	}
}

*/
// This test case checks whether how does cluster behave in case of down servers
// According to design of the system the if the client machine is down the message would be lost
// Because we are assuming that the Cluster might be lossy

/*
func TestClusterTS3(t *testing.T)  {
	s,_ := makeDummyServer(1)
	env := Envelope{1, 0, "RoundRobin"}
	sendMessageTo(&s[0],1,&env)

	s1, err := New(1,"Config.json")
	if err != nil {
		t.Errorf("Can not craete a new server")
	}

	msg := waitForMessage(&s1)

	if msg != nil {
		t.Errorf("Message have arrive which is not expected. Is means message reached with delay at destination")
	}
}
*/

/*
func TestClusterTS4(t *testing.T)  {
	s,_ := makeDummyServer(1)
	env := Envelope{1, 0, "RoundRobin"}
	sendMessageTo(&s[0],1,&env)

	time.Sleep(20*time.Second)
	s1, err := New(1,"Config.json")
	if err != nil {
		t.Errorf("Can not craete a new server")
	}

	msg := waitForMessage(&s1)

	if msg != nil {
		t.Errorf("Message have arrive which is not expected. Is means message reached with delay at destination")
	}
}

*/
// The purpose of this test case to check the limits of the buffer.
// This testCase craete 2 server instances and send the message at maximum processor capability

func TestClusterTS5(t *testing.T) {
	s, err := makeDummyServer(NOFSERVER)
	if err != nil {
		t.Errorf("Server Instantiation Failed %q", err)
	}
	go sendMessage(&s[0], 0)
	go sendMessage(&s[1], 0)
	go sendMessage(&s[2], 0)
	go sendMessage(&s[3], 0)
	count := make([]int, NOFSERVER)

	for {
		select {
		case val := <-s[1].Inbox():
			fmt.Println("case 1 ", val)
			count[1] += 1
		case val := <-s[1].Inbox():
			fmt.Println("case 1 ", val)
			count[1] += 1
		case val := <-s[2].Inbox():
			fmt.Println("case 2 ", val)
			count[2] += 1
		case val := <-s[3].Inbox():
			fmt.Println("case 3 ", val)
			count[3] += 1
		case <-time.After(3 * time.Second):
			break
		}
	}
	fmt.Println("Reached")

	if count[1] != NOFMESSAGE || count[2] != NOFMESSAGE || count[3] != NOFMESSAGE {
		t.Errorf("Server Instantiation Failed %q", err)
	}

}
