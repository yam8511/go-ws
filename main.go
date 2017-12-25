package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 載入環境變數
	LoadEnv()
	// 載入設定檔資料
	Conf = LoadConfig(GetEnv("APP_ENV"))

	// 設置 Telegram 機器人
	SetupBot(Conf.Bot.Token)

	// 建立 Redis 連線
	redisConn, err := CreateRedisConnection(Conf.Redis.Publish.IP, Conf.Redis.Publish.Port)
	if err != nil {
		WriteLog("ERROR", fmt.Sprintf("ListenRedisPublish 錯誤: 建立Redis連線失敗 [%v]", err))
		return
	}

	// 設置 router
	router := SetupRouter()
	router.GET("/socket", StartServerPush(&redisConn))

	// 測試用廣播，正式使用時會註解掉
	router.GET("/publish/:message", func(c *gin.Context) {
		// 建立 Redis 連線
		pubConn, err := CreateRedisConnection(Conf.Redis.Publish.IP, Conf.Redis.Publish.Port)
		if err != nil {
			errMsg := fmt.Sprintf("RedisPublish 錯誤: 建立Redis連線失敗 [%v]", err)
			WriteLog("ERROR", errMsg)
			c.String(http.StatusOK, errMsg)
			return
		}
		msg := c.Param("message")
		_, err = pubConn.Do("PUBLISH", "melon_member", msg)
		if err != nil {
			errMsg := fmt.Sprintf("RedisPublish 廣播失敗: %s [%v]", msg, err)
			WriteLog("ERROR", errMsg)
			c.String(http.StatusOK, errMsg)
			return
		}
		c.String(http.StatusOK, msg)
	})

	// 測試用的頁面，正式使用時會註解掉
	router.LoadHTMLGlob(GetAppRoot() + "/public/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 建立 server
	NotifyEngineer("Server Push 開始服務!")
	defer NotifyEngineer("Server Push 結束服務!")
	server := CreateServer(router, Conf.App.Port, Conf.App.Host)

	// 系統信號監聽
	SignalListenAndServe(server)
}
