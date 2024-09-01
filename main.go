package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type ICache interface {
	Cap() int
	Len() int
	Clear() // удаляет все ключи
	Add(key, value any)
	AddWithTTL(key, value any, ttl time.Duration) // добавляет ключ со сроком жизни ttl
	Get(key any) (value any, ok bool)
	Remove(key any)
}

type LRUCache struct {
	capacity int
	cache    map[any]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type element struct {
	key   any
	value any
	et    time.Time
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		capacity: cap,
		cache:    make(map[any]*list.Element, cap),
		list:     list.New(),
	}
}

func (c *LRUCache) Cap() int {
	return c.capacity
}

func (c *LRUCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list.Len()
}

func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[any]*list.Element)
	c.list.Init()
}

func main() {
	cache := NewLRUCache(3)

	fmt.Println(cache.Cap())
	fmt.Println(cache.Len())
}
