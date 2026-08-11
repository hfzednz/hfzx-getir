package nexora

// Client is a thin Open Platform / BFF HTTP client stub.
// Domain logic remains in NEXORA services — this SDK only wraps transport.
type Client struct {
	BaseURL string
	APIKey  string
}

func New(baseURL, apiKey string) *Client {
	return &Client{BaseURL: baseURL, APIKey: apiKey}
}
