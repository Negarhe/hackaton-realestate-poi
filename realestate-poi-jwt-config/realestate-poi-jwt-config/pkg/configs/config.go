package configs

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig
	Kenar    KenarConfig
	Session  SessionConfig
	Jwt      JWTConfig
	Server   ServerConfig
	Neshan   NeshanConfig
	Snapp    SnappConfig
	Tapsi    TapsiConfig
}

type NeshanConfig struct {
	NeshanApiKey string `mapstructure:"NeshanApiKey"`
}

type KenarConfig struct {
	AppSlug     string `mapstructure:"KenarAppSlug"`
	ApiKey      string `mapstructure:"KenarApiKey"`
	OauthSecret string `mapstructure:"OauthSecret"`
	BaseURL     string `mapstructure:"BaseUrl"`
}
type ServerConfig struct {
	Port string `mapstructure:"Port"`
}

type DatabaseConfig struct {
	Host                         string
	Port                         int
	Username                     string
	Password                     string
	DBName                       string
	SSLMode                      string
	MaxConns                     int32
	MinConns                     int32
	MaxConnLifetimeJitterMinutes int
	MaxConnLifetimeMinutes       int
	MaxConnIdleTimeMinutes       int
}

type SessionConfig struct {
	AuthKey string `mapstructure:"SessionAuthKey"`
	// EncKey  string `mapstructure:"SessionEncKey"`
}

type JWTConfig struct {
	JwtSecret string `mapstructure:"JwtSecret"`
}

type SnappConfig struct {
	ApiKey        string `mapstructure:"access_token"`
	Clck          string `mapstructure:"clck"`
	Clsk          string `mapstructure:"clsk"`
	GA            string `mapstructure:"ga"`
	GATracking    string `mapstructure:"ga_tracking"`
	YandexDate    string `mapstructure:"ym_d"`
	YandexAd      string `mapstructure:"ym_isad"`
	YandexUID     string `mapstructure:"ym_uid"`
	CookieSession string `mapstructure:"cookie_session"`
}
type TapsiConfig struct {
	Clck         string `mapstructure:"clck"`
	Clsk         string `mapstructure:"clsk"`
	AccessToken  string `mapstructure:"access_token"`
	RefreshToken string `mapstructure:"refresh_token"`
}

func (cfg *KenarConfig) Validate() error {

	if cfg.AppSlug == "" || cfg.OauthSecret == "" {
		return fmt.Errorf("missing required OAuth configurations")
	}
	return nil
}

func LoadConfig() (*Config, error) {
	if os.Getenv("ENV") == "development" {
		err := godotenv.Load("./pkg/configs/.env")
		if err != nil {
			log.Printf("Error loading .env file in LoadConfig" + err.Error())
		}
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./pkg/configs")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// check if config file is not provided
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file")
	}
	for _, key := range viper.AllKeys() {
		// log.Println(key)
		value := viper.GetString(key)
		expanded := os.ExpandEnv(value)
		// log.Println(key)
		// log.Println(value)
		viper.Set(key, expanded)
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	// log.Printf("Config Loaded: %+v", config)
	return &config, nil
}
