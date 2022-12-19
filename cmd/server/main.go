package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
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
	recv    chan []byte
	ln      net.Listener
}

func NewServer(addr string) (*server, error) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("listen on %s err %v", addr, err)
		return nil, err
	}

	s := &server{
		mu:      new(sync.Mutex),
		workers: make(map[string]*worker),
		msg:     make(chan []byte),
		recv:    make(chan []byte, 10),
		ln:      l,
	}

	go s.accept()

	return s, nil
}

func (s *server) accept() {
	log.Println("prepare for accept msg from worker")
	for {
		c, err := s.ln.Accept()
		if err != nil {
			continue
		}
		go s.recvMsg(c)
	}
}

func (s *server) recvMsg(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			continue
		}

		s.recv <- buf
	}
}

func (s *server) addWorker(w *worker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workers[w.name] = w
}

func (s *server) assignMsgToWorker() {
	var names []string
	for _, v := range s.workers {
		names = append(names, v.name)
	}

	for {
		log.Println("assign msg to worker")
		m := <-s.msg
		n := names[rand.Intn(len(names))]
		log.Printf("send %s to channel of worker %s \n", string(m), n)
		s.workers[n].send <- m
	}
}

type workerConfig struct {
	Name string
	Addr string
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "-port=:3000")
	flag.Parse()

	s, err := NewServer(port)
	if err != nil {
		log.Fatalf("start server error: %v", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("can't get current working dir: ", err.Error())
	}

	f, err := ioutil.ReadFile(dir + "/config/worker.yaml")
	if err != nil {
		log.Fatalf("can't open worker config file")
	}

	var wkCfg []workerConfig
	err = yaml.Unmarshal(f, &wkCfg)

	if err != nil {
		log.Fatalf("can't unmarshal worker config file: %v", err)
	}

	for _, v := range wkCfg {
		w, err := NewWorker(v.Name, v.Addr)
		if err != nil {
			log.Fatalf("%s err %v", v.Name, err)
		}
		s.addWorker(w)
	}

	go s.assignMsgToWorker()

	go func() {
		for {
			m := <-s.recv
			log.Printf("receive msg %s from worker\n", string(m))
		}
	}()

	msg := []string{"foo", "bar", "baz", "foobar", "foobaz"}
	for {
		time.Sleep(time.Second * 2)
		m := msg[rand.Intn(len(msg)-1)]
		log.Println("send msg to chan: ", m)
		s.msg <- []byte(m)
	}
}
