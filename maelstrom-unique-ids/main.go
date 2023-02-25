package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var counter = 0

func main() {
	var n *maelstrom.Node = maelstrom.NewNode()
	n.Handle("generate", func(msg maelstrom.Message) error {
		// Unmarshal
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "generate_ok"
		// Generate UUID and add to body with key id
		// UUID = Node ID + time of day + running counter
		uuid := msg.Dest + fmt.Sprintf("%v", time.Now().UnixNano()) + fmt.Sprintf("%v", counter)
		body["id"] = uuid
		counter = counter + 1
		return n.Reply(msg, body)
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
