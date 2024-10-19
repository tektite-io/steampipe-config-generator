package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		lvl = "info"
	}

	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.InfoLevel
	}

	log.SetLevel(ll)
	log.SetFormatter(&log.JSONFormatter{})
}
