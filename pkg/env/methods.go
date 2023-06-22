package env

import (
	"os"
)

func GetDefault(key, def string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return def
}

func GetOrNull(key string) *string {
	val, ok := os.LookupEnv(key)
	if ok {
		return &val
	}
	return nil
}
