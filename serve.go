package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter() *gin.Engine {
	var router *gin.Engine

	if GetEnv("GIN_MODE") == "debug" {
		router = gin.Default()
		// log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	return router
}

// CreateServer 建立伺服器
func CreateServer(router *gin.Engine, port, host string, args ...string) *http.Server {
	// 建立 Server
	server := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	return server
}

// SignalListenAndServe 開啟Server & 系統信號監聽
func SignalListenAndServe(server *http.Server) {
	defer func() {
		if err := recover(); err != nil {
			errMessage := fmt.Sprintf("SignalListenAndServe Error: %v", err)
			WriteLog("ERROR", errMessage)
			NotifyEngineer(errMessage)
		}
	}()

	// 啟動 Server
	go func() {
		defer func() {
			if err := recover(); err != nil {
				errMessage := fmt.Sprintf("Server Recover Error: %v", err)
				WriteLog("ERROR", errMessage)
				NotifyEngineer(errMessage)
			}
		}()

		WriteLog("INFO", "Server 開始服務！ 【"+Conf.App.Env+"】 "+Conf.App.Host+Conf.App.Port)
		server.ListenAndServe()
	}()

	/// 宣告系統信號
	sigs := make(chan os.Signal, 1)
	exit := make(chan interface{})
	signal.Notify(
		sigs,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGSYS,
		syscall.SIGABRT,
	)

	/// 設置監聽系統訊號機制
	go func() {
		// 等待系統關閉的信號
		receivedSignel := <-sigs

		// 關閉伺服器
		server.Close()

		// 離開程式
		exit <- receivedSignel
	}()

	/// 等待信號，結束程式前的最後一個任務
	signalMessage := fmt.Sprintf("Server 接受信號: %v\n", <-exit)
	WriteLog("INFO", signalMessage)
	WriteLog("INFO", "Server 結束服務")
}
