package models

import (
    "time"
)

// User 模型
type User struct {
    ID         []byte    `gorm:"type:binary(16);not null;primaryKey"`
    Username   string    `gorm:"type:varchar(255)"`
    Password   []byte    `gorm:"type:varbinary(32)"`
    InsertedAt time.Time `gorm:"not null"`
    UpdatedAt  time.Time `gorm:"not null"`
}

