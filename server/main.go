package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	port = ":3221"
)

type Client struct {
	conn   net.Conn
	roomID string
}
type Message struct {
	text   string
	conn   net.Conn
	roomID string
}
type Room struct {
	roomID string
}

func server(mesgs chan Message) {
	clients := make(map[string]*Client)
	for {
		msg := <-mesgs
		clients[msg.conn.RemoteAddr().String()] = &Client{
			conn:   msg.conn,
			roomID: msg.roomID,
		}

		for _, client := range clients {
			if msg.text == "init_roomid" {
				break
			}
			if err := sendMessage(client, &msg); err != nil {
				log.Println(err)
			}
		}
	}
}

func client(conn net.Conn, messages chan Message, defaultRoom string) {
	buf := make([]byte, 64)
	var roomidx string
	for {

		n, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			return
		}
		text := string(buf[0:n])
		if strings.Contains(text, ":roomid") {

			roomidx = text[8:]
			defaultRoom = roomidx
			conn.Write([]byte(roomidx))
			messages <- Message{
				text:   "init_roomid",
				conn:   conn,
				roomID: roomidx,
			}
		} else {
			messages <- Message{
				text:   text,
				conn:   conn,
				roomID: roomidx,
			}
		}

	}
}

func sendMessage(client *Client, msg *Message) error {
	if msg.roomID == client.roomID {

		hashedIP := hashIP(msg.conn.RemoteAddr().String())
		msg := fmt.Sprintf("%s: %s\n", hashedIP[0:10], msg.text)
		client.conn.Write([]byte(msg))
		return nil
	} else {
		return fmt.Errorf("room id is not equal")
	}
}

func hashIP(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func main() {
	defaultRoom := "0"
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
			conn:   conn,
			roomID: defaultRoom,
		}
		go client(conn, messages, defaultRoom)
	}
}
