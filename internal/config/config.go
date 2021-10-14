package config

import (
	"github.com/kovetskiy/ko"
	"gopkg.in/yaml.v2"
)

type Database struct {
	Name     string `yaml:"name" required:"true" env:"DATABASE_NAME"`
	Host     string `yaml:"host" required:"true" env:"DATABASE_HOST"`
	Port     string `yaml:"port" required:"true" env:"DATABASE_PORT"`
	User     string `yaml:"user" required:"true"`
	Password string `yaml:"password" required:"true"`
}

type Telegram struct {
	Token string `yaml:"token" required:"true" env:"TELEGRAM_TOKEN"`
}

type BetApi struct {
	Token                   string `yaml:"token" required:"true"`
	BaseUrlUpcomingEvents   string `yaml:"base_url_upcoming_events" required:"true"`
	BaseUrlGetEventOddsById string `yaml:"base_url_get_event_odds_by_id" required:"true"`
}

type Handler struct {
	ApiVersion string `yaml:"api_version" required:"true"`
	Port       string `yaml:"port" required:"true"`
}

type Config struct {
	Database Database `yaml:"database" required:"true"`
	Telegram Telegram `yaml:"telegram" required:"true"`
	BetApi   BetApi   `yaml:"bet_api" required:"true"`
	Handler  Handler  `yaml:"handler" required:"true"`
}

func Load(path string) (*Config, error) {
	config := &Config{}
	err := ko.Load(path, config, ko.RequireFile(false), yaml.Unmarshal)
	if err != nil {
		return nil, err
	}

	return config, nil
}
