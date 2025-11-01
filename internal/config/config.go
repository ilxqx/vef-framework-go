package config

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	ilog "github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mapx"
)

var (
	// DecodeUsingJsonTagOption configures mapstructure decoder with custom options.
	decodeUsingJsonTagOption viper.DecoderConfigOption = func(c *mapstructure.DecoderConfig) {
		c.TagName = "config"
		c.IgnoreUntaggedFields = true
		c.DecodeHook = mapx.DecoderHook
	}
	configName = "application"
	configDir  = "configs"
)

type ViperConfig struct {
	v *viper.Viper
}

func (v *ViperConfig) Unmarshal(key string, target any) error {
	return v.v.UnmarshalKey(key, target, decodeUsingJsonTagOption)
}

func newConfig() (config.Config, error) {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer(constants.Dot, constants.Underscore)),
		viper.KeyDelimiter(constants.Dot),
		viper.WithLogger(ilog.NewSLogger("config", 3, log.LevelWarn)),
	)
	v.SetEnvPrefix(constants.EnvKeyPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	v.SetConfigName(configName)
	v.SetConfigType("toml")
	v.AddConfigPath("./" + configDir)
	v.AddConfigPath(constants.Dollar + constants.EnvConfigPath)
	v.AddConfigPath(".")
	v.AddConfigPath("../" + configDir)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return &ViperConfig{v: v}, nil
}
