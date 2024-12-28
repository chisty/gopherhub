package env

import (
	"os"
	"strconv"
	"time"
)

func GetString(key, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return val
}

func GetInt(key string, defaultValue int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}

	return valAsInt
}

func GetDuration(key string, defaultValue time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	valAsDuration, err := time.ParseDuration(val)
	if err != nil {
		return defaultValue
	}

	return valAsDuration
}
