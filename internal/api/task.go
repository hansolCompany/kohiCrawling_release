package api

import "sync"

type taskRunner struct {
	mu      sync.Mutex
	running map[string]bool
}

func newTaskRunner() *taskRunner {
	return &taskRunner{
		running: make(map[string]bool),
	}
}

func (t *taskRunner) tryStart(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running[key] {
		return false
	}
	t.running[key] = true
	return true
}

func (t *taskRunner) finish(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.running, key)
}
