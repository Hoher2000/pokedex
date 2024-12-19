package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	data      []byte
}

type Cache struct {
	mu       sync.RWMutex
	cacheMap map[string]cacheEntry
}

func (c *Cache) String() string {
	res := ""
	for url := range c.cacheMap {
		if url != "" {
			res += url + " ,"
		}
	}
	return res
}

func (c *Cache) Add(url string, body []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Println("Adding data to cache at", time.Now())
	c.cacheMap[url] = cacheEntry{
		createdAt: time.Now(),
		data:      body,
	}
	//fmt.Println(c)
}

func (c *Cache) Get(url string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	//fmt.Println(c)
	//fmt.Println(url)
	if entry, ok := c.cacheMap[url]; ok {
		fmt.Println("Getting data from cache at", time.Now())
		//fmt.Println(c)
		return entry.data, ok
	}
	//fmt.Println("Not okay in get")
	return nil, false
}

func NewCache(d time.Duration) *Cache {
	c := Cache{
		mu:       sync.RWMutex{},
		cacheMap: map[string]cacheEntry{},
	}
	//fmt.Println("Starting reaploop at", time.Now())
	go c.reapLoop(d)
	return &c
}

func (c Cache) reapLoop(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		t := <-ticker.C
		for url, entry := range c.cacheMap {
			if t.Sub(entry.createdAt) > d {
				c.mu.Lock()
				delete(c.cacheMap, url)
				fmt.Println("deleting from cache", url)
				c.mu.Unlock()
			}
			//time.Sleep(d)
		}
	}
}
