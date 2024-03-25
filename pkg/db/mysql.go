package db

import (
	"fmt"
	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"sort"
	"strings"
	"time"
)

// NewMySQL 创建一个MySQL db，[path]db存储路径 [sqlDir]sql脚本目录
func NewMySQL(addr string, maxOpenConns int, maxIdleConns int, connMaxLifetime time.Duration) *gorm.DB {
	db, err := gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func Migration(sqlDir string, db *gorm.DB) error {
	migrations := &FileDirMigrationSource{
		Dir: sqlDir,
	}
	s, err := db.DB()
	if err != nil {
		return err
	}
	_, err = migrate.Exec(s, "mysql", migrations, migrate.Up)
	if err != nil {
		return err
	}
	return nil
}

type byID []*migrate.Migration

func (b byID) Len() int           { return len(b) }
func (b byID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byID) Less(i, j int) bool { return b[i].Less(b[j]) }

// FileDirMigrationSource 文件目录源 遇到目录进行递归获取
type FileDirMigrationSource struct {
	Dir string
}

// FindMigrations FindMigrations
func (f FileDirMigrationSource) FindMigrations() ([]*migrate.Migration, error) {
	filesystem := http.Dir(f.Dir)
	migrations := make([]*migrate.Migration, 0, 100)
	err := f.findMigrations(filesystem, &migrations)
	if err != nil {
		return nil, err
	}
	// Make sure migrations are sorted
	sort.Sort(byID(migrations))

	return migrations, nil
}

func (f FileDirMigrationSource) findMigrations(dir http.FileSystem, migrations *[]*migrate.Migration) error {
	file, err := dir.Open("/")
	if err != nil {
		return err
	}

	files, err := file.Readdir(0)
	if err != nil {
		return err
	}

	for _, info := range files {

		if strings.HasSuffix(info.Name(), ".sql") {
			file, err := dir.Open(info.Name())
			if err != nil {
				return fmt.Errorf("Error while opening %s: %s", info.Name(), err)
			}

			migration, err := migrate.ParseMigration(info.Name(), file)
			if err != nil {
				return fmt.Errorf("Error while parsing %s: %s", info.Name(), err)
			}
			*migrations = append(*migrations, migration)

		} else if info.IsDir() {
			err = f.findMigrations(http.Dir(fmt.Sprintf("%s/%s", f.Dir, info.Name())), migrations)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
