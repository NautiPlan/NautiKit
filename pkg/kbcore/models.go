package kbcore

type Document struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Content   string         `gorm:"not null" json:"content"`
	Summary   string         `json:"summary"`
	Embedding []float64      `gorm:"serializer:json" json:"-"`
	Metadata  map[string]any `gorm:"serializer:json" json:"metadata"`
	CreatedAt string         `json:"created_at"`
}
