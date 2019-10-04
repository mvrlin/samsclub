package samsclub

import (
	"errors"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/mvrlin/samsclub/pkg/chromium"
	"github.com/mvrlin/samsclub/pkg/config"
)

var mu sync.Mutex
var wg sync.WaitGroup

// SamsClub represents a Checker.
type SamsClub struct {
	Accounts      chan *Account
	AccountsQueue chan *Account
	Config        *config.Config
	ExecPath      string
	ProxyList     []*url.URL
}

// GetExecPath is getting the executive path of chromium.
func (sc *SamsClub) GetExecPath() error {
	execPath, err := chromium.ExecPath()
	if err != nil {
		return err
	}

	sc.ExecPath = execPath
	return nil
}

// Run is initializing SamsClub.
func (sc *SamsClub) Run() error {
	if len(os.Args) < 2 {
		return errors.New("No accounts file is provided")
	}

	var err error

	if err = sc.GetExecPath(); err != nil {
		return err
	}

	go func() {
		for {
			if err = sc.GetProxies(); err != nil {
				log.Fatalln(err)
				break
			}

			time.Sleep(30 * time.Second)
		}
	}()

	if err = sc.GetAccounts(); err != nil {
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
		Accounts:      make(chan *Account),
		AccountsQueue: make(chan *Account),
		Config:        cfg,
	}, nil
}
