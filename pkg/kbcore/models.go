package kbcore

type Document struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Content   string         `gorm:"not null" json:"content"`
	Metadata  map[string]any `gorm:"serializer:json" json:"metadata"`
	CreatedAt string         `json:"created_at"`
}

type Chunk struct {
	ID         uint      `gorm:"primaryKey" json:"-"`
	DocumentID uint      `gorm:"index" json:"document_id"`
	Index      int       `json:"index"`
	Content    string    `gorm:"not null" json:"content"`
	Embedding  []float64 `gorm:"serializer:json" json:"-"`
}
