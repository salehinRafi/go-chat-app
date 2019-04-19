package main

import (
	"github.com/gorilla/websocket"
	"goWork_chat/trace"
	"log"
	"net/http"
)

type room struct {
	forward chan []byte      // forward is a channel that holds incoming messages that should be forwarded to the other clients.
	join    chan *client     // join is a channel for clients wishing to join the room.
	leave   chan *client     // leave is a channel for clients wishing to leave the room.
	clients map[*client]bool // clients holds all current clients in this room.
	tracer  trace.Tracer     // tracer will receive trace information of activity in the room.
}

// users of our code need only call the newRoom function (to start to use)
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	// The top for loop indicates that this method will run forever, until the program is terminated.
	// if we run this code as a Go routine, it will run in the background, which won't block the rest of our application.
	for {
		// Keep watching the three channel
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true // update the r.clients map to keep a reference of the client that has joined the room
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			// leaving
			delete(r.clients, client) // delete the client type from the map
			close(client.send)
			r.tracer.Trace("client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", string(msg))
			// forward message to all clients
			for client := range r.clients { // iterate over all the clients
				client.send <- msg
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

/*Turning room into a HTTP handler*/
const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// reusable so we need only create one
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write() //go routines
	client.read()     // will block operations (keeping the connection alive) until it's time to close it.
}
