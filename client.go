package main

import (
	"github.com/gorilla/websocket"
)

// client represents a single chatting user.
type client struct {
	// socket is the web socket for this client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the room this client is chatting in.
	room *room
}

// read method allows our client to read from the socket via the ReadMessage method, continually sending any received messages to the forward channel on  the room type. If it encounters an error (such as 'the socket has died'), the loop will break and the socket will be closed
func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

//write method continually accepts messages from the send channel writing everything out of the socket via  the WriteMessage method. If writing to the socket fails, the for loop is broken  and the socket is closed.
func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
