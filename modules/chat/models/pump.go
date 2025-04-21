package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// ReadPump handles reading messages from a client
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		c.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Time{})
	// c.Conn.SetPongHandler(func(string) error {
	// 	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println(err)
			}
			break
		}
		if string(msg) == "PONG" {
			fmt.Println("Received PONG")
			c.PingNotify <- struct{}{}
			continue
		}
		if string(msg) == "PING" {
			fmt.Println("Received PING")
			c.writeMessage([]byte("PONG"), true)
			fmt.Println("SENDING PONG")
		}
	}
}

func (c *Client) writeMessage(message []byte, ok bool) error {
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	if !ok {
		// The hub closed the channel
		c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
		return errors.New("hub closed the channel")
	}

	w, err := c.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write(message)

	// Add queued messages to the current websocket message
	n := len(c.Send)
	for range n {
		w.Write(newline)
		w.Write(<-c.Send)
	}

	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// WritePump handles sending messages to a client
func (c *Client) WritePump(hub *Hub) {
	ticker := time.NewTicker(time.Second * 2)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			err := c.writeMessage(message, ok)
			if err != nil {
				fmt.Println("Error writing message:", err)
				return
			}

		case <-ticker.C:
			fmt.Println("Sending ping")
			err := c.writeMessage([]byte("PING"), true)
			if err != nil {
				fmt.Println("Error sending ping:", err)
				return
			}
			select {
			case <-c.PingNotify:
				continue
			case <-time.After(time.Second * 5):
				fmt.Println("Ping timeout")
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		}
	}
}

// Constants for WebSocket
const (
	// Time allowed to write a message to the peer
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 5 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 1024
)

var (
	newline = []byte{'\n'}
)
