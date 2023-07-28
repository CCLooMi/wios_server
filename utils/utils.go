package utils

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/CCLooMi/sql-mak/mysql/entity"
)

// 生成随机ID
func GenerateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func UUID() *entity.ID {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	id := entity.ID(b)
	return &id
}
