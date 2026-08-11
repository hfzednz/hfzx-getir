package main

import (
	"fmt"
	"os"
)

// CLI stub: documents hyperscale certification flow (HTTP client can be added in CI).
func main() {
	fmt.Println("hyperscale-cert tool")
	fmt.Println("1) POST /v1/hyperscale/bootstrap")
	fmt.Println("2) GET  /v1/hyperscale/gates")
	fmt.Println("3) POST /v1/hyperscale/certificates")
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Println("Set HTTP_ADDR=http://localhost:8124 and X-Tenant-Id in CI.")
	}
}
