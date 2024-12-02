//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/sqkam/systemdx/ioc"
)

func InitChan() chan struct{} {
	ListenConfigChan = make(chan struct{})
	return ListenConfigChan

}
func InitConfig() *ioc.ServerConfig {
	panic(wire.Build(InitChan, ioc.InitConfig))

}
