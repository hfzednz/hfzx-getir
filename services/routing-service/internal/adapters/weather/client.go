package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nexora/routing-service/internal/app/ports"
)

// Client fetches weather and maps conditions to an ETA multiplier.
// Empty baseURL → always 1.0 (safe default). HTTP failures also degrade to 1.0.
type Client struct {
	baseURL string
	apiKey  string
	log     *slog.Logger
	http    *http.Client
}

// NewClient returns a WeatherClient. Configure via WEATHER_URL / OPENWEATHER_URL.
func NewClient(baseURL, apiKey string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	c := &Client{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(apiKey),
		log:     log,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
	if baseURL != "" {
		log.Info("weather.http", "url", baseURL, "apiKeyConfigured", apiKey != "")
	} else {
		log.Info("weather.noop", "note", "WEATHER_URL/OPENWEATHER_URL empty; factor=1.0")
	}
	return c
}

// Factor returns a weather multiplier for ETA (≥1 in adverse conditions).
func (c *Client) Factor(ctx context.Context, req ports.WeatherFactorRequest) (float64, error) {
	if c == nil || c.baseURL == "" {
		return 1.0, nil
	}
	factor, err := c.fetchFactor(ctx, req.Lat, req.Lon)
	if err != nil {
		c.log.Warn("weather.factor.fallback", "err", err, "lat", req.Lat, "lon", req.Lon)
		return 1.0, nil
	}
	if factor <= 0 {
		factor = 1.0
	}
	c.log.Debug("weather.factor", "lat", req.Lat, "lon", req.Lon, "factor", factor)
	return factor, nil
}

func (c *Client) fetchFactor(ctx context.Context, lat, lon float64) (float64, error) {
	u, err := c.buildURL(lat, lon)
	if err != nil {
		return 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	res, err := c.http.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return 0, err
	}
	if res.StatusCode >= 300 {
		return 0, fmt.Errorf("weather status %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	return mapWeatherPayload(raw)
}

func (c *Client) buildURL(lat, lon float64) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	// API root → OpenWeather current-weather path; otherwise keep configured path.
	if u.Path == "" || u.Path == "/" {
		u.Path = "/data/2.5/weather"
	}
	q := u.Query()
	q.Set("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', 6, 64))
	if c.apiKey != "" && q.Get("appid") == "" {
		q.Set("appid", c.apiKey)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type weatherPayload struct {
	Factor  float64 `json:"factor"`
	Weather []struct {
		Main string `json:"main"`
		ID   int    `json:"id"`
	} `json:"weather"`
}

func mapWeatherPayload(raw []byte) (float64, error) {
	var p weatherPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return 0, err
	}
	if p.Factor > 0 {
		return p.Factor, nil
	}
	if len(p.Weather) == 0 {
		return 1.0, nil
	}
	return conditionFactor(p.Weather[0].Main), nil
}

func conditionFactor(main string) float64 {
	switch strings.ToLower(strings.TrimSpace(main)) {
	case "clear", "clouds":
		return 1.0
	case "mist", "fog", "haze", "smoke", "dust", "sand", "ash", "squall":
		return 1.1
	case "drizzle":
		return 1.15
	case "rain":
		return 1.25
	case "snow":
		return 1.4
	case "thunderstorm":
		return 1.45
	case "tornado":
		return 1.6
	default:
		return 1.0
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

var _ ports.WeatherClient = (*Client)(nil)
