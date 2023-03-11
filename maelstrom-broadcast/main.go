package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"golang.org/x/exp/slices"
)

var messages = make([]int, 0)
var broadcasedMessages = make([]int, 0)
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
		newMessages := compareSlices(messages, broadcasedMessages)
		for _, neighbor := range neighbors {
			for _, message := range newMessages {
				gossipBody := map[string]any{"type": "broadcast", "message": message}
				// Provide custom handler for "broadcast_ok" messages
				n.RPC(neighbor.(string), gossipBody, func(gossipMsg maelstrom.Message) error { return nil })
				broadcasedMessages = append(broadcasedMessages, message)
			}
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

// Returns the difference from two slices
func compareSlices(a []int, b []int) []int {
	diffSlice := make([]int, 0)
	trackingMap := make(map[int]int)
	for _, i := range a {
		trackingMap[i] = 1
	}
	for _, i := range b {
		trackingMap[i] = trackingMap[i] + 1
	}
	for key, val := range trackingMap {
		if val == 1 {
			diffSlice = append(diffSlice, key)
		}
	}

	return diffSlice
}
