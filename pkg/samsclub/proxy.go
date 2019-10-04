package samsclub

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// handleProxy is appending good proxy to array.
func (sc *SamsClub) handleProxy(address string, protocol string) {
	defer wg.Done()

	proxyURL, err := url.Parse(fmt.Sprintf("%s://%s", protocol, address))
	if err != nil {
		return
	}

	c := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := c.Get("https://www.samsclub.com/tealeaf/tealeafTarget.jsp")
	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		mu.Lock()
		sc.ProxyList = append(sc.ProxyList, proxyURL)
		mu.Unlock()
	}
}

// GetProxies is fetching proxy list & store it in memory.
func (sc *SamsClub) GetProxies() error {
	const MaximumRetries = 3

	for i := 0; i < MaximumRetries; i++ {
		protocol := sc.Config.Proxy.Protocol
		url := sc.Config.Proxy.URL

		if protocol == "" || url == "" {
			return errors.New("Proxy is not configurated")
		}

		log.Println("Downloading proxy list..")

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		log.Println("Ping proxies..")

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			wg.Add(1)
			go sc.handleProxy(scanner.Text(), protocol)
		}

		wg.Wait()
		total := len(sc.ProxyList)

		if total > 0 {
			break
		}
	}

	return nil
}
