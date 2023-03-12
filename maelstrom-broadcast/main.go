package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"golang.org/x/exp/slices"
)

var lock sync.RWMutex
var messages = make([]float64, 0)
var neighbors = make([]string, 0)

func main() {
	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		lock.Lock()
		defer lock.Unlock()

		var body struct {
			Message float64 `json:"message"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// Prevent repeated messages
		if slices.Contains(messages, body.Message) {
			return nil
		} else {
			messages = append(messages, body.Message)
		}

		// Gossip until all neighbors have recieved the message
		for _, neighbor := range neighbors {
			// Skip the node that sent this message
			if neighbor == msg.Src {
				continue
			}
			go asyncRPC(n, neighbor, body.Message)
		}
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})
	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "read_ok"
		body["messages"] = messages
		return n.Reply(msg, body)
	})
	n.Handle("topology", func(msg maelstrom.Message) error {
		var body struct {
			Topology map[string][]string `json:"topology"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		neighbors = body.Topology[n.ID()]
		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func asyncRPC(n *maelstrom.Node, dest string, message float64) {
	sendTimeout := 50 * time.Millisecond
	hasDestRecieved := false
	body := map[string]any{"type": "broadcast", "message": message}
	// Provide custom handler for "broadcast_ok" messages
	err := n.RPC(dest, body, func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// Stop broadcasting to neighbors that respond successfully
		if body["type"] == "broadcast_ok" {
			hasDestRecieved = true
		}
		return nil
	})
	// Messages may time out or drop due to network partitions. Retry so
	// messages are propogated when nodes are able to communicate again.
	time.Sleep(sendTimeout)
	if err != nil || !hasDestRecieved {
		asyncRPC(n, dest, message)
	}
}
