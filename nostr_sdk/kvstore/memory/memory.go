package memory

import (
	"sync"

	"github.com/jerry-harm/nosmec/nostr_sdk/kvstore"
)

var _ kvstore.KVStore = (*Store)(nil)

type Store struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string][]byte),
	}
}

func (s *Store) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[string(key)], nil
}

func (s *Store) Set(key []byte, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[string(key)] = value
	return nil
}

func (s *Store) Delete(key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, string(key))
	return nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) Update(key []byte, f func([]byte) ([]byte, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	newVal, err := f(s.data[string(key)])
	if err != nil {
		return err
	}
	if newVal == nil {
		delete(s.data, string(key))
	} else {
		s.data[string(key)] = newVal
	}
	return nil
}

func (s *Store) Iterate(fn func(key, value []byte) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.data {
		if err := fn([]byte(k), v); err != nil {
			return err
		}
	}
	return nil
}