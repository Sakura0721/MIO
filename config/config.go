package config

import (
	"io/ioutil"

	"github.com/gotd/td/tg"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AppSid       string `yaml:"app_sid"`
	ApiId        int    `yaml:"api_id"`
	ApiHash      string `yaml:"api_hash"`
	Uid          string `yaml:"uid"`
	Phone        string `yaml:"phone"`
	User         *tg.User
	GiveHearts   GiveHeartsConfig   `yaml:"give_hearts"`
	ReplyHearts  ReplyHeartsConfig  `yaml:"reply_hearts"`
	AddTime      AddTimeConfig      `yaml:"add_time"`
	AddTimeAll   AddTimeAllConfig   `yaml:"add_time_all"`
	ReplyAddTime ReplyAddTimeConfig `yaml:"reply_add_time"`
	Explore      ExploreConfig      `yaml:"explore"`
}

type GiveHeartsConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Usernames []string `yaml:"usernames"`
}

type ReplyHeartsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type AddTimeAllConfig struct {
	Enabled             bool     `yaml:"enabled"`
	ExcludeUsernames    []string `yaml:"exclude_usernames"`
	ExcludeUsernamesMap map[string]struct{}
}

type AddTimeConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Usernames []string `yaml:"usernames"`
}

type ReplyAddTimeConfig struct {
	Enabled bool `yaml:"enabled"`
}

type ExploreConfig struct {
	Enabled bool `yaml:"enabled"`
}

var C Config

func init() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	// Unmarshal the YAML data into the struct
	err = yaml.Unmarshal(data, &C)
	if err != nil {
		panic(err)
	}
	C.AddTimeAll.ExcludeUsernamesMap = make(map[string]struct{})
	for _, username := range C.AddTimeAll.ExcludeUsernames {
		C.AddTimeAll.ExcludeUsernamesMap[username] = struct{}{}
	}
}
