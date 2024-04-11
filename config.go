package main

/*
Описание конфигурационного файла
*/

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/log"
)

type settings struct {
	Host          string
	Port          string
	ConLiveSec    int    `toml:"con_live_sec"`
	LogLevel      string `toml:"log_level"`
	Dbhost        string
	Dbport        uint16
	Dbuser        string
	Dbpass        string
	Database      string
	LogRecvPacket bool   `toml:"logRecvPacket"`
	LogDecPacket  bool   `toml:"logDecPacket"`
	ZabbixIp      string `toml:"zabbixIp"`
	ZabbixPort    string `toml:"zabbixPort"`
}

func (c *settings) Load(confPath string) error {
	if _, err := toml.DecodeFile(confPath, c); err != nil {
		//return fmt.Errorf("Ошибка разбора файла настроек: %v", err)
		return err
	}

	return nil
}

func (c *settings) getZabbixHost() string {
	return fmt.Sprintf("%v:%v", c.ZabbixIp, c.ZabbixPort)
}

func (c *settings) getConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable", c.Dbuser, c.Dbpass, c.Dbhost, c.Dbport, c.Database)
}

func (c *settings) getListenAddress() string {
	return c.Host + ":" + c.Port
}

func (c *settings) getLogLevel() log.Lvl {
	var lvl log.Lvl

	switch c.LogLevel {
	case "DEBUG":
		lvl = log.DEBUG
	case "INFO":
		lvl = log.INFO
	case "WARN":
		lvl = log.WARN
	case "ERROR":
		lvl = log.ERROR
	default:
		lvl = log.INFO
	}
	return lvl
}

func (c *settings) getEmptyConnTTL() time.Duration {
	return time.Duration(c.ConLiveSec) * time.Second
}
