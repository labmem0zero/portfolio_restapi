package service

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"time"
	"wbEmployeeApi/model"
)

type (
	Config struct {
		AppName string 					`toml:"appName"`
		AppVer string                 `toml:"appVersion"`
		Servers map[string]server     `toml:"servers"`
		Databases map[string]database `toml:"databases"`
	}

	server struct {
		Ip string		`toml:"ip"`
		Port string		`toml:"port"`
	}

	database struct {
		Host string		`toml:"host"`
		Port string		`toml:"port"`
		DBName string	`toml:"dbName"`
		User string		`toml:"user,omitempty"`
		Password string	`toml:"password,omitempty"`
	}
)

func (t *Config) GetDBFromConfig(dbname string) model.DBConfig{
	var tmpconf model.DBConfig
	tmpconf.Host=t.Databases[dbname].Host
	tmpconf.Port=t.Databases[dbname].Port
	tmpconf.DBName=t.Databases[dbname].DBName
	tmpconf.User=t.Databases[dbname].User
	tmpconf.Password=t.Databases[dbname].Password
	return tmpconf
}

func ConfigInitialize(cfgPath string) *Config {
	var t *Config
	_,err:=toml.DecodeFile(cfgPath,&t)
	if err!=nil{
		log.Fatalln(time.Now(),":Невозможно запарсить файл конфигурации:",err)
	}
	for dbname,dbopts:=range t.Databases{
		if dbopts.User==""&&dbopts.Password==""{
			tmpDBCfg:=t.Databases[dbname]
			tmpDBCfg.User=os.Getenv("user_"+dbname)
			tmpDBCfg.Password=os.Getenv("pass_"+dbname)
			t.Databases[dbname]=tmpDBCfg
		}
	}
	return t
}
