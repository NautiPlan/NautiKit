package taskcore

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(path string) error {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户目录失败: %w", err)
		}
		path = filepath.Join(home, ".nautikit", "data.db")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	var err error
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := db.AutoMigrate(&Task{}, &Plan{}); err != nil {
		return fmt.Errorf("自动建表失败: %w", err)
	}

	return nil
}

func Close() error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}
	return sqlDB.Close()
}

func AddTask(t Task) Task {
	db.Create(&t)
	return t
}

func ListTasks(planID uint) []Task {
	var out []Task
	q := db.Model(&Task{})
	if planID != 0 {
		q = q.Where("plan_id = ?", planID)
	}
	q.Order("date ASC, priority DESC").Find(&out)
	return out
}

func AddPlan(p Plan) Plan {
	p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	db.Create(&p)
	return p
}

func GetPlan(id uint) (Plan, error) {
	var p Plan
	if err := db.First(&p, id).Error; err != nil {
		return Plan{}, err
	}
	return p, nil
}

func ListPlans() []Plan {
	var out []Plan
	db.Order("created_at DESC").Find(&out)
	return out
}
