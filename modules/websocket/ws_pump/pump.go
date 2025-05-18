package ws_pump

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func NewPump(conn *websocket.Conn, sendChan chan []byte) *Pump {
	return &Pump{
		conn:       conn,
		send:       sendChan,
		pingNotify: make(chan struct{}),

		Unregister: make(chan struct{}),
	}
}

type Pump struct {
	conn       *websocket.Conn
	send       chan []byte
	pingNotify chan struct{}

	Unregister chan struct{}
}

func (p *Pump) Start() {
	go p.writePump()
	go p.readPump()
}

// readPump handles reading messages from a client
func (p *Pump) readPump() {
	defer func() {
		p.Unregister <- struct{}{}
		p.conn.Close()
	}()

	// p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadLimit(4) //ping
	p.conn.SetReadDeadline(time.Time{})

	for {
		_, msg, err := p.conn.ReadMessage()
		if err != nil {
			// if websocket.IsUnexpectedCloseError(err,
			// 	websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// }
			fmt.Println(err)
			break
		}
		if string(msg) == "PONG" {
			fmt.Println("Received PONG")
			p.pingNotify <- struct{}{}
			continue
		}
		if string(msg) == "PING" {
			fmt.Println("Received PING")
			err := p.writeMessage([]byte("PONG"))
			if err != nil {
				break
			}
			fmt.Println("SENDING PONG")
		}
	}
}

func (c *Pump) writeMessage(message []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))

	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	_, err = w.Write(message)
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// writePump handles sending messages to a client
func (c *Pump) writePump() {
	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		fmt.Println("Closing WritePump")
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel
				// c.conn.WriteMessage(websocket.CloseMessage, []byte("CLOSED"))
				// fmt.Println(errors.New("hub closed the channel"))
				break
			}
			err := c.writeMessage(message)
			if err != nil {
				fmt.Println("Error writing message:", err)
				break
			}

		case <-ticker.C:
			fmt.Println("sending ping")
			err := c.writeMessage([]byte("PING"))
			if err != nil {
				fmt.Println("Error sending ping:", err)
				break
			}
			select {
			case <-c.pingNotify:
				continue
			case <-time.After(time.Second * 5):
				fmt.Println("Ping timeout")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
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

	// send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 1024
)
