package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func generateId() string {
	now := time.Now().Unix()
	ran := rand.Int31()
	return fmt.Sprintf("%d%d", now, ran)
}

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "generate_ok"
		body["id"] = generateId()

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
