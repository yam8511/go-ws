package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv" // ----> 手動載入檔案
	// _ "github.com/joho/godotenv/autoload" ----> 打開註解的話，會自動載入 .env
	"gopkg.in/telegram-bot-api.v4"
)

// LoadEnv 載入環境變數
func LoadEnv() {
	/// 設定 Command Line 參數
	envFile := flag.String("e", "", "指定 env 檔案名稱，或者手動設定 APP_ENV 環境變數")
	flag.Parse()

	if *envFile == "" {
		return
	}

	/// 讀取 ENV 設定檔
	var err interface{}
	err = godotenv.Load(*envFile)
	if err != nil {
		log.Fatalf("[ERROR] 載入環境變數錯誤： %v", err)
	}
}

// LoadConfig 載入 config
func LoadConfig(env string) *Config {
	env = strings.TrimSpace(env)
	if env == "" {
		log.Println("[ERROR] 載入Config錯誤： env 不能為空！")
		log.Fatalf("[ERROR] 請看 ./melonSocket -h")
	}
	var configBody *Config
	configFile := GetAppRoot() + "/config/" + env + "_config.toml"
	if _, err := toml.DecodeFile(configFile, &configBody); err != nil {
		log.Fatalf("[ERROR] 載入Config錯誤： %v", err)
	}
	return configBody
}

// SetupBot 設定機器人
func SetupBot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		WriteLog("WARNING", fmt.Sprintf("建立通知機器人錯誤: %v\n", err))
		return
	}
	Bot = bot
}
