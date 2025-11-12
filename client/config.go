package client

import (
	"net/url"
	"time"
)

const (
	defaultTimeout = time.Second * 15
)

type FileServiceConfig struct {
	Addr    string
	Timeout time.Duration
}

func (config *FileServiceConfig) validate() error {

	_, err := url.Parse(config.Addr)
	if err != nil {
		return ErrInvalidAddr
	}

	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}

	return nil
}
