package taskcore

type Task struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Priority string `json:"priority"` // "high" | "medium" | "low"
	Done     bool   `json:"done"`
}
