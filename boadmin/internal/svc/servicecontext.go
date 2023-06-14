package svc

import (
	"com.copo/bo_service/boadmin/internal/config"
	"fmt"
	"github.com/copo888/copo_otp/rpc/otpclient"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/gioco-play/go-driver/logrusz"
	"github.com/gioco-play/go-driver/mysqlz"
	"github.com/go-redis/redis/v8"
	_ "github.com/neccoys/go-zero-extension/consul"
	"github.com/zeromicro/go-zero/zrpc"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"strings"
	"sync"
)

type ServiceContext struct {
	Config         config.Config
	RedisClient    *redis.Client
	MyDB           *gorm.DB
	RpcServices    sync.Map
	MailService    *gomail.Dialer
	OtpRpc         otpclient.Otp
	TransactionRpc transactionclient.Transaction
}

//
//func (s *ServiceContext) RpcService(channel string) zrpc.Client {
//
//	rpc, ok := s.RpcServices.Load(channel)
//
//	if !ok {
//		ch := strings.Replace(s.Config.Target, "@", channel, 1)
//		client, err := zrpc.NewClientWithTarget(ch)
//
//		if err != nil {
//			log.Panicln("Consul Error:", err)
//		}
//
//		return client
//	}
//
//	return rpc.(zrpc.Client)
//
//}

//func (s *ServiceContext) GetRpcService(rpcName string) *grpc.ClientConn {
//
//	rpc, ok := s.RpcServices.Load(rpcName)
//
//	if !ok {
//		fmt.Println("init ", rpcName, s.Config.RpcService.Etcd.Hosts)
//
//		client := zrpc.MustNewClient(zrpc.RpcClientConf{
//			Etcd: discov.EtcdConf{
//				Hosts: s.Config.RpcService.Etcd.Hosts,
//				Key:   rpcName,
//			},
//		})
//
//		conn := client.Conn()
//
//		s.RpcServices.Store(rpcName, conn)
//
//		return conn
//	}
//
//	return rpc.(*grpc.ClientConn)
//}

func NewServiceContext(c config.Config) *ServiceContext {

	// Redis
	redisCache := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    c.RedisCache.RedisMasterName,
		SentinelAddrs: strings.Split(c.RedisCache.RedisSentinelNode, ";"),
		DB:            c.RedisCache.RedisDB,
	})

	// DB
	db, err := mysqlz.New(c.Mysql.Host, fmt.Sprintf("%d", c.Mysql.Port), c.Mysql.UserName, c.Mysql.Password, c.Mysql.DBName).
		SetCharset("utf8mb4").
		SetLoc("UTC").
		SetLogger(logrusz.New().SetLevel(c.Mysql.DebugLevel).Writer()).
		Connect(mysqlz.Pool(50, 100, 180))

	if err != nil {
		panic(err)
	}

	// Tracer
	//ztrace.StartAgent(ztrace.Config{
	//	Name:     c.Telemetry.Name,
	//	Endpoint: c.Telemetry.Endpoint,
	//	Batcher:  c.Telemetry.Batcher,
	//	Sampler:  c.Telemetry.Sampler,
	//})

	return &ServiceContext{
		Config:         c,
		RedisClient:    redisCache,
		MyDB:           db,
		MailService:    gomail.NewDialer(c.Smtp.Host, c.Smtp.Port, c.Smtp.User, "&iNasw28"),
		TransactionRpc: transactionclient.NewTransaction(zrpc.MustNewClient(c.TransactionRpc)),
		OtpRpc:         otpclient.NewOtp(zrpc.MustNewClient(c.OtpRpc)),
	}
}
