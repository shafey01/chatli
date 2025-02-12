package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net"
)

const (
	port = ":3221"
)

type Client struct {
	conn net.Conn
}
type Message struct {
	text string
	conn net.Conn
}

func server(mesgs chan Message) {
	clients := make(map[string]*Client)
	for {
		msg := <-mesgs
		clients[msg.conn.RemoteAddr().String()] = &Client{
			conn: msg.conn,
		}
		for _, client := range clients {
			hashedIP := hashIP(msg.conn.RemoteAddr().String())
			msgIp := fmt.Sprintf("%s: %s\n", hashedIP[0:10], msg.text)
			client.conn.Write([]byte(msgIp))
		}
	}
}

func client(conn net.Conn, messages chan Message) {
	buf := make([]byte, 64)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			return
		}
		text := string(buf[0:n])
		messages <- Message{
			text: text,
			conn: conn,
		}
	}
}

func hashIP(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		// handle error
		log.Fatal("couldn't start the Server...", err)
	}
	messages := make(chan Message)
	go server(messages)
	for {
		conn, err := lis.Accept()
		if err != nil {
			// handle error
			log.Printf("couldn't Accept the connection: %v  ", err)
		}
		messages <- Message{
			conn: conn,
		}
		go client(conn, messages)
	}
}
