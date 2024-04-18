package db

import "github.com/xochat/xochat_im_server_lib/pkg/redis"

func NewRedis(addr string, password string, db int) *redis.Conn {
	return redis.New(addr, password, db)
}
