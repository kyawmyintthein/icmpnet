package broker

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
)

// Broker type
type Broker struct {
	welcome  string
	connPool map[int]net.Conn
	cpMtx    sync.RWMutex
}

// New create a new Broker
func New(welcome string) *Broker {
	return &Broker{
		welcome:  welcome,
		connPool: make(map[int]net.Conn, 100),
	}
}

// Serve serves connections from the listener
func (b *Broker) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go b.handleConn(conn)
	}
}

func (b *Broker) handleConn(conn net.Conn) {
	key := rand.Int()
	log.Printf("Connected: %s\n", conn.RemoteAddr())
	b.storeConn(key, conn)

	b.serveConn(conn)

	log.Printf("Disconnected: %s\n", conn.RemoteAddr())
	b.deleteConn(key)
}

func (b *Broker) serveConn(conn net.Conn) {
	fmt.Fprintf(conn, b.welcome)
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			break
		}
		log.Printf("%s >> %s", conn.RemoteAddr(), msg)
		b.broadcast(msg)
	}
}

func (b *Broker) broadcast(msg string) {
	conns := b.allConns()
	for _, conn := range conns {
		fmt.Fprint(conn, msg)
	}
}

func (b *Broker) allConns() []net.Conn {
	b.cpMtx.RLock()
	defer b.cpMtx.RUnlock()
	ret := make([]net.Conn, 0, len(b.connPool))
	for _, conn := range b.connPool {
		ret = append(ret, conn)
	}
	return ret
}

func (b *Broker) loadConn(key int) net.Conn {
	b.cpMtx.RLock()
	defer b.cpMtx.RUnlock()
	return b.connPool[key]
}

func (b *Broker) storeConn(key int, conn net.Conn) {
	b.cpMtx.Lock()
	defer b.cpMtx.Unlock()
	b.connPool[key] = conn
}

func (b *Broker) deleteConn(key int) {
	b.cpMtx.Lock()
	defer b.cpMtx.Unlock()
	delete(b.connPool, key)
}
