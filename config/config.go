package config

type Email struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	To     string `json:"to"`
	Enable bool   `json:"enable"`
}

type Tg struct {
	Token  string `json:"token"`
	ChatID int64  `json:"chatID"`
	Enable bool   `json:"enable"`
}

type Node struct {
	Rpc          string `json:"rpc"`
	User         string `json:"user"`
	Pass         string `json:"pass"`
	Alert        Alert  `json:"alert"`
	Gap          int64  `json:"gap"`
	UseStateRoot bool   `json:"useStateRoot"`
}

type Alert struct {
	MaxAllowErrorTimes int64 `json:"maxAllowErrorTimes"`
	MaxBlockTime       int64 `json:"maxBlockTime"`
}

type Config struct {
	Email Email  `json:"email"`
	Tg    Tg     `json:"tg"`
	Nodes []Node `json:"nodes"`
}
