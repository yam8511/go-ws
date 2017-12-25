package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

// GetEnv 取環境變數
func GetEnv(key string) string {
	return os.Getenv(key)
}

// GetAppRoot 取專案的根目錄
func GetAppRoot() string {
	root, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		WriteLog("WARNING", "GetAppRoot：取根目錄失敗，自動抓取 APP_ROOT 的環境變數")
		return GetEnv("APP_ROOT")
	}
	return root
}

// WriteLog 寫Log記錄檔案
func WriteLog(tag string, msg string) {
	//設定時間
	now := time.Now()

	// 組合字串
	logStr := now.Format("[2006-01-02 15:04:05]") + "【" + tag + "】" + msg + "\n"
	log.Print(logStr)

	// 設定檔案位置
	fileName := "melon-server.log"
	folderPath := GetAppRoot() + now.Format("/logs/2006-01-02/15/")

	//檢查今日log檔案是否存在
	if _, err := os.Stat(folderPath + fileName); os.IsNotExist(err) {
		//建立資料夾
		os.MkdirAll(folderPath, 0777)
		//建立檔案
		_, err := os.Create(folderPath + fileName)
		if err != nil {
			log.Printf("WriteLog: 建立檔案錯誤 [%v] \n----> %s\n", err, msg)
			return
		}
	}

	//開啟檔案準備寫入
	logFile, err := os.OpenFile(folderPath+fileName, os.O_RDWR|os.O_APPEND, 0777)
	defer logFile.Close()
	if err != nil {
		log.Printf("WriteLog: 開啟檔案錯誤 [%v]\n----> %s\n", err, msg)
		return
	}

	_, err = logFile.WriteString(logStr)

	if err != nil {
		log.Printf("WriteLog: 寫入檔案錯誤 [%v] \n----> %s\n", err, msg)
	}
}

// NotifyEngineer 通知工程師
func NotifyEngineer(msg string) {
	if Bot == nil {
		WriteLog("WARNING", fmt.Sprintf("通知訊息失敗： 機器人尚未設定 [%s]\n", msg))
		return
	}

	message := tgbotapi.NewMessage(Conf.Bot.ChatID, msg)
	_, err := Bot.Send(message)

	if err != nil {
		WriteLog("WARNING", fmt.Sprintf("通知訊息發送失敗： %v\n", err))
	}
}
