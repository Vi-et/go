package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

// Config chứa toàn bộ cấu hình ứng dụng
type Config struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`

	DB struct {
		DSN          string `mapstructure:"dsn"`
		MaxOpenConns int    `mapstructure:"max_open_conns"`
		MaxIdleConns int    `mapstructure:"max_idle_conns"`
		MaxIdleTime  string `mapstructure:"max_idle_time"`
	} `mapstructure:"db"`

	Limiter struct {
		RPS     float64 `mapstructure:"rps"`
		Burst   int     `mapstructure:"burst"`
		Enabled bool    `mapstructure:"enabled"`
	} `mapstructure:"limiter"`

	SMTP struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Sender   string `mapstructure:"sender"`
	} `mapstructure:"smtp"`

	CORS struct {
		TrustedOrigins []string `mapstructure:"trusted_origins"`
	} `mapstructure:"cors"`
}

// Load đọc config từ file và environment variables
func Load() (Config, error) {
	v := viper.New()

	// Tên file config (không có extension)
	v.SetConfigName("config")
	// Định dạng file
	v.SetConfigType("yaml")
	// Tìm file ở thư mục gốc project
	v.AddConfigPath(".")
	// Tìm thêm ở thư mục home (tuỳ chọn)
	v.AddConfigPath("$HOME/.greenlight")

	// Cho phép override bằng ENV vars
	// VD: GREENLIGHT_DB_DSN sẽ ghi đè db.dsn
	v.SetEnvPrefix("GREENLIGHT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Đọc file config
	if err := v.ReadInConfig(); err != nil {
		// Nếu không tìm thấy file → vẫn chạy được nhờ ENV vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
		slog.Warn("config.yaml not found, relying on environment variables")
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
