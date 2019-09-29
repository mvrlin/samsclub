package samsclub

import (
	"github.com/mvrlin/samsclub/pkg/config"
)

// SamsClub represents a Checker.
type SamsClub struct {
	Config *config.Config
}

// Run is initializing SamsClub.
func (sc *SamsClub) Run() error {
	return nil
}

// New creates a new instance of SamsClub.
func New() (*SamsClub, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	return &SamsClub{
		Config: cfg,
	}, nil
}
