package teamweek

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.2"
	userAgent      = "go-teamweek/" + libraryVersion
	defaultBaseURL = "https://teamweek.com/api/v4/"
)

type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
}

func (c *Client) get(url string, dest interface{}) error {
	return c.request(url, http.MethodGet, dest, nil)
}

func (c *Client) post(url string, dest, body interface{}) error {
	return c.request(url, http.MethodPost, dest, body)
}

func (c *Client) put(url string, dest, body interface{}) error {
	return c.request(url, http.MethodPut, dest, body)
}

func (c *Client) remove(url string) error {
	return c.request(url, http.MethodDelete, nil, nil)
}

func (c *Client) request(urlStr, method string, dest, body interface{}) error {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	u := c.BaseURL.ResolveReference(rel)

	var req *http.Request
	if body != nil {
		var buf bytes.Buffer // request are quite small anyway...
		if err := json.NewEncoder(&buf).Encode(&body); err != nil {
			return err
		}
		req, err = http.NewRequest(method, u.String(), &buf)
	} else {
		req, err = http.NewRequest(method, u.String(), nil)
	}

	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := handleResponseStatuses(resp); err != nil {
		return err
	}

	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	return nil
}

func handleResponseStatuses(resp *http.Response) error {
	switch resp.StatusCode {

	case http.StatusUnauthorized: // 401
		return errors.New("Invalid authorization, check credentials and/or reauthenticate")

	case http.StatusForbidden: // 403
		return errors.New("The ressource your are trying to access is beyond the scope of your current user")

	case http.StatusNotFound: // 404
		return errors.New("The requested ressource could not be found.")

	case http.StatusInternalServerError: // 500
		return errors.New("Teamweek API experienced an internal error. Please try again later.")

	default:
		if resp.StatusCode >= 400 {
			return fmt.Errorf("Teamweek API returned an unexpected status code: %d", resp.StatusCode)
		}
	}
	return nil
}

// NewClient returns a new Teamweek API client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	client := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	return client
}
