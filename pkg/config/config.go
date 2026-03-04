package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var Config *Configuration

type Configuration struct {
	Server   ServerConfig   `mapstructure:"server"`
	UserService    ServiceConfig  `mapstructure:"user_service"`
	ContentService ServiceConfig  `mapstructure:"content_service"`
	MessageService ServiceConfig  `mapstructure:"message_service"`
	ChatbotService ServiceConfig  `mapstructure:"chatbot_service"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	Etcd     EtcdConfig     `mapstructure:"etcd"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Microservices MicroservicesConfig `mapstructure:"microservices"`
	LogLevel string         `mapstructure:"log_level"`
	LLM      LLMConfig      `mapstructure:"llm"`
}

type LLMConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
}

type MicroservicesConfig struct {
	User    string `mapstructure:"user"`
	Content string `mapstructure:"content"`
	Message string `mapstructure:"message"`
	Chatbot string `mapstructure:"chatbot"`
}

type ServiceConfig struct {
	Port string `mapstructure:"port"`
}

type ServerConfig struct {
	RunMode      string `mapstructure:"run_mode"`
	HttpPort     string `mapstructure:"http_port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessExpire  int64  `mapstructure:"access_expire"`
	RefreshExpire int64  `mapstructure:"refresh_expire"`
}

func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	
	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TICKTOK")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
