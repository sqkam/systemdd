//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/sqkam/systemdd/ioc"
)

func InitChan() chan struct{} {
	ListenChan = make(chan struct{})
	return ListenChan

}
func InitConfig() *ioc.ServerConfig {
	panic(wire.Build(InitChan, ioc.InitConfig))

}
