package config

type Config interface {
	Unmarshal(key string, target any) error
}
