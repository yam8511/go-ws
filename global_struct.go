package main

import (
	"gopkg.in/telegram-bot-api.v4"
)

// Conf 全域設定變數
var Conf *Config

// Bot 通知機器人
var Bot *tgbotapi.BotAPI

// Config 型態
type Config struct {
	App struct {
		Env  string `toml:"env"`
		Host string `toml:"host"`
		Port string `toml:"port"`
	} `toml:"app"`
	Redis struct {
		Publish struct {
			IP      string `toml:"ip"`
			Host    string `toml:"host"`
			Port    string `toml:"port"`
			Channel string `toml:"channel"`
			MaxLink int64  `toml:"max_link"`
		} `toml:"publish"`
	} `toml:"redis"`
	Bot struct {
		Token  string `toml:"token"`
		ChatID int64  `toml:"chat_id"`
	} `toml:"bot"`
}

// UserSocket 型態
type UserSocket struct {
	UserID int64 `json:"user_id"`
}
