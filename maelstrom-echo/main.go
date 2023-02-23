package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	// Register a handler callback function for the "echo" message
	n.Handle("echo", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map
		var body map[string]any
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return err
		}
		// Update the message type to return back
		body["type"] = "echo_ok"
		// Return the original message back with the updated message type
		return n.Reply(msg, body)
	})
	err := n.Run()
	if err != nil {
		log.Fatal(err)
	}
}
