/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * failed_messages.go
 */

package hatchet

import "sync"

type FailedMessages struct {
	mu       sync.Mutex
	counters map[string]int
}

func (c *FailedMessages) inc(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := c.counters[name]
	c.counters[name] = n + 1
}
