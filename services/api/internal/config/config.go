package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	MySQL       MySQLConfig       `yaml:"mysql"`
	JWT         JWTConfig         `yaml:"jwt"`
	Crypto      CryptoConfig      `yaml:"crypto"`
	WeChat      WeChatConfig      `yaml:"wechat"`
	LLM         LLMConfig         `yaml:"llm"`
	MiniProgram MiniProgramConfig `yaml:"miniprogram"`
	Seed        SeedConfig        `yaml:"seed"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
	Mode string `yaml:"mode"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type JWTConfig struct {
	AdminSecret   string `yaml:"admin_secret"`
	UserSecret    string `yaml:"user_secret"`
	AdminTTLHours int    `yaml:"admin_ttl_hours"`
	UserTTLHours  int    `yaml:"user_ttl_hours"`
}

type CryptoConfig struct {
	AESKey string `yaml:"aes_key"`
}

type WeChatConfig struct {
	AppID        string `yaml:"app_id"`
	AppSecret    string `yaml:"app_secret"`
	MchID        string `yaml:"mch_id"`
	MchAPIv3Key  string `yaml:"mch_api_v3_key"`
	MchSerialNo  string `yaml:"mch_serial_no"`
	NotifyURL    string `yaml:"notify_url"`
	MockPay      bool   `yaml:"mock_pay"`
	MockLogin    bool   `yaml:"mock_login"`
}

type LLMConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Mock    bool   `yaml:"mock"`
}

type MiniProgramConfig struct {
	PathPrefix     string `yaml:"path_prefix"`
	LinkBase       string `yaml:"link_base"`
	SessionTTLDays int    `yaml:"session_ttl_days"`
}

type SeedConfig struct {
	AdminUsername string `yaml:"admin_username"`
	AdminPassword string `yaml:"admin_password"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyEnv()
	cfg.setDefaults()
	return &cfg, nil
}

func (c *Config) applyEnv() {
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		c.MySQL.DSN = v
	}
	if v := os.Getenv("SERVER_ADDR"); v != "" {
		c.Server.Addr = v
	}
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("LLM_MOCK"); v != "" {
		c.LLM.Mock = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("WECHAT_MOCK_PAY"); v != "" {
		c.WeChat.MockPay = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("WECHAT_MOCK_LOGIN"); v != "" {
		c.WeChat.MockLogin = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("AES_KEY"); v != "" {
		c.Crypto.AESKey = v
	}
	if v := os.Getenv("JWT_ADMIN_SECRET"); v != "" {
		c.JWT.AdminSecret = v
	}
	if v := os.Getenv("JWT_USER_SECRET"); v != "" {
		c.JWT.UserSecret = v
	}
}

func (c *Config) setDefaults() {
	if c.Server.Addr == "" {
		c.Server.Addr = ":8080"
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.JWT.AdminTTLHours == 0 {
		c.JWT.AdminTTLHours = 24
	}
	if c.JWT.UserTTLHours == 0 {
		c.JWT.UserTTLHours = 168
	}
	if c.MiniProgram.SessionTTLDays == 0 {
		c.MiniProgram.SessionTTLDays = 7
	}
	if c.MiniProgram.PathPrefix == "" {
		c.MiniProgram.PathPrefix = "/pages/discover/index"
	}
	if c.MiniProgram.LinkBase == "" {
		c.MiniProgram.LinkBase = "https://example.com/mp"
	}
	if c.Seed.AdminUsername == "" {
		c.Seed.AdminUsername = "admin"
	}
	if c.Seed.AdminPassword == "" {
		c.Seed.AdminPassword = "admin123"
	}
	if c.LLM.Model == "" {
		c.LLM.Model = "gpt-4o-mini"
	}
}
