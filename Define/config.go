package Define

import (
	"fmt"
	"strings"
)

type AppConfig struct {
	NSQ         NSQConfig   `json:"nsq" yaml:"nsq"`
	MeiliSearch MeiliSearch `json:"meili_search" yaml:"meili_search"`
	Mysql       Mysql       `json:"mysql" yaml:"mysql"`
	Redis       Redis       `json:"redis" yaml:"redis"`
	Gorse       Gorse       `json:"gorse" yaml:"gorse"`
}

type NSQConfig struct {
	Address   string `json:"address" yaml:"address"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`
	Port      string `json:"port" yaml:"port"`
}

type Mysql struct {
	DB       string `json:"db" yaml:"db"`
	Account  string `json:"account" yaml:"account"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
}

type Redis struct {
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	Account  string `json:"account" yaml:"account"`
	Password string `json:"password" yaml:"password"`
}

type MeiliSearch struct {
	Address string `json:"address" yaml:"address"`
	ApiKey  string `json:"api_key" yaml:"api_key"`
	Port    string `json:"port" yaml:"port"`
}

type Microservices struct {
	Address string `json:"address" yaml:"address"`
	API     struct {
		Key string `json:"key" yaml:"key"`
	} `json:"api" yaml:"api"`
}

type Gorse struct {
	Host   string `json:"host" yaml:"host"`
	ApiKey string `json:"api_key" yaml:"api_key"`
	Port   string `json:"port" yaml:"port"`
}

func (this *MeiliSearch) ToHost() string {

	if IsHttpsURL(this.Address) {
		if len(this.Port) > 0 && strings.Compare(this.Port, "443") != 0 {
			return this.Address + ":" + this.Port
		}
		return this.Address
	}

	if this.Port == "443" {
		return "https://" + this.Address
	}
	return "http://" + this.Address + ":" + this.Port
}

func (this *NSQConfig) ToHost() string {

	if len(this.Port) <= 0 {
		return this.Address
	}
	return fmt.Sprintf("%s:%s", this.Address, this.Port)
}

func (g *Gorse) ToEndPoint() string {

	var endPoint string
	if !IsHttpsURL(g.Host) {
		if strings.Compare(g.Port, "443") == 0 {
			endPoint = "https://" + g.Host
		} else {
			endPoint = "http://" + g.Host
		}
	}
	if len(g.Port) <= 0 {
		return endPoint
	} else {
		return endPoint + ":" + g.Port
	}
}
