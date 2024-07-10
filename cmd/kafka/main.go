package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	store := make(LogStore)

	n := maelstrom.NewNode()

	n.Handle("send", func(msg maelstrom.Message) error {
		var body SendMsg
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		offset := store.Save(body.Key, body.Msg)

		return n.Reply(msg, map[string]interface{}{
			"type":   "send_ok",
			"offset": offset,
		})
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		var body PollCommitMsg
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		res := PollRes{
			Type: "poll_ok",
			Msgs: make(map[string][][2]int),
		}

		for k, v := range body.Offsets {
			res.Msgs[k] = store.Poll(k, v)
		}

		return n.Reply(msg, res)
	})

	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		var body PollCommitMsg
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		for k, v := range body.Offsets {
			if store[k] == nil {
				store[k] = &Logs{}
			}
			store[k].Committed = v
		}

		return n.Reply(msg, map[string]string{
			"type": "commit_offsets_ok",
		})
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var body CommittedMsg
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		res := make(map[string]int)
		for _, v := range body.Keys {
			if store[v] == nil {
				store[v] = &Logs{}
			}
			res[v] = store[v].Committed
		}

		return n.Reply(msg, map[string]any{
			"type":    "list_committed_offsets_ok",
			"offsets": res,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

type Logs struct {
	List      []int
	Committed int
}

type LogStore map[string]*Logs

func (s LogStore) CreateIfNotExist(Key string) {
	if s[Key] == nil {
		s[Key] = &Logs{}
	}
}

func (s LogStore) Save(key string, value int) int {
	s.CreateIfNotExist(key)
	s[key].List = append(s[key].List, value)

	return len(s[key].List) - 1
}

func (s LogStore) Poll(key string, offset int) [][2]int {
	s.CreateIfNotExist(key)
	res := [][2]int{}

	for i := offset; i < len(s[key].List); i++ {
		res = append(res, [2]int{i, s[key].List[i]})
	}

	return res
}

type SendMsg struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	Msg  int    `json:"msg"`
}

type PollCommitMsg struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type PollRes struct {
	Type string              `json:"type"`
	Msgs map[string][][2]int `json:"msgs"`
}

type CommittedMsg struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}
