package main

import (
	"context"
	"fmt"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Store struct {
	kv    *maelstrom.KV
	mu    *sync.RWMutex
	store map[string][]int
}

func NewStore(node *maelstrom.Node) Store {
	return Store{
		kv:    maelstrom.NewLinKV(node),
		store: make(map[string][]int),
		mu:    &sync.RWMutex{},
	}
}

func (s *Store) Update(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	last := len(s.store[key])
	n, err := s.kv.ReadInt(context.Background(), fmt.Sprintf("%s-NextOffset", key))
	if err != nil {
		s.kv.CompareAndSwap(context.Background(), fmt.Sprintf("%s-NextOffset", key), 0, 0, true)
		return
	}
	for i := last; i < n; i++ {
		keyOffset := fmt.Sprintf("%s-%d", key, i)
		x, err := s.kv.ReadInt(context.Background(), keyOffset)
		if err != nil {
			break
		}
		s.store[key] = append(s.store[key], x)
	}
}

func (s *Store) Save(key string, value int) int {
	var NewOffset int
	keyNextOffset := fmt.Sprintf("%s-NextOffset", key)
	for {
		NewOffset, _ = s.kv.ReadInt(context.Background(), keyNextOffset)
		err := s.kv.CompareAndSwap(context.Background(), keyNextOffset, NewOffset, NewOffset+1, true)
		if err == nil {
			break
		}
	}

	err := s.kv.Write(context.Background(), fmt.Sprintf("%s-%d", key, NewOffset), value)
	if err != nil {
		panic(fmt.Sprint("something happened while saving: ", err))
	}

	return NewOffset
}

func (s *Store) Poll(key string, offset int) [][2]int {
	s.Update(key)
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := [][2]int{}
	for i := offset; i < len(s.store[key]); i++ {
		res = append(res, [2]int{i, s.store[key][i]})
	}

	return res
}

func (s *Store) Commit(key string, offset int) {
	realKey := fmt.Sprintf("%s-committed", key)
	for {
		value, err := s.kv.ReadInt(context.Background(), realKey)
		if err != nil {
			value = 0
			s.kv.Write(context.Background(), realKey, value)
		}
		err = s.kv.CompareAndSwap(context.Background(), realKey, value, offset, false)
		if err == nil {
			break
		}
	}
}

func (s *Store) Committed(key string) int {
	realKey := fmt.Sprintf("%s-committed", key)
	value, err := s.kv.ReadInt(context.Background(), realKey)
	if err != nil {
		value = 0
		s.kv.Write(context.Background(), realKey, value)
	}

	return value
}
