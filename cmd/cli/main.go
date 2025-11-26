package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gomonov/otus-go-project/internal/cli"
)

func main() {
	var baseURL string
	flag.StringVar(&baseURL, "url", "http://localhost:8080", "Base URL of the anti-brute force service")
	flag.Parse()

	if err := cli.Execute(baseURL, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
