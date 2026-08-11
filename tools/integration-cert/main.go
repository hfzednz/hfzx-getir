package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	root := findRoot()
	regPath := filepath.Join(root, "docs", "launch", "service-registry.yaml")
	b, err := os.ReadFile(regPath)
	if err != nil {
		fail("registry: %v", err)
	}
	names := parseServiceNames(string(b))
	if len(names) == 0 {
		fail("no services in registry")
	}
	missing := []string{}
	for _, n := range names {
		mod := filepath.Join(root, "services", n, "go.mod")
		if _, err := os.Stat(mod); err != nil {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		fail("missing go.mod for: %s", strings.Join(missing, ", "))
	}
	fmt.Printf("CERT registry_ok services=%d\n", len(names))

	// Critical path unit tests
	critical := []string{
		"identity-service", "checkout-service", "payment-service", "order-service",
		"inventory-service", "bff-customer", "realtime-gateway", "security-service",
		"platform-ops-service", "location-service", "quality-service",
	}
	for _, svc := range critical {
		dir := filepath.Join(root, "services", svc)
		cmd := exec.Command("go", "test", "./...")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", out)
			fail("test failed: %s", svc)
		}
		fmt.Printf("CERT test_ok %s\n", svc)
	}
	fmt.Println("CERT RESULT=PASS launch_wave_b")
}

func parseServiceNames(yaml string) []string {
	out := []string{}
	lines := strings.Split(yaml, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- name:") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "- name:"))
			out = append(out, name)
		}
	}
	return out
}

func findRoot() string {
	wd, _ := os.Getwd()
	for d := wd; d != filepath.Dir(d); d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "docs", "launch", "service-registry.yaml")); err == nil {
			return d
		}
	}
	return wd
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "CERT FAIL "+format+"\n", args...)
	os.Exit(1)
}
