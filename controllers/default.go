package controllers

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

const (
	LogDebug     = 0
	LogInfo      = 1
	LogWarning   = 2
	LogError     = 3
	LogAlert     = 4
	LogEmergency = 5

	MeiliSearchIndexVideo   = "video"
	MeiliSearchIndexProgram = "program"
)

var (
	logging *logrus.Logger
)

func init() {
	fileName := fmt.Sprintf("log/default.log-%s", time.Now().Format("2006-01-02"))
	var file *os.File
	var err error

	//log 文件夹是否存在
	if Exists("log") == false {
		//不存在则创建log文件夹
		err = os.Mkdir("log", os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//创建文件
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	} else {
		//该文件是否存在
		if Exists(fileName) == false {
			//假如不存在，则创建文件
			file, err = os.Create(fileName)
			if err != nil {
				return
			}
		} else {
			//存在则打开文件
			file, err = os.OpenFile(
				fileName, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}

	logging = logrus.New()

	logging.SetReportCaller(true)
	logging.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp:  false,
		DisableHTMLEscape: true,
	})
	logging.SetOutput(file)
}

func Log(name, msg string, level int) {

	switch level {
	case 0:
		//调试信息
		logging.Debugf("%s:%s", name, msg)
	case 1:
		//输出信息
		logging.Infof("%s,%s", name, msg)
	case 2:
		//警告等级
		logging.Warnf("%s:%s", name, msg)
	case 3:
		//错误等级
		logging.Errorf("%s:%s", name, msg)
	default:
		logging.Infof("%s:%s", name, msg)
	}
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
