// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"encoding/json"
	"fmt"
	"log"
)

type BroadcastMessage struct {
	GroupId string `json:"groupId"`
	Message string `json:"message"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]map[string]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("new client %s connected\n", client.ID)
			groupId := client.groupId
			if h.clients[groupId] == nil {
				h.clients[groupId] = make(map[string]*Client)
			}

			h.clients[groupId][client.ID] = client

		case client := <-h.unregister:
			log.Printf("client %s disconected\n", client.ID)
			if _, ok := h.clients[client.groupId]; ok {
				delete(h.clients[client.groupId], client.ID)
				close(client.send)
			}
		case message := <-h.broadcast:
			var payload BroadcastMessage
			json.Unmarshal(message, &payload)

			log.Printf("broadcast mesage for group %s", payload.GroupId)

			clients := h.clients[payload.GroupId]
			for _, client := range clients {
				select {
				case client.send <- []byte(payload.Message):
				default:
					fmt.Println("aqui")
					close(client.send)
					delete(h.clients[payload.GroupId], client.ID)
				}
			}
		}
	}
}
