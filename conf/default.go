package conf

import (
	"DuckBox/Define"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var AppConfig *Define.AppConfig

func init() {

	cd, err := LoadConfig()
	if err != nil {
		panic(err.Error())
	}
	AppConfig = cd
}

func LoadConfig() (*Define.AppConfig, error) {

	file, err := ioutil.ReadFile("./conf/app.yaml")
	if err != nil {
		return nil, err
	}

	var config Define.AppConfig

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
