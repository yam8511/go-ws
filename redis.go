package main

import (
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
)

// CreateRedisConnection 建立Redis連線
func CreateRedisConnection(ip, port string) (redis.Conn, error) {
	redisConn, err := redis.Dial("tcp", ip+port)
	if err != nil {
		logStr := fmt.Sprintf("Redis 連線失敗 > %s [%v]", ip+":"+port, err)
		WriteLog("WARNING", logStr)
		return nil, err
	}
	return redisConn, nil
}

// ListenRedisPublish 監聽Redis的廣播
func ListenRedisPublish(redisConn *redis.Conn, clientMap map[string]map[int]chan []byte) {
	defer NotifyEngineer("Server Push Redis監聽停止了!")

	WriteLog("INFO", "開始監聽Redis的廣播!")

	psc := redis.PubSubConn{Conn: *redisConn}
	err := psc.PSubscribe(Conf.Redis.Publish.Channel)
	if err != nil {
		WriteLog("ERROR", fmt.Sprint("Redis 監聽廣播錯誤: ", err))
		log.Fatal()
	}
	for {
		switch v := psc.Receive().(type) {
		case redis.PMessage:
			for UID := range clientMap {
				for STime := range clientMap[UID] {
					clientMap[UID][STime] <- v.Data
				}
			}
		default:
			continue
		}
	}
}
