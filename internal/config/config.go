package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	Redis  RedisConfig  `mapstructure:"redis"`
	NATS   NATSConfig   `mapstructure:"nats"`
	SMTP   SMTPConfig   `mapstructure:"smtp"`
	Worker WorkerConfig `mapstructure:"worker"`
}

type ServerConfig struct {
	Port           int    `mapstructure:"port"`
	JWTSecret      string `mapstructure:"jwt_secret"`
	JWTExpireHours int    `mapstructure:"jwt_expire_hours"`
}

type DBConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	DBName         string `mapstructure:"dbname"`
	SSLMode        string `mapstructure:"sslmode"`
	MaxOpenConns   int    `mapstructure:"max_open_conns"`    // 0 = unlimited
	MaxIdleConns   int    `mapstructure:"max_idle_conns"`    // 0 = Go default (2)
	ConnMaxIdleSec int    `mapstructure:"conn_max_idle_sec"` // 0 = no limit
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type NATSConfig struct {
	URL           string `mapstructure:"url"`
	MemoryStorage bool   `mapstructure:"memory_storage"` // true = 内存存储，吞吐更高但重启丢消息
	Replicas      int    `mapstructure:"replicas"`       // JetStream 副本数，单节点填 1（默认）
}

// WorkerConfig controls which NATS subjects this worker process subscribes to.
// Defaults to ["task.>"] (all task types).
// To run a specialised worker (e.g. GPU node for video only), set:
//
//	worker:
//	  subjects:
//	    - "task.video.*"
type WorkerConfig struct {
	Subjects []string `mapstructure:"subjects"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
