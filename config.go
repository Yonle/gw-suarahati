package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"os"
)

var config struct {
	Token  string `yaml:"BOT_TOKEN"`
	Chat_ID int64  `yaml:"CHAT_ID"`

	Mastodon_Host_Url   string `yaml:"MASTODON_HOST_URL"`
	Mastodon_Access_Key string `yaml:"MASTODON_ACCESS_KEY"`
	Mastodon_User_Url   string `yaml:"MASTODON_USER_URL"`
}

func ReadConfig(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			panic("Config file does not exist.")
		}

		panic(fmt.Sprintf("error when reading %s: %s", filename, err))
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(fmt.Sprintf("error when parsing %s: %s", filename, err))
	}
}
