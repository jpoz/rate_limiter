package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	rate "github.com/jpoz/redis-rate"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:8379"})

	globalRate := 2500
	processID := fmt.Sprintf("pid: %d", os.Getpid())

	limiter := rate.NewDistributedLimitier(rdb, context.Background(), "basic", processID, globalRate)
	var ops uint64

	go func() {
		for {
			if limiter.Allow() {
				atomic.AddUint64(&ops, 1)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			fmt.Printf("ops = %+v, processes = %+v\n", ops, limiter.CurrentProcesses)
			atomic.SwapUint64(&ops, 0)
			select {
			case <-ticker.C:
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
