package internal

// Client is a placeholder for the Keboola SDK client.
type Client struct {
	URL   string
	Token string
}

// NewClient creates a new Keboola client instance.
func NewClient(url, token string) *Client {
	return &Client{
		URL:   url,
		Token: token,
	}
}

// TODO: Add methods for interacting with the Keboola API using your SDK.
