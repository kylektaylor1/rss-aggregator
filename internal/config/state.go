package config

import "github.com/kylektaylor1/rss-aggregator/internal/database"

type State struct {
	Cfg *Config
	Db  *database.Queries
}
