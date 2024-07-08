package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type saveMsg struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

var store map[int]struct{}

func main() {
	store = make(map[int]struct{})
	n := maelstrom.NewNode()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body saveMsg
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		store[body.Message] = struct{}{}

		ids := n.NodeIDs()

		for _, v := range ids {
			if v == n.ID() {
				continue
			}
			n.Send(v, saveMsg{
				Type:    "save",
				Message: body.Message,
			})
		}

		return n.Reply(msg, map[string]string{
			"type": "broadcast_ok",
		})
	})

	n.Handle("save", func(msg maelstrom.Message) error {
		var body saveMsg
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		log.Println(body.Message)

		store[body.Message] = struct{}{}

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var arr []int = []int{}

		for k := range store {
			arr = append(arr, k)
		}

		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": arr,
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]string{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
