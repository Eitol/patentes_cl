package patentes_cl

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	httpClient     *http.Client
	token          string
	clientUser     string
	clientPassword string
	authUser       string
	authPassword   string
	tokenMutex     sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		clientUser:     clientUser,
		clientPassword: clientPassword,
		authUser:       authUser,
		authPassword:   authPassword,
	}
}

func (c *Client) assertAuth() error {
	c.tokenMutex.RLock()
	tokenIsEmpty := c.token == ""
	c.tokenMutex.RUnlock()

	if tokenIsEmpty {
		err := c.refreshToken()
		if err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
	}
	return nil
}

func (c *Client) refreshToken() error {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	if c.token != "" {
		return nil
	}
	reqBodyBytes := fmt.Sprintf("grant_type=password&username=%s&password=%s", c.clientUser, c.clientPassword)
	reqBody := bytes.NewBuffer([]byte(reqBodyBytes))
	req, err := http.NewRequest("POST", tokenURL, reqBody)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.authUser, c.authPassword)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to get token")
	}

	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return err
	}
	c.token = tokenResp.AccessToken
	return nil
}

func (c *Client) GetByRut(rut string) ([]Vehicle, error) {
	err := c.assertAuth()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", vehiclesURL+rut, nil)
	if err != nil {
		return nil, err
	}
	var currentToken string

	c.tokenMutex.RLock()
	currentToken = c.token
	c.tokenMutex.RUnlock()

	req.Header.Add("Authorization", "bearer "+currentToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Token might be expired, try to refresh it
		c.tokenMutex.RLock()
		tokenWasRefreshed := c.token != currentToken
		c.tokenMutex.RUnlock()
		if !tokenWasRefreshed {
			c.tokenMutex.Lock()
			c.token = ""
			c.tokenMutex.Unlock()
			if err := c.refreshToken(); err != nil {
				return nil, err
			}
		}
		return c.GetByRut(rut)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get vehicle information")
	}

	var vehicles []Vehicle
	if err := json.NewDecoder(resp.Body).Decode(&vehicles); err != nil {
		return nil, err
	}

	return vehicles, nil
}
