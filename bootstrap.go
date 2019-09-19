package main

import (
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
	"tcp-multiplier/config"
	"tcp-multiplier/server"
	"time"
)

func main() {
	defer func() {
		err := recover()
		if nil != err {
			log.Println(err)
		}
	}()

	if "" == strings.Trim(config.DestSvrAddrs, " ") {
		log.Println("you actually did not specify a valid \"DestSvrAddrs\",it is virtually empty")
		//flag.Usage()
		//os.Exit(0)
	}

	log.Println("mode ", config.Mode)

	// zap.NewDevelopment 格式化输出
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	logger.Info("无法获取网址",
		zap.String("url", "http://www.baidu.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)

	// verify mode
	switch config.Mode {
	case config.TCP_MODE:
		server.ListenAndServeTcp()
	case config.UDP_MODE:
		server.ServeUdp()
	default:
		log.Println("mode must be tcp or udp")
		os.Exit(0)
	}

}
