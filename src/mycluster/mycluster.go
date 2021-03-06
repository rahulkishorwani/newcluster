//package main
package mycluster

//7588554572
import (
	"encoding/xml"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*func main() {

	//testcase1()
	//testcase2()

	testcase3()
	//testcase4()

}*/

func New(id string, sizeofinchan int, sizeooutchan int, fnm string, delaybeforeconn time.Duration) Servermainstruct {

	//iid,ip,port,serv,ownindex:=getithserveraddr(id,fnm)
	iid, _, port, serv, ownindex := getithserveraddr(id, fnm)

	//fmt.Printf("%s %s\n",ip,port)

	inchan := make(chan *Envelope, sizeofinchan)
	outchan := make(chan *Envelope, sizeooutchan)

	//inchan:=make(chan *Envelope)
	//outchan:=make(chan *Envelope)

	//newservsock, _ := zmq.NewSocket(zmq.DEALER)
	newservsock, _ := zmq.NewSocket(zmq.PULL)

	addr := "tcp://*:" + port
	newservsock.Bind(addr)
	//fmt.Printf("Listening on %s : %s\n",ip,port)
	//var newpeersock [10]*zmq.Socket

	newpeersock := make([]*zmq.Socket, len(serv))

	Actuallysnd := make([][2]int, len(serv))
	Actuallyrcvd := make([][2]int, len(serv))

	//time.Sleep(30000 * time.Millisecond)
	time.Sleep(delaybeforeconn)

	for k, _ := range serv {
		if k != ownindex {
			//n, _ := zmq.NewSocket(zmq.DEALER)
			n, _ := zmq.NewSocket(zmq.PUSH)

			str := "tcp://" + serv[k].Ip + ":" + serv[k].Port

			n.Connect(str)

			newpeersock[k] = n
			//fmt.Printf("Connected on %s %s\n",serv[k].Port,str)

		}

	}
	limit := len(serv)
	mynewpeersock := newpeersock[:limit]

	m := Servermainstruct{servsock: newservsock, peersock: mynewpeersock, pid: iid, peers: serv, ownindexinpeers: ownindex, in: inchan, out: outchan, Actuallysend: Actuallysnd, Actuallyrecvd: Actuallyrcvd}

	return m
}
func Serverfunc(myservermainstruct Servermainstruct, id string, noofmessagestosend int, numberofservers int, sizeofinchan int, sizeooutchan int, wg *sync.WaitGroup) {
	//ffnm := "serverlist" + id + ".xml"
	//myservermainstruct := New(id,sizeofinchan,sizeooutchan,ffnm)

	go myservermainstruct.send(noofmessagestosend, numberofservers)
	go myservermainstruct.receive(noofmessagestosend, numberofservers)

	//go myservermainstruct.Sendtooutbox(id, noofmessagestosend, numberofservers)

	var connarr = make([]int, numberofservers)
	for k := 0; k < numberofservers; k++ {
		connarr[k] = 0
	}

	//fnm := "recvd" + id

	tt, _ := strconv.Atoi(id)
	connarr[tt] = 1
	//f2, _ := os.OpenFile(fnm, os.O_WRONLY|os.O_CREATE, 0777)
	for i := 0; ; {

		envelope := <-myservermainstruct.Inbox()
		//fmt.Printf("Received msg from %s %d: '%s'\n", envelope.Pid,envelope.MsgId,envelope.Msg)
		//fmt.Fprintf(f2, "%s,%s,%d,%s\n", envelope.Pid, id, envelope.MsgId, envelope.Msg)

		if strings.EqualFold(envelope.Msg.(string), "FIN") {
			tmp, _ := strconv.Atoi(envelope.Pid)
			connarr[tmp] = 1
		}

		j := 0
		for j = 0; j < numberofservers; j++ {
			if connarr[j] == 0 {
				break
			}
		}
		if j == numberofservers {
			break
		}
		/*if( connarr[0]==1 && connarr[1]==1 && connarr[2]==1 && connarr[3]==1 && connarr[4]==1 ) {
			break
		}*/

		i++

	}
	//f2.Close()

	wg.Done()

}

const (
	BROADCAST = "-1"
)

type Envelope struct {
	Pid string

	MsgId int

	Msg interface{}
}

type Myserver interface {
	Pid() string

	Peers() []Server

	Outbox() chan *Envelope

	Inbox() chan *Envelope

	Sendtooutbox(string, int)
}

type Servermainstruct struct {
	servsock        *zmq.Socket
	peersock        []*zmq.Socket
	pid             string
	peers           []Server
	ownindexinpeers int
	in              chan *Envelope
	out             chan *Envelope
	Actuallysend    [][2]int
	Actuallyrecvd   [][2]int
}

func (servermainstruct Servermainstruct) Pid() string {
	return servermainstruct.pid
}
func (servermainstruct Servermainstruct) Peers() []Server {
	return servermainstruct.peers
}
func (servermainstruct Servermainstruct) Outbox() chan *Envelope {
	return servermainstruct.out
}
func (servermainstruct Servermainstruct) Inbox() chan *Envelope {
	return servermainstruct.in
}

func (servermainstruct Servermainstruct) Sendtooutbox(id string, noofmessagestosend int, broadcastmsg string, peermsg string, noofbroadcastmsgs int, noofpeermsgs int, numberofservers int) {

	//fnm := "outbuffer" + id
	//f2, _ := os.OpenFile(fnm, os.O_WRONLY|os.O_CREATE, 0777)

	i, j := 0, 0
	for i, j = 0, 0; i < noofpeermsgs || j < noofbroadcastmsgs; {
		index := rand.Intn(numberofservers)
		q := servermainstruct.peers[index].Id

		//fmt.Printf("\nqqq:%s",q)
		if strings.EqualFold(id, q) {
			if j >= noofbroadcastmsgs {
				continue
			}
			j++
			//servermainstruct.Outbox() <- &Envelope{Pid: BROADCAST, MsgId: i, Msg: "BROADCAST"}
			servermainstruct.Outbox() <- &Envelope{Pid: BROADCAST, MsgId: i + j, Msg: broadcastmsg}
			//fmt.Fprintf(f2, "%s,%s,%d,%s\n", id, BROADCAST, i+j, broadcastmsg)

		} else {
			if i >= noofpeermsgs {
				continue
			}
			i++
			servermainstruct.Outbox() <- &Envelope{Pid: q, MsgId: i + j, Msg: peermsg}
			//fmt.Fprintf(f2, "%s,%s,%d,%s\n", id, q, i, peermsg)
		}

	}
	//fmt.Printf("End  %d %d",i,j)

	servermainstruct.Outbox() <- &Envelope{Pid: BROADCAST, MsgId: noofmessagestosend, Msg: "FIN"}

	//f2.Close()

}

func (servermainstruct Servermainstruct) Sendtooutbox3(id string, noofmessagestosend int, broadcastmsg string, peermsg string, noofbroadcastmsgs int, noofpeermsgs int, numberofservers int) {

	//fnm := "outbuffer" + id
	//f2, _ := os.OpenFile(fnm, os.O_WRONLY|os.O_CREATE, 0777)

	k := 0
	i, j := 0, 0
	for i, j = 0, 0; i < noofpeermsgs || j < noofbroadcastmsgs; {
		//index:=rand.Intn(numberofservers)
		index := k % numberofservers
		k++
		q := servermainstruct.peers[index].Id

		if strings.EqualFold(id, q) {
			if j >= noofbroadcastmsgs {
				continue
			}
			j++
			//servermainstruct.Outbox() <- &Envelope{Pid: BROADCAST, MsgId: i, Msg: "BROADCAST"}
			servermainstruct.Outbox() <- &Envelope{Pid: BROADCAST, MsgId: i + j, Msg: broadcastmsg}
			//fmt.Fprintf(f2, "%s,%s,%d,%s\n", id, BROADCAST, i+j, broadcastmsg)

		} else {
			if i >= noofpeermsgs {
				continue
			}
			i++
			servermainstruct.Outbox() <- &Envelope{Pid: q, MsgId: i + j, Msg: peermsg}
			//fmt.Fprintf(f2, "%s,%s,%d,%s\n", id, q, i, peermsg)
		}

	}
	//fmt.Printf("End  %d %d",i,j)

	servermainstruct.Outbox() <- &Envelope{Pid: BROADCAST, MsgId: noofmessagestosend, Msg: "FIN"}

	//f2.Close()

}

func (servermainstruct Servermainstruct) send(noofmessagestosend int, numberofservers int) {
	//fnm := "send" + servermainstruct.Pid()
	//f2, _ := os.OpenFile(fnm, os.O_WRONLY|os.O_CREATE, 0777)
	for {

		d := <-servermainstruct.out
		tmp := d.Pid
		d.Pid = servermainstruct.Pid()

		b := "Id:" + d.Pid + ",MsgId:" + strconv.Itoa(d.MsgId) + ",Msg:" + d.Msg.(string)

		for k, _ := range servermainstruct.peers {

			if k != servermainstruct.ownindexinpeers {
				if strings.EqualFold(tmp, "-1") {
					//fmt.Printf("Sending to %d",k)
					servermainstruct.peersock[k].Send(b, 0)
					//fmt.Fprintf(f2, "%s,%s,%d,%s\n", d.Pid, servermainstruct.peers[k].Id, d.MsgId, d.Msg.(string))
					servermainstruct.Actuallysend[k][0]++

				} else {

					if strings.EqualFold(tmp, servermainstruct.peers[k].Id) {
						//fmt.Printf("Sending to %d",k)
						servermainstruct.peersock[k].Send(b, 0)
						//fmt.Fprintf(f2, "%s,%s,%d,%s\n", d.Pid, servermainstruct.peers[k].Id, d.MsgId, d.Msg.(string))
						servermainstruct.Actuallysend[k][1]++
					}
				}
			}
		}
		if strings.EqualFold(d.Msg.(string), "FIN") {
			break
		}

	}
	//f2.Close()
}

func (servermainstruct Servermainstruct) receive(noofmessagestosend int, numberofservers int) {

	//fnm := "inbuffer" + servermainstruct.Pid()
	//f2, _ := os.OpenFile(fnm, os.O_WRONLY|os.O_CREATE, 0777)

	var connarr = make([]int, numberofservers)
	for k := 0; k < numberofservers; k++ {
		connarr[k] = 0
	}

	for {

		servermainstruct.servsock.SetRcvtimeo(-1)
		msg, err := servermainstruct.servsock.Recv(0)

		if err != nil {
			estr := "Receiver problem in " + servermainstruct.pid
			panic(estr)
		}

		v := strings.Split(msg, ",")
		vv := strings.Split(v[0], ":")
		vvv := strings.Split(v[1], ":")
		vvvv := strings.Split(v[2], ":")
		o, _ := strconv.Atoi(vvv[1])
		q := Envelope{Pid: vv[1], MsgId: o, Msg: vvvv[1]}

		//fmt.Printf("Received:%s",msg)
		//fmt.Fprintf(f2, "%s,%s,%d,%s\n", vv[1], servermainstruct.Pid(), o, vvvv[1])

		servermainstruct.Inbox() <- &q

		if strings.EqualFold(vvvv[1], "FIN") {
			tmp, _ := strconv.Atoi(vv[1])
			connarr[tmp] = 1
		}
		j := 0
		for j = 0; j < numberofservers; j++ {
			if connarr[j] == 0 {
				break
			}
		}
		if j == numberofservers {
			break
		}

		k, _ := strconv.Atoi(vv[1])
		if strings.HasPrefix(vvvv[1], "BROADCAST") {
			servermainstruct.Actuallyrecvd[k][0]++
		} else {
			servermainstruct.Actuallyrecvd[k][1]++
		}

	}
	//f2.Close()
}

type Serverinfo struct {
	XMLName    Servermeta `xml:"serverinfo"`
	Servermeta Servermeta `xml:"servermeta"`
	Serverlist Serverlist `xml:"serverlist"`
}
type Servermeta struct {
	Servercount int `xml:"servercount"`
}
type Serverlist struct {
	Server []Server `xml:"server"`
}
type Server struct {
	Id   string `xml:"id"`
	Ip   string `xml:"ip"`
	Port string `xml:"port"`
}

/*max char of xmlfile are 10000 words*/
func getithserveraddr(id string, fnm string) (string, string, string, []Server, int) {
	//xmlFile, err := os.Open("serverlist.xml")
	xmlFile, err := os.Open(fnm)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return "0", "0", "0", nil, -1
	}
	defer xmlFile.Close()
	data := make([]byte, 10000)
	count, err := xmlFile.Read(data)
	if err != nil {
		fmt.Println("Can't read the data", count, err)
		return "0", "0", "0", nil, -1
	}
	var q Serverinfo
	xml.Unmarshal(data[:count], &q)
	checkError(err)

	for k, sobj := range q.Serverlist.Server {
		if strings.EqualFold(sobj.Id, id) {
			return q.Serverlist.Server[k].Id, q.Serverlist.Server[k].Ip, q.Serverlist.Server[k].Port, q.Serverlist.Server, k

		}
	}

	return "0", "0", "0", nil, -1
}
func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
