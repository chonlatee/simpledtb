package main

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type worker struct {
	name       string
	remoteAddr string
	recv       chan []byte
	send       chan []byte
}

func NewWorker(name, remoteAddr string) (*worker, error) {
	w := &worker{
		name:       name,
		remoteAddr: remoteAddr,
		recv:       make(chan []byte),
		send:       make(chan []byte),
	}

	err := w.dial()
	if err != nil {
		return nil, err
	}

	log.Printf("worker %s dial on %s success\n", name, remoteAddr)

	return w, nil
}

func (w *worker) dial() error {
	c, err := net.Dial("tcp", w.remoteAddr)
	if err != nil {
		log.Printf("worker: %s dial %s err %s\n", w.name, w.remoteAddr, err)
		return err
	}

	go w.sendMsg(c)
	return nil
}

func (w *worker) sendMsg(conn net.Conn) {
	log.Println("prepare write msg to network")
	for {
		msg := <-w.send
		log.Printf("%s write msg %s to network", w.name, msg)
		_, err := conn.Write(msg)
		if err != nil {
			continue
		}
	}
}

type server struct {
	mu      *sync.Mutex
	workers map[string]*worker
	msg     chan []byte
}

func NewServer() (*server, error) {
	s := &server{
		mu:      new(sync.Mutex),
		workers: make(map[string]*worker),
		msg:     make(chan []byte),
	}

	go s.assignMsgToWorker()

	return s, nil
}

func (s *server) addWorker(w *worker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers[w.name] = w
}

func (s *server) assignMsgToWorker() {
	ws := []string{"w1", "w2"}
	for {
		log.Println("assign msg to worker")
		m := <-s.msg
		n := ws[rand.Intn(2)]
		log.Printf("send %s to channel of worker %s \n", string(m), n)
		s.workers[n].send <- m
	}
}

func main() {

	s, err := NewServer()
	if err != nil {
		log.Fatalf("start server error: %v", err)
	}

	w1, err := NewWorker("w1", ":3333")
	if err != nil {
		log.Fatalf("worker 1 err: %v", err)
	}
	s.addWorker(w1)
	w2, err := NewWorker("w2", ":3334")
	if err != nil {
		log.Fatalf("worker 2 err: %v", err)
	}
	s.addWorker(w2)

	msg := []string{"foo", "bar", "baz", "foobar", "foobaz"}
	for {
		time.Sleep(time.Second * 2)
		m := msg[rand.Intn(4)]
		log.Println("send msg to chan: ", m)
		s.msg <- []byte(m)
	}
}
