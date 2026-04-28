// @title           Project1_Shop API
// @version         1.0
// @description     一个基于 Gin + GORM + Redis + RabbitMQ 的在线书店后端服务，提供用户认证、书籍管理、购物车、订单及秒杀等功能。
// @termsOfService  http://swagger.io/terms/

// @contact.name   Me
// @contact.email  frottenice@outlook.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:9090
// @BasePath  /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 格式：Bearer {token}
package main

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/logger"
	"Project1_Shop/pkg/Worker"
	"Project1_Shop/pkg/mq"
	"Project1_Shop/router"
	"Project1_Shop/settings"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// 加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	// 初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success...")
	// 初始化 MySQL 连接
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close()
	mysql.AutoMigration()
	// 初始化 Redis 连接
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()
	// 初始化 RabbitMQ（声明队列拓扑 + 保持长连接）
	if err := mq.Init(settings.Conf.RabbitMQConfig); err != nil {
		fmt.Printf("init RabbitMQ failed, err:%v\n", err)
		return
	}
	defer mq.Close()
	// 启动消费者
	if err := mq.StartOrderExpiredConsumer(); err != nil {
		fmt.Printf("start order expired consumer failed, err:%v\n", err)
		return
	}
	if err := mq.StartOrderPaymentConsumer(); err != nil {
		fmt.Printf("start order payment consumer failed, err:%v\n", err)
		return
	}
	if err := mq.StartSeckillConsumer(); err != nil {
		fmt.Printf("start seckill consumer failed, err:%v\n", err)
		return
	}
	// 注册路由
	r := router.SetUp()
	// 启动工作池
	go Worker.StartRateWorker(ctx)
	// 启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutdown Server ...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Fatal("Server Shutdown: ", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
