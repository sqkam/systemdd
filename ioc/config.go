package ioc

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/sqkam/systemdd/color"
)

type ServerConfig struct {
	Units []*Unit `mapstructure:"units"`
}
type Unit struct {
	Exec    string `mapstructure:"exec"`
	WorkDir string `mapstructure:"work_dir"`
	Disable bool   `mapstructure:"disable"`
}

var path = flag.String("c", "./config.yaml", "配置文件路径")

func InitConfig(listenChan chan struct{}) *ServerConfig {

	var config ServerConfig
	v := viper.New()
	v.SetConfigFile(*path)
	if !flag.Parsed() {
		flag.Parse()
	}
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {

		if err := v.ReadInConfig(); err != nil {
			fmt.Printf("%s %s %s\n", color.Red, "read viper config failed:"+err.Error(), color.Reset)

			return
		}
		if err := v.Unmarshal(&config); err != nil {
			fmt.Printf("%s %s %s\n", color.Red, "read viper config failed:"+err.Error(), color.Reset)
			return
		}

		fmt.Printf("%s %s %s\n", color.Green, "update conf", color.Reset)

		listenChan <- struct{}{}
	})
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("read viper config failed: %s", err.Error())
		panic(err)
	}
	if err := v.Unmarshal(&config); err != nil {
		fmt.Printf("unmarshal err failed: %s", err.Error())
		panic(err)
	}

	return &config

}
