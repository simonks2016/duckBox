package controllers

import "github.com/astaxie/beego/logs"

const (
	LogDebug     = 0
	LogInfo      = 1
	LogWarning   = 2
	LogError     = 3
	LogAlert     = 4
	LogEmergency = 5

	MeiliSearchHost   = "http://localhost:7700"
	MeiliSearchAPIKey = "7Ow5P0k-gu6Ss40iPehILYCyKZQI8ACjRfm5NZh9C48"

	MeiliSearchIndexVideo   = "video"
	MeiliSearchIndexProgram = "program"
)

func Log(name, msg string, level int) {

	log := logs.NewLogger(10000)
	_ = log.SetLogger("file", `{"filename":"default.log"}`)
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(3)
	switch level {
	case 0:
		//调试信息
		log.Debug("%s:%s", name, msg)
	case 1:
		//输出信息
		log.Informational("%s,%s", name, msg)
	case 2:
		//警告等级
		log.Warning("%s:%s", name, msg)
	case 3:
		//错误等级
		log.Error("%s:%s", name, msg)
	case 4:
		//提醒信息
		log.Alert("%s:%s", name, msg)
	case 5:
		//紧急信息
		log.Emergency("%s:%s", name, msg)
	}
}
