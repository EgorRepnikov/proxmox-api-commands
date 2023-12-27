package env

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

type Env struct {
	PORT         int
	LOGGER_LEVEL string

	PROXMOX_HOST      string
	PROXMOX_USERNAME  string
	PROXMOX_PASSWORD  string
	REQUEST_TIMEOUT   time.Duration
	AUTHORIZATION_TTL time.Duration
}

func InitEnv() (*Env, error) {
	_ = godotenv.Load()

	env := Env{}
	var err error

	env.PORT, err = getEnvInt("PORT")
	if err != nil {
		return nil, err
	}

	env.LOGGER_LEVEL, err = getEnv("LOGGER_LEVEL")
	if err != nil {
		return nil, err
	}

	env.PROXMOX_HOST, err = getEnv("PROXMOX_HOST")
	if err != nil {
		return nil, err
	}

	env.PROXMOX_USERNAME, err = getEnv("PROXMOX_USERNAME")
	if err != nil {
		return nil, err
	}

	env.PROXMOX_PASSWORD, err = getEnv("PROXMOX_PASSWORD")
	if err != nil {
		return nil, err
	}

	env.REQUEST_TIMEOUT, err = getEnvDuration("REQUEST_TIMEOUT")
	if err != nil {
		return nil, err
	}

	env.AUTHORIZATION_TTL, err = getEnvDuration("AUTHORIZATION_TTL")
	if err != nil {
		return nil, err
	}

	return &env, nil
}

func getEnv(path string) (string, error) {
	env := os.Getenv(path)
	if env != "" {
		return env, nil
	} else {
		return "", errors.Errorf("%s env variable is not defined", path)
	}
}

func getEnvInt(path string) (int, error) {
	env, err := getEnv(path)
	if err != nil {
		return 0, err
	}
	intEnv, err := strconv.Atoi(env)
	if err != nil {
		return 0, errors.Errorf("%s env variable is not int convertible", path)
	}
	return intEnv, nil
}

func getEnvIntBool(path string) (bool, error) {
	env, err := getEnvInt(path)
	if err != nil {
		return false, err
	}
	boolEnv := env != 0
	return boolEnv, nil
}

func getEnvDuration(path string) (time.Duration, error) {
	env, err := getEnv(path)
	if err != nil {
		return 0, err
	}
	durationEnv, err := time.ParseDuration(env)
	if err != nil {
		return 0, err
	}
	return durationEnv, nil
}

func getEnvStringArray(path string) ([]string, error) {
	env, err := getEnv(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(env, ","), nil
}

func getEnvFloat(path string) (float64, error) {
	env, err := getEnv(path)
	if err != nil {
		return 0, err
	}
	intEnv, err := strconv.ParseFloat(env, 64)
	if err != nil {
		return 0, errors.Errorf("%s env variable is not int convertible", path)
	}
	return intEnv, nil
}
