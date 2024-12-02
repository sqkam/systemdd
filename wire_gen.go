// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/sqkam/systemdx/ioc"
)

// Injectors from wire.go:

func InitConfig() *ioc.ServerConfig {
	v := InitChan()
	serverConfig := ioc.InitConfig(v)
	return serverConfig
}

// wire.go:

func InitChan() chan struct{} {
	ListenConfigChan = make(chan struct{})
	return ListenConfigChan

}
