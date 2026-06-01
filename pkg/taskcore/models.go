package taskcore

type Task struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	PlanID      uint   `gorm:"index" json:"plan_id"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	Date        string `gorm:"index" json:"date"`              // "2026-05-28"
	Priority    string `gorm:"default:medium" json:"priority"` // high / medium / low
	Done        bool   `json:"done"`
}

type Plan struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"` // (RFC3339)
}
