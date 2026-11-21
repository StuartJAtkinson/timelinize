package main

import (
	"context"
	"os"

	"github.com/StuartJAtkinson/timelinize/mcp"
)

func main() {
	if err := mcp.Serve(context.Background(), os.Stdin, os.Stdout, os.Getenv("TLZ_URL")); err != nil {
		os.Exit(1)
	}
}
