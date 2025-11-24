package elasticsearch

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/opensearch-project/opensearch-go/v2"
)

type Client struct {
	client *opensearch.Client
	index  string
}

func NewClient(url, index, username, password string) (*Client, error) {
	cfg := opensearch.Config{
		Addresses: []string{url},
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create opensearch client: %w", err)
	}

	c := &Client{
		client: client,
		index:  index,
	}

	return c, nil
}

func (c *Client) GetClient() *opensearch.Client {
	return c.client
}
