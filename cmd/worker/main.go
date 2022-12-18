package main

import (
	"flag"
	"log"
	"net"
	"strings"
)

type server struct {
	recv chan []byte
	send chan []byte
	ln   net.Listener
}

func NewServer(addr string) (*server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("listen on %s err: %v", addr, err)
		return nil, err
	}

	s := &server{
		recv: make(chan []byte),
		send: make(chan []byte),
		ln:   l,
	}

	go s.accept()

	go s.dialServer()

	return s, nil

}

func (s *server) dialServer() {

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

	s, err := NewServer(port)
	if err != nil {
		log.Fatalf("start server err: %v\n", err)
	}

	log.Printf("start worker server on %s success\n", port)

	for v := range s.recv {
		log.Println(upperString(string(v)))
	}

}
