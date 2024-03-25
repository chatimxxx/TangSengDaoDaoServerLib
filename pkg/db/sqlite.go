package db

import (
	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path"
)

// NewSqlite 创建一个sqlite db，[path]db存储路径 [sqlDir]sql脚本目录
func NewSqlite(filepath string, sqlDir string) (*gorm.DB, error) {
	err := os.Mkdir(path.Dir(filepath), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	migrations := &migrate.FileMigrationSource{
		Dir: sqlDir,
	}
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	s, err := db.DB()
	if err != nil {
		return nil, err
	}
	_, err = migrate.Exec(s, "sqlite3", migrations, migrate.Up)
	if err != nil {
		return nil, err
	}
	return db, nil
}
