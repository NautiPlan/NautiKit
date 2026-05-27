package taskcore

import (
	"fmt"
	"sync"
	"time"
)

var (
	store = make([]Task, 0)
	mu    sync.RWMutex
	idSeq int

	planStore = make([]Plan, 0)
	planMu    sync.RWMutex
	planIDSeq int
)

func AddTask(t Task) Task {
	mu.Lock()
	defer mu.Unlock()
	idSeq++
	t.ID = fmt.Sprintf("task-%d", idSeq)
	store = append(store, t)
	return t
}

func ListTasks(planID string) []Task {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Task, 0)
	for _, t := range store {
		if planID == "" || t.PlanID == planID {
			out = append(out, t)
		}
	}
	return out
}

func AddPlan(p Plan) Plan {
	planMu.Lock()
	defer planMu.Unlock()
	planIDSeq++
	p.ID = fmt.Sprintf("plan-%d", planIDSeq)
	p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	planStore = append(planStore, p)
	return p
}

func GetPlan(id string) (Plan, bool) {
	planMu.RLock()
	defer planMu.RUnlock()
	for _, p := range planStore {
		if p.ID == id {
			return p, true
		}
	}
	return Plan{}, false
}

func ListPlans() []Plan {
	planMu.RLock()
	defer planMu.RUnlock()
	out := make([]Plan, len(planStore))
	copy(out, planStore)
	return out
}
