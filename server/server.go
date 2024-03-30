package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/judwhite/go-svc"
	"github.com/unrolled/secure"
	"github.com/xochat/xochat_im_server_lib/config"
	"github.com/xochat/xochat_im_server_lib/module"
	"github.com/xochat/xochat_im_server_lib/pkg/log"
	"github.com/xochat/xochat_im_server_lib/pkg/register"
	"github.com/xochat/xochat_im_server_lib/pkg/xohttp"
)

// Server唐僧叨叨server
type Server struct {
	r *xohttp.XOHttp
	log.TLog
	sslAddr  string
	addr     string
	grpcAddr string
	ctx      *config.Context
}

// New 创建悟空聊天server
func New(ctx *config.Context) *Server {
	r := xohttp.New()
	r.Use(xohttp.CORSMiddleware())
	s := &Server{
		ctx:      ctx,
		r:        r,
		addr:     ctx.GetConfig().Addr,
		sslAddr:  ctx.GetConfig().SSLAddr,
		grpcAddr: ctx.GetConfig().GRPCAddr,
	}
	return s
}

func (s *Server) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		dir := filepath.Dir(os.Args[0])
		return os.Chdir(dir)
	}
	return nil
}

// Run 运行
func (s *Server) run(sslAddr string, addr ...string) error {
	// s.r.LoadHTMLGlob("assets/webroot/**/*.html")
	s.r.Static("/web", "./assets/web")
	s.r.Any("/v1/ping", func(c *xohttp.Context) {
		c.ResponseOK()
	})

	s.r.Any("/swagger/:module", func(c *xohttp.Context) {
		m := c.Param("module")
		module := register.GetModuleByName(m, s.ctx)
		if strings.TrimSpace(module.Swagger) == "" {
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, module.Swagger)
	})

	if len(addr) != 0 {
		if sslAddr != "" {
			go func() {
				err := s.r.Run(addr...)
				if err != nil {
					panic(err)
				}
			}()
		} else {
			err := s.r.Run(addr...)
			if err != nil {
				return err
			}
		}
	}

	// https 服务
	if sslAddr != "" {
		s.r.Use(TlsHandler(sslAddr))
		currDir, _ := os.Getwd()
		return s.r.RunTLS(sslAddr, currDir+"/assets/ssl/ssl.pem", currDir+"/assets/ssl/ssl.key")
	}
	return nil
}

func (s *Server) Start() error {
	go func() {
		err := s.run(s.sslAddr, s.addr)
		if err != nil {
			panic(err)
		}
	}()

	err := module.Start(s.ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {

	return module.Stop(s.ctx)
}

func TlsHandler(sslAddr string) xohttp.HandlerFunc {
	return func(c *xohttp.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     sslAddr,
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}

// GetRoute 获取路由
func (s *Server) GetRoute() *xohttp.XOHttp {
	return s.r
}
