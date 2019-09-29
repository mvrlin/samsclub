package samsclub

import (
	"net/url"
	"sync"

	"github.com/mvrlin/samsclub/pkg/config"
)

var mu sync.Mutex
var wg sync.WaitGroup

// SamsClub represents a Checker.
type SamsClub struct {
	Config    *config.Config
	ProxyList []*url.URL
}

// Run is initializing SamsClub.
func (sc *SamsClub) Run() error {
	var err error

	if err = sc.GetProxies(); err != nil {
		return err
	}

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
