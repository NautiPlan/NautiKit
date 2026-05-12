package taskcore

import (
	"fmt"
	"sync"
)

var (
	store = make([]Task, 0)
	mu    sync.RWMutex
	idSeq int
)

func AddTask(t Task) Task {
	mu.Lock()
	defer mu.Unlock()
	idSeq++
	t.ID = fmt.Sprintf("task-%d", idSeq)
	store = append(store, t)
	return t
}

func ListTasks() []Task {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Task, len(store))
	copy(out, store)
	return out
}
