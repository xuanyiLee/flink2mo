package conf

import (
	"gopkg.in/ini.v1"
)

var Cfg *Conf
var MyCnf MysqlConf
var MoConf MysqlConf

type Conf struct {
	Filename string
	Type     string
	Scale    string
}

type MysqlConf struct {
	HOST     string
	Port     int
	Username string
	Password string
	DataBase string
}

func NewConf(filename string) *Conf {
	Cfg = &Conf{Filename: filename}
	return Cfg
}

func (c *Conf) Load() error {
	cfg, err := ini.Load(c.Filename)
	if err != nil {
		return err
	}

	err = c.loadMysqlConf(cfg, "mysql")
	if err != nil {
		return err
	}

	err = c.loadMysqlConf(cfg, "matrixone")
	if err != nil {
		return err
	}

	return nil
}

func (c *Conf) loadMysqlConf(cfg *ini.File, dataSource string) error {
	port, err := cfg.Section(dataSource).Key("port").Int()
	if err != nil {
		return err
	}

	sourceConf := MysqlConf{
		HOST:     cfg.Section(dataSource).Key("host").String(),
		Port:     port,
		Username: cfg.Section(dataSource).Key("username").String(),
		Password: cfg.Section(dataSource).Key("password").String(),
		DataBase: cfg.Section(dataSource).Key("database").String(),
	}
	switch dataSource {
	case "mysql":
		MyCnf = sourceConf
	case "matrixone":
		MoConf = sourceConf
	}

	return nil
}
