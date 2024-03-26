package module

import (
	"fmt"
	"gorm.io/gorm"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/chatimxxx/TangSengDaoDaoServerLib/config"
	"github.com/chatimxxx/TangSengDaoDaoServerLib/pkg/register"
	migrate "github.com/rubenv/sql-migrate"
)

func Setup(ctx *config.Context, initSql bool) error {
	// 获取所有模块
	ms := register.GetModules(ctx)
	// 初始化SQL
	if initSql {
		var sqlFSs []*register.SqlFS
		for _, m := range ms {
			if m.SQLDir != nil {
				sqlFSs = append(sqlFSs, m.SQLDir)
			}
		}
		db, err := ctx.DB()
		if err != nil {
			return err
		}
		err = executeSQL(sqlFSs, db)
		if err != nil {
			return err
		}
	}
	// 注册api
	for _, m := range ms {
		if m.SetupAPI != nil {
			a := m.SetupAPI()
			if a != nil {
				a.Route(ctx.GetHttpRoute())
			}
		}
		if ctx.SetupTask {
			if m.SetupTask != nil {
				t := m.SetupTask()
				if t != nil {
					t.RegisterTasks()
				}
			}
		}
	}
	return nil
}

func Start(ctx *config.Context) error {
	// 获取所有模块
	ms := register.GetModules(ctx)
	for _, m := range ms {
		if m.Start != nil {
			err := m.Start()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func Stop(ctx *config.Context) error {
	// 获取所有模块
	ms := register.GetModules(ctx)
	for _, m := range ms {
		if m.Stop != nil {
			err := m.Stop()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 执行sql
func executeSQL(sqlFss []*register.SqlFS, db *gorm.DB) error {
	migrations := &FileDirMigrationSource{
		sqlFss: sqlFss,
	}
	s, err := db.DB()
	if err != nil {
		return err
	}
	_, fms, err := migrations.FindMigrations()
	if err != nil {
		return err
	}
	for _, fm := range fms {
		_, err := migrate.Exec(s, "mysql", fm, migrate.Up)
		if err != nil {
			return err
		}
	}
	return nil
}

type byID []*migrate.Migration

func (b byID) Len() int           { return len(b) }
func (b byID) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byID) Less(i, j int) bool { return b[i].Less(b[j]) }

// FileDirMigrationSource 文件目录源 遇到目录进行递归获取
type FileDirMigrationSource struct {
	sqlFss []*register.SqlFS
}

// FindMigrations FindMigrations
func (f FileDirMigrationSource) FindMigrations() ([]*migrate.Migration, []*migrate.FileMigrationSource, error) {
	if len(f.sqlFss) == 0 {
		return nil, nil, nil
	}
	migrations := make([]*migrate.Migration, 0, 100)
	fmss := make([]*migrate.FileMigrationSource, 0)
	for _, sqlFs := range f.sqlFss {
		fms, err := f.findMigrations(sqlFs, &migrations)
		if err != nil {
			return nil, nil, err
		}
		fmss = append(fmss, fms...)
	}
	// Make sure migrations are sorted
	sort.Sort(byID(migrations))
	return migrations, fmss, nil
}

func (f FileDirMigrationSource) findMigrations(fs *register.SqlFS, migrations *[]*migrate.Migration) ([]*migrate.FileMigrationSource, error) {
	files, err := fs.ReadDir("sql")
	if err != nil {
		return nil, err
	}
	fms := make([]*migrate.FileMigrationSource, 0)
	for _, info := range files {
		if strings.HasSuffix(info.Name(), ".sql") {
			file, err := fs.Open(path.Join("sql", info.Name()))
			if err != nil {
				return nil, fmt.Errorf("error while opening %s: %s", info.Name(), err)
			}
			migration, err := migrate.ParseMigration(info.Name(), file.(io.ReadSeeker))
			if err != nil {
				return nil, fmt.Errorf("error while parsing %s: %s", info.Name(), err)
			}
			*migrations = append(*migrations, migration)
			fm := &migrate.FileMigrationSource{
				Dir: path.Join("sql", info.Name()),
			}
			fms = append(fms, fm)
		}
	}
	return fms, nil
}
