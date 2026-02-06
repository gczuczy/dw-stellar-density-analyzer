package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	DB DBConfig `koanf:"db"`
	HTTP HTTPConfig `koanf:"http"`
	OAuth2 OAuth2Config `koanf:"oauth2"`
}

type DBConfig struct {
	Host string `koanf:"host"`
	Port *uint16 `koanf:"port"`
	Database string `koanf:"dbname"`
	User string `koanf:"user"`
	Password string `koanf:"password"`
	MaxConns int32 `koanf:"maxconns"`
	MinConns int32 `koanf:"minconns"`
}

type HTTPConfig struct {
	Port uint16 `koanf:"port"`
}

type OAuth2Config struct {
	ClientID string `koanf:"clientid"`
	ClientSecret string `koanf:"clientsecret"`
	Issuer string `koanf:"issuer"`
}


func ParseConfig(k *koanf.Koanf) (*Config, error) {
	var (
		err error
	)
	cfgfile := k.String(`config`)

	if err = k.Load(file.Provider(cfgfile), yaml.Parser()); err != nil {
		return nil, err
	}

	cfg := Config{
		DB: DBConfig{
			MaxConns: 8,
			MinConns: 1,
		},
		HTTP: HTTPConfig{
			Port: 80,
		},
	}
	if err = k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
