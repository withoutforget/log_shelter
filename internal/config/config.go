package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type LoggerConfig struct {
	DeveloperMode bool `toml:"developer_mode"`
}

type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	Driver   string `toml:"driver"`
}

func (p *PostgresConfig) Dsn() string {
	url := "host=%v port=%v dbname=%v user=%v password=%v sslmode=disable"
	return fmt.Sprintf(url, p.Host, p.Port, p.Database, p.Username, p.Password)
}

type TelegramConfig struct {
	Enabled      bool     `toml:"enabled"`
	APIKey       string   `toml:"api_key"`
	Levels       []string `toml:"levels"`
	NotificateTo []int64  `toml:"notificate_to"`
}

type NatsConfig struct {
	Url      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type LogConfig struct {
	RetencionPolicy string   `toml:"retencion_policy"`
	DeleteAfter     Duration `toml:"delete_after"`
	CycleTime       Duration `toml:"cycle_time"`
}

type Config struct {
	Logger   LoggerConfig   `toml:"logger"`
	Postgres PostgresConfig `toml:"postgres"`
	Nats     NatsConfig     `toml:"nats"`
	Telegram TelegramConfig `toml:"telegram"`
	Logs     LogConfig
}

func readConfigFile(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	return data
}

func GetConfig(filename string) *Config {
	var cfg Config
	err := toml.Unmarshal(readConfigFile(filename), &cfg)
	if err != nil {
		panic(err.Error())
	}
	return &cfg
}
