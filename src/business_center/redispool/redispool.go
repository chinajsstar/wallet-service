package redispool

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	pool *redis.Pool = nil
)

func Get() redis.Conn {
	return pool.Get()
}

func init() {
	fmt.Println("init redispool")
	pool = newPool()
}

//初始化一个pool
func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			//if _, err := c.Do("AUTH", password); err != nil {
			//	c.Close()
			//	return nil, err
			//}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}
