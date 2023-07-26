package Define

type AppConfig struct {
	NSQ         NSQConfig   `json:"nsq" yaml:"nsq"`
	MeiliSearch MeiliSearch `json:"meili_search" yaml:"meili_search"`
	Mysql       Mysql       `json:"mysql" yaml:"mysql"`
	Redis       Redis       `json:"redis" yaml:"redis"`
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

func (this *MeiliSearch) ToHost() string {

	if this.Port == "" || this.Port == "443" {
		return "https://" + this.Address
	}
	return "http://" + this.Address + ":" + this.Port
}