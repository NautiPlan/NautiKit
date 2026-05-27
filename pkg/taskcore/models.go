package taskcore

type Task struct {
	ID          string `json:"id"`
	PlanID      string `json:"plan_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`     // "2026-05-28"
	Priority    string `json:"priority"` // "high" | "medium" | "low"
	Done        bool   `json:"done"`
}

type Plan struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}
