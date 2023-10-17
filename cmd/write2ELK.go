package main

import (
	"context"
	"github.com/bshuster-repo/logrus-logstash-hook" // github.com/bshuster-repo/logrus-logstash-hook v0.4.1
	"github.com/sirupsen/logrus"                    // github.com/sirupsen/logrus v1.9.3
	"log"
	"time"
)

func GetContextLogger(ctx context.Context) (x *logrus.Logger) {
	k := "logger"
	v, ok := ctx.Value(k).(*logrus.Logger)
	if !ok {
		return nil
	}
	return v
}
func SetContextLogger(ctx context.Context, v interface{}) (x context.Context) {
	k := "logger"
	return context.WithValue(ctx, k, v)
}

func main() {
	logger := logrus.New()
	logServer := "192.168.36.6:5000"
	logger.SetLevel(logrus.TraceLevel) // 后续改为 配置中心处理
	//logger.SetFormatter(&logrus.TextFormatter{
	//	FullTimestamp: true,
	//})
	hook, err := logrustash.NewHook("tcp", logServer, "serviceName002")
	if err != nil {
		log.Fatal(err)
	}
	logger.Hooks.Add(hook)

	go ABC(SetContextLogger(context.Background(), logger))

	// 业务逻辑
	log := logger.WithFields(logrus.Fields{
		"method": "methodMain",
	})
	for true {
		s := time.Now().String()

		log.Error(" Error " + s)
		time.Sleep(time.Second)
		log.Warn(" Warn " + s)
		time.Sleep(time.Second)
		log.Info(" Info " + s)
		time.Sleep(time.Second)
		log.Debug(" Debug " + s)
		time.Sleep(time.Second)
		log.Error("{ 'foo' : 'bar' }")
		time.Sleep(time.Second)
	}
}

func ABC(ctx context.Context) {
	logger := GetContextLogger(ctx)
	log := logger.WithFields(logrus.Fields{
		"method": "methodABC",
	})

	for true {
		s := time.Now().String()

		log.Error(" Error " + s)
		time.Sleep(time.Second)
		log.Warn(" Warn " + s)
		time.Sleep(time.Second)
		log.Info(" Info " + s)
		time.Sleep(time.Second)
		log.Debug(" Debug " + s)
		time.Sleep(time.Second)
		log.Error("{ 'foo' : 'bar' }")
		time.Sleep(time.Second)
	}
}
