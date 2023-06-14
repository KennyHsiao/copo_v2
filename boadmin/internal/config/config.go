package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Host           string
	Server         string
	ChannelHost    string
	FrontEndDomain string
	rest.RestConf
	service.ServiceConf
	RpcService zrpc.RpcClientConf
	Auth       struct {
		AccessSecret string
		AccessExpire int64
	}
	Mysql struct {
		Host       string
		Port       int
		DBName     string
		UserName   string
		Password   string
		DebugLevel string
	}
	RedisCache struct {
		RedisSentinelNode string
		RedisMasterName   string
		RedisDB           int
	}
	ApiKey struct {
		PublicKey string
		PayKey    string
		ProxyKey  string
		LineKey   string
	}
	ResourceHost string
	Target       string

	Smtp struct {
		Host     string
		Port     int
		User     string
		Password string
	}

	Bucket struct {
		Host            string
		Name            string
		AccessKeyId     string
		AccessKeySecret string
	}

	LineSend struct {
		Host string
		Port int
	}
	Version string

	OtpRpc         zrpc.RpcClientConf
	TransactionRpc zrpc.RpcClientConf
}
