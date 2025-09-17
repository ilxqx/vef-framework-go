package config

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
	logPkg "github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/utils"
	"github.com/spf13/viper"
)

var (
	// decodeUsingJsonTagOption configures mapstructure decoder with custom options
	decodeUsingJsonTagOption viper.DecoderConfigOption = func(c *mapstructure.DecoderConfig) {
		c.TagName = "config"
		c.IgnoreUntaggedFields = true
		c.DecodeHook = utils.MapDecoderHook
	}
)

type viperConfig struct {
	v *viper.Viper
}

func (v *viperConfig) Unmarshal(key string, target any) error {
	return v.v.UnmarshalKey(key, target, decodeUsingJsonTagOption)
}

func newConfig() (config.Config, error) {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer(constants.Dot, constants.Underscore)),
		viper.KeyDelimiter(constants.Dot),
		viper.WithLogger(log.NewSLogger("config", 3, logPkg.LevelWarn)),
	)
	v.SetEnvPrefix(constants.EnvKeyPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	v.SetConfigName("application")
	v.SetConfigType("toml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("$VEF_CONFIG_PATH")
	v.AddConfigPath(".")
	v.AddConfigPath("../configs")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return &viperConfig{v: v}, nil
}
