package rate

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

const windowSize = int64(time.Second)
const heartbeatInterval = 500 * time.Microsecond

type DistributedLimiter struct {
	ProcessRate      int
	CurrentProcesses int
	LastHeatbeat     time.Time

	rdb       *redis.Client
	ctx       context.Context
	key       string
	processID string
	rate      int
	limiter   *rate.Limiter
}

func NewDistributedLimitier(
	rdb *redis.Client,
	ctx context.Context,
	key string,
	processID string,
	rateS int,
) *DistributedLimiter {
	limiter := &DistributedLimiter{
		rdb:       rdb,
		ctx:       ctx,
		key:       key,
		processID: processID,
		rate:      rateS,
		limiter:   rate.NewLimiter(rate.Limit(rateS), rateS),
	}

	limiter.Checkin()
	limiter.Start()

	return limiter
}

func (dl *DistributedLimiter) Start() {
	ticker := time.NewTicker(heartbeatInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				dl.Checkin()
			case <-dl.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// Checkin will check with redis that the process is in the current list
// This shouldn't need to be called manually. When NewDistrivutedLimiter
// is called a goroutine will call this in the background
// Returned in the current process count
func (dl *DistributedLimiter) Checkin() {
	now := time.Now()

	cmd := limiterHeartbeat.Eval(
		dl.ctx,
		dl.rdb,
		[]string{dl.key, dl.processID},
		now.UnixNano(),
		windowSize,
	)

	// TODO make this safe / check error
	dl.CurrentProcesses = int(cmd.Val().(int64))
	newProcessRate := dl.rate / dl.CurrentProcesses
	if newProcessRate != dl.ProcessRate {
		dl.ProcessRate = newProcessRate
		dl.limiter.SetLimit(rate.Limit(newProcessRate))
	}

	dl.LastHeatbeat = now
}

func (dl *DistributedLimiter) Allow() bool {
	allowed := dl.limiter.Allow()
	// fmt.Printf("a = %+v\n", allowed)
	return allowed
}
