package rate

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/stretchr/testify/assert"
)

func Test_Heartbeat(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:8379",
	})
	_ = rdb.FlushDB(context.Background()).Err()

	now := time.Now().UnixNano()
	windowSize := int64(time.Second)

	returned := limiterHeartbeat.Eval(
		context.Background(),
		rdb,
		[]string{"test_1", "process_1"},
		now,
		windowSize,
	)
	actual, err := returned.Result()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), actual, "they should be equal")
}

func Test_Heartbeat_two_processes(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:8379",
	})
	_ = rdb.FlushDB(context.Background()).Err()

	now := time.Now().UnixNano()
	windowSize := int64(time.Second)

	limiterHeartbeat.Eval(
		context.Background(),
		rdb,
		[]string{"test_2", "process_1"},
		now,
		windowSize,
	)

	returned := limiterHeartbeat.Eval(
		context.Background(),
		rdb,
		[]string{"test_2", "process_2"},
		now,
		windowSize,
	)

	actual, err := returned.Result()
	assert.Nil(t, err)
	assert.Equal(t, int64(2), actual, "they should be equal")
}

func Test_Heartbeat_test_expire(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:8379",
	})
	_ = rdb.FlushDB(context.Background()).Err()

	now := time.Now().UnixNano()
	windowSize := int64(time.Second)
	assert.Nil(t, windowSize)

	limiterHeartbeat.Eval(
		context.Background(),
		rdb,
		[]string{"test_2", "process_1"},
		now,
		windowSize,
	)

	time.Sleep(2 * time.Second)

	now = time.Now().UnixNano()
	returned := limiterHeartbeat.Eval(
		context.Background(),
		rdb,
		[]string{"test_2", "process_2"},
		now,
		windowSize,
	)

	actual, err := returned.Result()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), actual, "they should be equal")
}
