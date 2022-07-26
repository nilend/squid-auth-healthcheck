package main

import (
	"encoding/json"
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/nilend/squid-auth-healthcheck/app/checker"
	"github.com/nilend/squid-auth-healthcheck/app/slice"
)

const (
	version = "0.0.4"
)

var opts struct {
	URL               []string `short:"u" long:"url" description:"url to check for availability (required)" required:"true"`
	ProxyAddr         string   `long:"proxy-addr" description:"proxy server address (required)" required:"true"`
	ProxyPort         int      `long:"proxy-port" description:"proxy server port (default: 3128)" default:"3128"`
	ConnectionTimeout int      `long:"timeout" description:"healthcheck connection timeout in seconds (default: 2)" default:"2"`
	Verbose           bool     `short:"v" long:"verbose" description:"output verbose healthcheck information"`
}

func exitOK(verbose bool) {
	if !verbose {
		fmt.Println(1)
	}
	os.Exit(0)
}

func exitErr(verbose bool) {
	if !verbose {
		fmt.Println(0)
	}
	os.Exit(1)
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = fmt.Sprintf("\n\nVersion: %s", version)

	if len(os.Args) == 1 {
		parser.WriteHelp(os.Stderr)
		os.Exit(0)
	}
	_, err := parser.Parse()

	if err != nil {
		os.Exit(1)
	}

	ch := make(chan checker.HealthResponse, len(opts.URL))
	var no checker.Interface = checker.NewAuthNo(opts.ProxyAddr, opts.ProxyPort, opts.ConnectionTimeout)
	go no.Check(opts.URL, ch)

	close(ch)

	var result []checker.HealthResponse

	var okURLResult []string
	var okAuthResult []string

	for response := range ch {
		if response.Status == 1 {
			okURLResult = slice.AppendIfMissing(okURLResult, response.URL)
			okAuthResult = slice.AppendIfMissing(okAuthResult, response.AuthType)
		}
		result = append(result, response)
	}

	if opts.Verbose {
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(result)
	}

	exitOK(opts.Verbose)
}
