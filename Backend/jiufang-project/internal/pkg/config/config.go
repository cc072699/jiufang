package config

import (
	"fmt"
	"net"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig
	CORS         CORSConfig
	Database     DatabaseConfig
	JWT          JWTConfig
	LLM          LLMConfig
	ERP          ERPConfig
	Redis        RedisConfig
	QueryService QueryServiceConfig
	Dialog       DialogConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type CORSConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	ExpireTime time.Duration
}

type LLMConfig struct {
	Provider      string  `yaml:"provider"`
	Model         string  `yaml:"model"`
	APIKey        string  `yaml:"api_key"`
	Endpoint      string  `yaml:"endpoint"`
	Temperature   float64 `yaml:"temperature"`
	MaxTokens     int     `yaml:"max_tokens"`
	Timeout       int     `yaml:"timeout"`
	RetryCount    int     `yaml:"retry_count"`
	FallbackModel string  `yaml:"fallback_model"`
}

type ERPConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Database        string        `yaml:"database"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	QueryTimeout    time.Duration `yaml:"query_timeout"`
	MaxResultRows   int           `yaml:"max_result_rows"`
}

type RedisConfig struct {
	Address      string `yaml:"address"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
	MaxRetries   int    `yaml:"max_retries"`
	DialTimeout  int    `yaml:"dial_timeout"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type QueryServiceConfig struct {
	MaxHistory   int `yaml:"max_history"`
	QueryTimeout int `yaml:"query_timeout"`
}

type DialogConfig struct {
	ContextTTL     int  `yaml:"context_ttl"`
	MaxTurns       int  `yaml:"max_turns"`
	EnableAnaphora bool `yaml:"enable_anaphora"`
	EnableMerge    bool `yaml:"enable_merge"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/jiufang")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("cors.allow_origins", []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:3000",
	})
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "jiufang")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.secret", "jiufang-default-secret-key")
	viper.SetDefault("jwt.expire_time", "24h")
	viper.SetDefault("llm.provider", "deepseek")
	viper.SetDefault("llm.model", "deepseek-chat")
	viper.SetDefault("llm.temperature", 0.7)
	viper.SetDefault("llm.max_tokens", 4096)
	viper.SetDefault("llm.timeout", 30)
	viper.SetDefault("llm.retry_count", 3)
	viper.SetDefault("erp.driver", "mysql")
	viper.SetDefault("erp.host", "localhost")
	viper.SetDefault("erp.port", 3306)
	viper.SetDefault("erp.max_open_conns", 10)
	viper.SetDefault("erp.max_idle_conns", 5)
	viper.SetDefault("erp.conn_max_lifetime", 30)
	viper.SetDefault("erp.query_timeout", 30)
	viper.SetDefault("erp.max_result_rows", 10000)
	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 20)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", 5)
	viper.SetDefault("redis.read_timeout", 3)
	viper.SetDefault("redis.write_timeout", 3)
	viper.SetDefault("query_service.max_history", 10)
	viper.SetDefault("query_service.query_timeout", 30)
	viper.SetDefault("dialog.context_ttl", 30)
	viper.SetDefault("dialog.max_turns", 5)
	viper.SetDefault("dialog.enable_anaphora", true)
	viper.SetDefault("dialog.enable_merge", true)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	expireTime, err := time.ParseDuration(viper.GetString("jwt.expire_time"))
	if err != nil {
		expireTime = 24 * time.Hour
	}

	return &Config{
		Server: ServerConfig{
			Port: viper.GetInt("server.port"),
			Mode: viper.GetString("server.mode"),
		},
		CORS: CORSConfig{
			AllowOrigins: appendLocalIPIfNeeded(viper.GetStringSlice("cors.allow_origins")),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		JWT: JWTConfig{
			Secret:     viper.GetString("jwt.secret"),
			ExpireTime: expireTime,
		},
		LLM: LLMConfig{
			Provider:      viper.GetString("llm.provider"),
			Model:         viper.GetString("llm.model"),
			APIKey:        viper.GetString("llm.api_key"),
			Endpoint:      viper.GetString("llm.endpoint"),
			Temperature:   viper.GetFloat64("llm.temperature"),
			MaxTokens:     viper.GetInt("llm.max_tokens"),
			Timeout:       viper.GetInt("llm.timeout"),
			RetryCount:    viper.GetInt("llm.retry_count"),
			FallbackModel: viper.GetString("llm.fallback_model"),
		},
		ERP: ERPConfig{
			Driver:          viper.GetString("erp.driver"),
			Host:            viper.GetString("erp.host"),
			Port:            viper.GetInt("erp.port"),
			Database:        viper.GetString("erp.database"),
			Username:        viper.GetString("erp.username"),
			Password:        viper.GetString("erp.password"),
			MaxOpenConns:    viper.GetInt("erp.max_open_conns"),
			MaxIdleConns:    viper.GetInt("erp.max_idle_conns"),
			ConnMaxLifetime: time.Duration(viper.GetInt("erp.conn_max_lifetime")) * time.Minute,
			QueryTimeout:    time.Duration(viper.GetInt("erp.query_timeout")) * time.Second,
			MaxResultRows:   viper.GetInt("erp.max_result_rows"),
		},
		Redis: RedisConfig{
			Address:      viper.GetString("redis.address"),
			Password:     viper.GetString("redis.password"),
			DB:           viper.GetInt("redis.db"),
			PoolSize:     viper.GetInt("redis.pool_size"),
			MinIdleConns: viper.GetInt("redis.min_idle_conns"),
			MaxRetries:   viper.GetInt("redis.max_retries"),
			DialTimeout:  viper.GetInt("redis.dial_timeout"),
			ReadTimeout:  viper.GetInt("redis.read_timeout"),
			WriteTimeout: viper.GetInt("redis.write_timeout"),
		},
		QueryService: QueryServiceConfig{
			MaxHistory:   viper.GetInt("query_service.max_history"),
			QueryTimeout: viper.GetInt("query_service.query_timeout"),
		},
		Dialog: DialogConfig{
			ContextTTL:     viper.GetInt("dialog.context_ttl"),
			MaxTurns:       viper.GetInt("dialog.max_turns"),
			EnableAnaphora: viper.GetBool("dialog.enable_anaphora"),
			EnableMerge:    viper.GetBool("dialog.enable_merge"),
		},
	}, nil
}

func (c *JWTConfig) GetExpireTime() time.Time {
	return time.Now().Add(c.ExpireTime)
}

func appendLocalIPIfNeeded(origins []string) []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return origins
	}
	ports := []string{"5173", "3000"}
	existing := make(map[string]bool, len(origins))
	for _, o := range origins {
		existing[o] = true
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}
		ip := ipNet.IP.String()
		for _, port := range ports {
			origin := fmt.Sprintf("http://%s:%s", ip, port)
			if !existing[origin] {
				origins = append(origins, origin)
			}
		}
	}
	return origins
}
