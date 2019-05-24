package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		redisClient = nil
	}
}

// GetCachedToken use redis to fetch an existing token, redis will not return a
// expired token
func GetCachedToken(clusterID, roleARN string) *token.Token {
	if redisClient != nil {
		key := _cacheKey(clusterID, roleARN)
		cacheResult, err := redisClient.Get(key).Result()
		if err != nil {
			log.Println("error reading cache", err)
			return nil
		}
		cachedToken := &token.Token{}
		if err := json.Unmarshal([]byte(cacheResult), cachedToken); err != nil {
			return nil
		}
		return cachedToken
	}
	return nil
}

// CacheToken cache a token if redis is available, set the expiration to the
// expiration on the generated token
func CacheToken(clusterID, roleARN string, t token.Token) {
	if redisClient != nil {
		key := _cacheKey(clusterID, roleARN)
		data, err := json.Marshal(t)
		if err != nil {
			return
		}
		_, err = redisClient.Set(key, data, time.Until(t.Expiration)).Result()
		if err != nil {
			log.Println("error setting cache", err)
		}
	}
}

func _cacheKey(clusterID, roleARN string) string {
	return fmt.Sprintf("awsiamauth::%s:%s", clusterID, roleARN)
}
