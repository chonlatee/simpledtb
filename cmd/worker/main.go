package main

import (
	"flag"
	"log"
	"net"
	"strings"
	"time"
)

type server struct {
	recv       chan []byte
	send       chan []byte
	ln         net.Listener
	serverAddr string
}

func NewServer(addr, serverAddr string) (*server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("listen on %s err: %v", addr, err)
		return nil, err
	}

	s := &server{
		recv:       make(chan []byte, 10),
		send:       make(chan []byte, 10),
		ln:         l,
		serverAddr: serverAddr,
	}

	go s.accept()

	go s.dialServer()

	return s, nil

}

func (s *server) dialServer() {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", s.serverAddr)
		if err != nil {
			log.Println("trying to connect to server")
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}
	go s.sendMsg(conn)
}

func (s *server) sendMsg(conn net.Conn) {
	log.Println("prepare for send msg to server")
	for {
		m := <-s.send
		_, err := conn.Write(m)
		if err != nil {
			log.Printf("send msg err: %v\n", err)
		}
	}
}

func (s *server) accept() {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			log.Println("accept err")
			continue
		}
		go s.handleConn(c)
	}
}

func (s *server) handleConn(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("read err: %v", err)
			return
		}
		s.recv <- buf
	}
}

func upperString(s string) string {
	return strings.ToUpper(s)
}

func main() {
	var port string
	var serverAddr string

	flag.StringVar(&port, "port", "", "-port=:3333")
	flag.StringVar(&serverAddr, "server", "", "-server=:3000")
	flag.Parse()

	s, err := NewServer(port, serverAddr)
	if err != nil {
		log.Fatalf("start server err: %v\n", err)
	}

	log.Printf("start worker server on %s success\n", port)

	for v := range s.recv {
		s.send <- []byte(upperString(string(v)))
	}

}
