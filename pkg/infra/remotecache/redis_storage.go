package remotecache

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util/errutil"
	redis "gopkg.in/redis.v2"
)

const redisCacheType = "redis"

type redisStorage struct {
	c *redis.Client
}

// parseRedisConnStr parses k=v pairs in csv and builds a redis Options object
func parseRedisConnStr(connStr string) (*redis.Options, error) {
	keyValueCSV := strings.Split(connStr, ",")
	options := &redis.Options{Network: "tcp"}
	for _, rawKeyValue := range keyValueCSV {
		keyValueTuple := strings.SplitN(rawKeyValue, "=", 2)
		if len(keyValueTuple) != 2 {
			if strings.HasPrefix(rawKeyValue, "password") {
				// don't log the password
				rawKeyValue = "password******"
			}
			return nil, fmt.Errorf("检测到 '%v' 的redis连接字符串格式不正确 , 格式为 key=value,key=value", rawKeyValue)
		}
		connKey := keyValueTuple[0]
		connVal := keyValueTuple[1]
		switch connKey {
		case "addr":
			options.Addr = connVal
		case "password":
			options.Password = connVal
		case "db":
			i, err := strconv.ParseInt(connVal, 10, 64)
			if err != nil {
				return nil, errutil.Wrap("redis连接字符串中db的值必须是数字", err)
			}
			options.DB = i
		case "pool_size":
			i, err := strconv.Atoi(connVal)
			if err != nil {
				return nil, errutil.Wrap("redis连接字符串中pool_size的值必须是数字", err)
			}
			options.PoolSize = i
		default:
			return nil, fmt.Errorf("redis连接字符串中无法识别的选项 '%v' ", connVal)
		}
	}
	return options, nil
}

func newRedisStorage(opts *setting.RemoteCacheOptions) (*redisStorage, error) {
	opt, err := parseRedisConnStr(opts.ConnStr)
	if err != nil {
		return nil, err
	}
	return &redisStorage{c: redis.NewClient(opt)}, nil
}

// Set sets value to given key in session.
func (s *redisStorage) Set(key string, val interface{}, expires time.Duration) error {
	item := &cachedItem{Val: val}
	value, err := encodeGob(item)
	if err != nil {
		return err
	}
	status := s.c.SetEx(key, expires, string(value))
	return status.Err()
}

// Get gets value by given key in session.
func (s *redisStorage) Get(key string) (interface{}, error) {
	v := s.c.Get(key)

	item := &cachedItem{}
	err := decodeGob([]byte(v.Val()), item)

	if err == nil {
		return item.Val, nil
	}
	if err.Error() == "EOF" {
		return nil, ErrCacheItemNotFound
	}
	return nil, err
}

// Delete delete a key from session.
func (s *redisStorage) Delete(key string) error {
	cmd := s.c.Del(key)
	return cmd.Err()
}
