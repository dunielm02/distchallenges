package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	n.Handle("add", func(msg maelstrom.Message) error {
		body := AddMsg{}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		counter, err := kv.ReadInt(context.Background(), "counter")
		if err != nil {
			counter = 0
			kv.Write(context.Background(), "counter", 0)
		}
		for kv.CompareAndSwap(context.Background(), "counter", counter, counter+body.Delta, false) != nil {
			counter, _ = kv.ReadInt(context.Background(), "counter")
		}

		kv.Write(context.Background(), fmt.Sprintf("random_%d", rand.Int()), rand.Int())

		return n.Reply(msg, map[string]string{
			"type": "add_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		counter, err := kv.ReadInt(context.Background(), "counter")
		if err != nil {
			counter = 0
			kv.CompareAndSwap(context.Background(), "counter", 0, 0, true)
		}
		return n.Reply(msg, ReadMsg{
			Type:  "read_ok",
			Value: counter,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

type AddMsg struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
}

type ReadMsg struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}
