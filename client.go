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

const (
	tokenURL    = "https://pagosrvm.srcei.cl/PortalRvm/oauth/token"
	vehiclesURL = "https://pagosrvm.srcei.cl/PortalRvm/api/lista/ppu/"

	clientUser     = "PORT_RVM_2021" // Asumiendo que estas son las credenciales (clientId:clientSecret)
	clientPassword = "E5ED00A617CADBBCE1C11EF3689FCFDD2CF599E1CA9B71D0F8E2CDD592F593DD"
	authUser       = "4F7C1B936BD1F698138445C05591B52A2B64DB367AF66F30D270E25F62ADA00F"
	authPassword   = "D358F19A5FF25CDC0CF7C3FD5DDD2434A211BB2F27BD15DA0146E730B6134D30"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type Vehicle struct {
	Ppu             string `json:"ppu"`
	Marca           string `json:"marca"`
	Modelo          string `json:"modelo"`
	Tipo            string `json:"tipo"`
	AFabricacion    string `json:"aFabricacion"`
	NroMotor        string `json:"nroMotor"`
	NroChasis       string `json:"nroChasis"`
	NroSerie        string `json:"nroSerie"`
	NroVin          string `json:"nroVin"`
	CodigoColorBase string `json:"codigoColorBase"`
	DescColorBase   string `json:"descColorBase"`
	RestoColor      string `json:"restoColor"`
	Calidad         string `json:"calidad"`
	DvPpu           string `json:"dvPpu"`
	TipoPropietario string `json:"tipoPropietario"`
}

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
