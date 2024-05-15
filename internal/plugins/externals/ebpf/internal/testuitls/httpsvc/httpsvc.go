package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/go-redis/redis/v8"
	"github.com/petermattis/goid"
	"golang.org/x/sys/unix"
)

var log = logger.DefaultSLogger("httpcli")

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "abc123", // no password set
		DB:       0,        // use default DB
	})
	r := rdb.Ping(ctx)
	log.Info(r.String())

	err := rdb.Set(ctx, "abc", "123", 0).Err()
	if err != nil {
		panic(err)
	}
	log.Infof("pid %d\n", unix.Getpid())

	l, _ := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.1.2"),
		Port: 61095,
	})
	_ = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		log.Info("goid ", goid.Get())
		val, err := rdb.Get(context.Background(), "abc").Result()
		if err != nil {
			log.Infof("val %v, err: %w", val, err)
		} else {
			log.Infof("val %v", val)
		}
		_, _ = w.Write([]byte("OK"))
	}))

	log.Infof("server url %s\n", l.Addr().String())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, unix.SIGINT, unix.SIGTERM)
	<-sig
}
