package db

import (
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewSqlite 创建一个sqlite db，[path]db存储路径 [sqlDir]sql脚本目录
func NewSqlite(filepath string, sqlDir string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(filepath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
