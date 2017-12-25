package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// StartServerPush 開始ServerPush
func StartServerPush(redisConn *redis.Conn) gin.HandlerFunc {
	WriteLog("INFO", "Websocket Start!")

	// clientMap 記住目前有哪些客戶在連 websocket
	clientMap := make(map[string]map[int]chan []byte)
	// mu 讀寫鎖
	mu := new(sync.RWMutex)

	// 開始監聽 Redis 廣播
	go ListenRedisPublish(redisConn, clientMap)

	upgrader := websocket.Upgrader{
		// 先允許所有的Origin都可以進來
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(c *gin.Context) {
		// 建立 websocket 連線
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		defer ws.Close()
		if err != nil {
			WriteLog("ERROR", fmt.Sprintf("建立 Websocket 失敗： %v", err))
			return
		}

		// 讀取 WS 訊息
		_, message, err := ws.ReadMessage()
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("ws.ReadMessager Error: %v", err))
			return
		}

		// 期待訊息 {"user_id": 123}
		wsData := UserSocket{}
		err = json.Unmarshal(message, &wsData)
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("Websocket 解析訊息錯誤: %v", err))
			return
		}

		clientKey := fmt.Sprintf("UID_%d", wsData.UserID)

		// 如果使用者第一次連
		if _, exists := clientMap[clientKey]; !exists {
			mu.Lock()
			clientMap[clientKey] = map[int]chan []byte{}
			mu.Unlock()
		}

		// 檢查連線數量
		mu.RLock()
		linkNum := int64(len(clientMap[clientKey]))
		mu.RUnlock()

		// 如果數量超出過上限，則斷掉最舊的連線
		if linkNum > Conf.Redis.Publish.MaxLink-1 {
			WriteLog("WARNING", fmt.Sprintf("Weboscket 連線數量超過上限 %d : UserID = %d", Conf.Redis.Publish.MaxLink, wsData.UserID))
			times := GetSortedMapKeys(clientMap[clientKey])
			mu.Lock()
			delete(clientMap[clientKey], times[0])
			mu.Unlock()
		}

		// 登記 WS
		clientTime := time.Now().Nanosecond()
		clientChan := make(chan []byte)
		mu.Lock()
		clientMap[clientKey][clientTime] = clientChan
		mu.Unlock()

		wg := new(sync.WaitGroup)
		wg.Add(1)
		wg.Add(1)

		// 持續檢查 socket 是否還連線中
		go CheckSocketClosed(ws, wg, clientChan)
		// 持續監聽並推送訊息到前端
		go WaitMessageToPush(ws, wg, clientChan)

		wg.Wait()

		mu.Lock()
		delete(clientMap[clientKey], clientTime)
		if len(clientMap[clientKey]) == 0 {
			delete(clientMap, clientKey)
		}
		mu.Unlock()
		return
	}
}

// CheckSocketClosed 確認socket連線有無關閉
func CheckSocketClosed(ws *websocket.Conn, wg *sync.WaitGroup, clientChan chan []byte) {
	for {
		_, _, err := ws.NextReader()
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("Websocket 連線已斷線 [%v]", err))
			close(clientChan)
			ws.Close()
			wg.Done()
			return
		}
	}
}

// WaitMessageToPush 等待訊息去推送
func WaitMessageToPush(ws *websocket.Conn, wg *sync.WaitGroup, clientChan chan []byte) {
	for {
		message := <-clientChan
		err := ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			wg.Done()
			if string(message) != "" {
				WriteLog("WARNING", fmt.Sprintf("Websocket 發送訊息失敗： %s [%v]", string(message), err))
			}
			return
		}
	}
}

// GetSortedMapKeys 取 Map 的 key值陣列並且排序過
func GetSortedMapKeys(data map[int]chan []byte) []int {
	keys := []int{}
	for key := range data {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}
