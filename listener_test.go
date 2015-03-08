package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLen(t *testing.T) {
	a, _ := strconv.Atoi("12344324")
	t.Log(a)
}

func TestListener(t *testing.T) {
	var wg sync.WaitGroup
	cert, privKey, _ := genCert()
	listener, err := Listen(cert, privKey, "localhost:8091", true)
	if err != nil {
		t.Fatal(err)
	}

	go listener.Accept(func(msg []byte) {
		wg.Add(1)
		defer wg.Done()
		t.Logf("Incoming message to listener from a client: %v", string(msg))
	})

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(cert)
	if !ok {
		panic("failed to parse root certificate")
	}

	conn, err := tls.Dial("tcp", "localhost:8091", &tls.Config{RootCAs: roots})
	if err != nil {
		t.Fatal(err)
	}

	send(t, conn, "4   ping")
	send(t, conn, "56  Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
	send(t, conn, "49  In sit amet lectus felis, at pellentesque turpis.")
	send(t, conn, "64  Nunc urna enim, cursus varius aliquet ac, imperdiet eget tellus.")
	// send(t, conn, "8671"+randString(8671))
	send(t, conn, "5   close")

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	conn.Close()
	listener.Close()
}

func send(t *testing.T, conn *tls.Conn, msg string) {
	n, err := io.WriteString(conn, msg)
	if err != nil {
		t.Fatalf("Error while writing message to connection %v", err)
	}
	t.Logf("Sending message to listener from client: %v (%v bytes)", msg, n)
}