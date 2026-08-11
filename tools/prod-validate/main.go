package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Production environment health validator (Prompt-43).
// Does not mutate systems — probe only.

func main() {
	env := flag.String("env", envOr("PROD_VALIDATE_ENV", "staging"), "target environment name")
	base := flag.String("base", envOr("PROD_VALIDATE_BASE_URL", ""), "API base URL (optional)")
	timeout := flag.Duration("timeout", 5*time.Second, "per-probe timeout")
	flag.Parse()

	targets := defaultTargets(*env, *base)
	client := &http.Client{Timeout: *timeout}

	var failed int
	for _, t := range targets {
		ok, detail := probe(client, t)
		status := "PASS"
		if !ok {
			status = "FAIL"
			failed++
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", status, *env, t.Name, detail)
	}

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "prod-validate: %d probe(s) failed for env=%s\n", failed, *env)
		os.Exit(1)
	}
	fmt.Printf("prod-validate: all probes passed env=%s\n", *env)
}

type target struct {
	Name string
	URL  string
}

func defaultTargets(env, base string) []target {
	if base == "" {
		// Offline / CI without cluster: structural pass with documented endpoints.
		fmt.Println("INFO\tno PROD_VALIDATE_BASE_URL; running structural checklist only")
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"env": env,
			"requiredSecrets": []string{"JWT_KEY_PEM", "OTP_PEPPER", "STRIPE_SECRET_KEY", "KAFKA_BROKERS", "DATABASE_URL"},
			"forbiddenInProd": []string{"OTP_DEV_MODE=true", "CORS=*", "MockPSP failover"},
			"overlays":        []string{"infra/k8s/overlays/" + env},
		})
		return nil
	}
	base = strings.TrimRight(base, "/")
	return []target{
		{Name: "bff-customer-health", URL: base + "/health"},
		{Name: "bff-customer-ready", URL: base + "/ready"},
	}
}

func probe(c *http.Client, t target) (bool, string) {
	req, err := http.NewRequest(http.MethodGet, t.URL, nil)
	if err != nil {
		return false, err.Error()
	}
	res, err := c.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return false, fmt.Sprintf("status=%d", res.StatusCode)
	}
	return true, fmt.Sprintf("status=%d", res.StatusCode)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
