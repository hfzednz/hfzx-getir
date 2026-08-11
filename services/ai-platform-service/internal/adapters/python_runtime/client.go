package python_runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexora/ai-platform-service/internal/app/ports"
	"github.com/nexora/ai-platform-service/internal/domain"
)

// Client calls the Python FastAPI sidecar; falls back to local heuristic runtime.
type Client struct {
	baseURL  string
	fallback ports.InferenceRuntime
	http     *http.Client
}

func NewClient(baseURL string, fallback ports.InferenceRuntime) *Client {
	return &Client{
		baseURL:  baseURL,
		fallback: fallback,
		http:     &http.Client{Timeout: 2 * time.Second},
	}
}

type predictReq struct {
	ModelKey  string             `json:"modelKey"`
	Version   string             `json:"version"`
	Features  map[string]float64 `json:"features"`
	Inputs    map[string]any     `json:"inputs"`
}

type predictResp struct {
	Predictions map[string]float64 `json:"predictions"`
	Outputs     map[string]any     `json:"outputs"`
}

func (c *Client) Predict(ctx context.Context, model domain.ModelCard, features map[string]float64, inputs map[string]any) (map[string]float64, map[string]any, error) {
	if c.baseURL != "" {
		body, _ := json.Marshal(predictReq{ModelKey: model.Key, Version: model.Version, Features: features, Inputs: inputs})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/predict", bytes.NewReader(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			resp, err := c.http.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					var out predictResp
					if json.NewDecoder(resp.Body).Decode(&out) == nil {
						return out.Predictions, out.Outputs, nil
					}
				}
			}
		}
	}
	if c.fallback != nil {
		return c.fallback.Predict(ctx, model, features, inputs)
	}
	return nil, nil, domain.ErrProviderUnavailable
}
