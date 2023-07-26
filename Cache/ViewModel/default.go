package ViewModel

import (
	"DuckBox/conf"
	"fmt"
	"github.com/gomodule/redigo/redis"
	subway "github.com/simonks2016/Subway"
)

var Pool *redis.Pool

func init() {
	Pool = subway.NewRedisConnWithSubway(fmt.Sprintf("%s:%s", conf.AppConfig.Redis.Host, conf.AppConfig.Redis.Port),
		conf.AppConfig.Redis.Account,
		conf.AppConfig.Redis.Password,
	)
}
