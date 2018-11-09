package main

import (
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// RedisHealthChecker defines redis health checker
type RedisHealthChecker struct {
	client *redis.Client
	logger *log.Logger
}

// NewRedisHealthChecker creates a new redis health checker
func NewRedisHealthChecker(addr string, password string, logger *log.Logger) *RedisHealthChecker {
	logger.Printf("checking redis: %s\n", addr)

	c := &RedisHealthChecker{
		logger: logger,
		client: redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DialTimeout:  1 * time.Second,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		}),
	}

	return c
}

// CheckHealth implements Checker interface
func (r *RedisHealthChecker) CheckHealth() error {
	r.logger.Println("Checking health")

	// check if its alive
	if err := r.client.Ping().Err(); err != nil {
		r.logger.Printf("Could not ping redis: %s\n", err)
		return errors.Errorf("Could not ping redis: %s", err)
	}

	// get the client info
	raw, err := r.client.Info().Result()
	if err != nil {
		r.logger.Printf("Could not get client info: %s\n", err)
		return errors.Errorf("Could not get client info: %s", err)
	}

	// parse out
	info := parseKeyValue(raw)

	// check if redis is loading data from disk
	if loading, ok := info["loading"]; ok && loading == "1" {
		return errors.New("Redis is currently loading data from disk")
	}

	// check if redis is syncing data from master
	if masterSyncInProgress, ok := info["master_sync_in_progress"]; ok && masterSyncInProgress == "1" {
		return errors.New("Redis slave is currently syncing data from its master instance")
	}

	// check if the master link is up
	if masterLinkUpString, ok := info["master_link_status"]; ok && masterLinkUpString != "up" {
		return errors.New("Redis slave do not have active connection to its master instance")
	}

	r.logger.Println("Redis healthy")

	return nil
}

func parseKeyValue(str string) map[string]string {
	res := make(map[string]string)

	lines := strings.Split(str, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		pair := strings.Split(line, ":")
		if len(pair) != 2 {
			continue
		}

		res[pair[0]] = pair[1]
	}

	return res
}
