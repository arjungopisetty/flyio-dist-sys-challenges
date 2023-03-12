package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"golang.org/x/exp/slices"
)

var messages = make([]int, 0)
var topology = make(map[string]any)

func main() {
	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "broadcast_ok"
		// Prevent repeated messages
		message := int(body["message"].(float64))
		if slices.Contains(messages, message) {
			return nil
		} else {
			messages = append(messages, message)
		}
		delete(body, "message")

		// Gossip
		neighbors := topology[n.ID()].([]any)
		for _, neighbor := range neighbors {
			gossipBody := map[string]any{"type": "broadcast", "message": message}
			// Provide custom handler for "broadcast_ok" messages
			n.RPC(neighbor.(string), gossipBody, func(gossipMsg maelstrom.Message) error { return nil })
		}
		return n.Reply(msg, body)
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
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "topology_ok"
		topology = body["topology"].(map[string]any)
		delete(body, "topology")
		return n.Reply(msg, body)
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
