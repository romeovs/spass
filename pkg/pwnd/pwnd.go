package pwnd

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

func (c *Client) Check(password string) (bool, error) {
	client := &http.Client{}

	h := sha1.New()
	io.WriteString(h, password)

	hash := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	prefix := hash[:5]

	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("hibp-api-key", c.apiKey)
	req.Header.Add("Agent", "spass 1.0")
	req.Header.Add("Add-Padding", "true")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to fetch range from havibeenpwnd.com")
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("response %d from haveibeenpwned.com", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body from havibeenpwnd.com")
	}

	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Split(line, ":")
		if parts[1] == "0" {
			continue
		}

		sum := prefix + parts[0]
		if sum != hash {
			continue
		}

		return true, nil
	}

	return false, nil
}
