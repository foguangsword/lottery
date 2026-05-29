package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Logger     LoggerConfig
	Blockchain BlockchainConfig
	Jwt        JwtConfig
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`
}

type LoggerConfig struct {
	Level    string `mapstructure:"level"`
	Filename string `mapstructure:"filename"`
}

type BlockchainConfig struct {
	RPCURL  string `mapstructure:"rpc_url"`
	ChainID int64  `mapstructure:"chain_id"`
}

type JwtConfig struct {
	Secret string `mapstructure:"secret"`
}

func Load() (*Config, error) {
	// 加载 .env 文件（开发环境）
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 用环境变量覆写敏感信息
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.Database.Password = dbPass
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.Jwt.Secret = secret
	}

	return &cfg, nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName, d.Charset)
}
