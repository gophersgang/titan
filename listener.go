package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

var (
	ping   = []byte("ping")
	closed = []byte("close")
)

// Listener accepts connections from devices.
type Listener struct {
	Conns    []*Conn
	debug    bool
	listener net.Listener
	connwg   *sync.WaitGroup
	reqwg    *sync.WaitGroup
}

// Listen creates a TCP listener with the given PEM encoded X.509 certificate and the private key on the local network address laddr.
// Debug mode logs all server activity.
func Listen(cert, privKey []byte, laddr string, connwg *sync.WaitGroup, reqwg *sync.WaitGroup, debug bool) (*Listener, error) {
	tlsCert, err := tls.X509KeyPair(cert, privKey)
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(cert)
	if err != nil || !ok {
		return nil, fmt.Errorf("failed to parse the certificate or the private key: %v", err)
	}

	conf := tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    pool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
	}

	l, err := tls.Listen("tcp", laddr, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS listener no network address %v: %v", laddr, err)
	}
	if debug {
		log.Printf("Listener created with local network address: %v\n", laddr)
	}

	return &Listener{
		Conns:    make([]*Conn, 0),
		debug:    debug,
		listener: l,
		connwg:   connwg,
		reqwg:    reqwg,
	}, nil
}

// Accept waits for incoming connections and forwards the client connect/message/disconnect events to provided handlers in a new goroutine.
// This function blocks and never returns, unless there is an error while accepting a new connection.
func (l *Listener) Accept(handleMsg func(conn *Conn, session *Session, msg []byte), handleDisconn func(conn *Conn, session *Session)) error {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			if operr, ok := err.(*net.OpError); ok && operr.Op == "accept" && operr.Err.Error() == "use of closed network connection" {
				return nil
			}
			return fmt.Errorf("error while accepting a new connection from a client: %v", err)
			// todo: it might not be appropriate to break the loop on recoverable errors (like client disconnect during handshake)
			// the underlying fd.accept() does some basic recovery though we might need more: http://golang.org/src/net/fd_unix.go
		}

		tlsconn, ok := conn.(*tls.Conn)
		if !ok {
			conn.Close()
			return errors.New("cannot cast net.Conn interface to tls.Conn type")
		}
		if l.debug {
			log.Println("Client connected: listening for messages from client IP:", conn.RemoteAddr())
		}

		l.connwg.Add(1)
		c := NewConn(tlsconn, 0, 0, 0)
		l.Conns = append(l.Conns, c)
		go handleClient(l.connwg, l.reqwg, c, l.debug, handleMsg, handleDisconn)
	}
}

// handleClient waits for messages from the connected client and forwards the client message/disconnect
// events to provided handlers in a new goroutine.
// This function never returns, unless there is an error while reading from the channel or the client disconnects.
func handleClient(connwg *sync.WaitGroup, reqwg *sync.WaitGroup, conn *Conn, debug bool, handleMsg func(conn *Conn, session *Session, msg []byte), handleDisconn func(conn *Conn, session *Session)) error {
	defer connwg.Done()

	session := &Session{}

	if debug {
		defer func() {
			if session.Disconnected {
				log.Println("Client disconnected on IP:", conn.RemoteAddr())
			} else {
				log.Println("Closed connection to client with IP:", conn.RemoteAddr())
			}
		}()
	}
	defer func() {
		session.Error = conn.Close() // todo: handle close error, store the error in conn object and return it to handleMsg/handleErr/handleDisconn or one level up (to server)
	}()

	for {
		if session.Error != nil {
			// todo: send error message to user, log the error, and close the conn and return
			return session.Error
		}

		n, msg, err := conn.Read()
		if err != nil {
			if err == io.EOF {
				session.Disconnected = true
				break
			}
			if operr, ok := err.(*net.OpError); ok && operr.Op == "read" && operr.Err.Error() == "use of closed network connection" {
				session.Disconnected = true
				break
			}
			log.Fatalln("Errored while reading:", err)
		}

		// shortcut 'ping' and 'close' messages, saves some processing time
		if n == 4 && bytes.Equal(msg, ping) {
			continue // send back pong?
		}
		if n == 5 && bytes.Equal(msg, closed) {
			reqwg.Add(1)
			go func() {
				defer reqwg.Done()
				handleDisconn(conn, session)
			}()
			return session.Error
		}

		reqwg.Add(1)
		go func() {
			defer reqwg.Done()
			handleMsg(conn, session, msg)
		}()
	}

	return session.Error
}

// Close closes the listener.
func (l *Listener) Close() error {
	if l.debug {
		defer log.Println("Listener was closed on local network address:", l.listener.Addr())
	}
	return l.listener.Close()
}
