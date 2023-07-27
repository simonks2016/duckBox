package ViewModel

import (
	"DuckBox/conf"
	"github.com/gomodule/redigo/redis"
	subway "github.com/simonks2016/Subway"
)

var Pool *redis.Pool

func init() {
	Pool = subway.NewRedisConnWithSubway(
		conf.AppConfig.Redis.Host+":"+conf.AppConfig.Redis.Port,
		conf.AppConfig.Redis.Account,
		conf.AppConfig.Redis.Password,
	)
}
