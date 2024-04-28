package config

import "time"

type Config struct {
	DryRun               bool
	PolicyResyncInterval time.Duration
}
