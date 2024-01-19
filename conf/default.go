package conf

import (
	"DuckBox/Define"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
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

	var cf *ConfigFile

	err = yaml.Unmarshal(file, &cf)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(cf.Mode) == "dev" {
		return cf.Dev, nil
	}
	return cf.Prod, nil
}

type ConfigFile struct {
	Mode string            `json:"mode" yaml:"mode"`
	Dev  *Define.AppConfig `json:"dev" yaml:"dev"`
	Prod *Define.AppConfig `json:"prod" yaml:"prod"`
}
