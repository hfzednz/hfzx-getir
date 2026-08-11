package main

import (
	"fmt"
	"os"
)

// CLI stub: documents Final Genesis certification flow.
func main() {
	fmt.Println("genesis-cert tool")
	fmt.Println("1) POST /v1/autonomy/bootstrap")
	fmt.Println("2) GET  /v1/autonomy/gates")
	fmt.Println("3) POST /v1/autonomy/genesis")
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Println("Set HTTP_ADDR=http://localhost:8125 and X-Tenant-Id in CI.")
	}
}
