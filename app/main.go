package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	flags "github.com/jessevdk/go-flags"
	"github.com/verdel/squid-auth-healthcheck/app/checker"
	"github.com/verdel/squid-auth-healthcheck/app/slice"
)

const (
	version = "0.0.1"
)

var opts struct {
	URL               []string `short:"u" long:"url" description:"url to check for availability (required)" required:"true"`
	AuthType          []string `long:"auth-type" description:"type of used proxy authentication mechanism. [ntlm, kerberos] (required)" required:"true"`
	ProxyAddr         string   `long:"proxy-addr" description:"proxy server address (required)" required:"true"`
	ProxyPort         int      `long:"proxy-port" description:"proxy server port (default: 3128)" default:"3128"`
	ProxyUsername     string   `long:"proxy-username" description:"proxy user login (required)" required:"true"`
	ProxyPassword     string   `long:"proxy-password" description:"proxy user password (required)" required:"true"`
	ConnectionTimeout int      `long:"timeout" description:"healthcheck connection timeout in seconds (default: 2)" default:"2"`
	StrictURL         bool     `long:"strict-url" description:"the check returns a positive result only if all URLs are available"`
	StrictAuth        bool     `long:"strict-auth" description:"the check returns a positive result only if url are available with all auth method"`
	Verbose           bool     `short:"v" long:"verbose" description:"output verbose healthcheck information"`
}
var allowAuthType = [2]string{"ntlm", "kerberos"}

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

	if len(opts.AuthType) > len(allowAuthType) {
		fmt.Println("Too many authentication type")
		os.Exit(1)
	}

	for _, item := range opts.AuthType {
		if !slice.StringInSlice(item, opts.AuthType) {
			fmt.Printf("Authentication type %s is not allowed", item)
			os.Exit(1)
		}
	}

	var wg sync.WaitGroup

	ch := make(chan checker.HealthResponse, len(opts.AuthType)*len(opts.URL))
	var ntlm checker.Interface = checker.NewAuthNTLM(opts.ProxyAddr, opts.ProxyPort, opts.ProxyUsername, opts.ProxyPassword, opts.ConnectionTimeout)
	kerberos, err := checker.NewAuthKerberos(opts.ProxyAddr, opts.ProxyPort, opts.ProxyUsername, opts.ProxyPassword, opts.ConnectionTimeout)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	wg.Add(len(opts.AuthType))

	if slice.StringInSlice("ntlm", opts.AuthType) {
		go ntlm.Check(opts.URL, ch, &wg)
	}
	if slice.StringInSlice("kerberos", opts.AuthType) {
		go kerberos.Check(opts.URL, ch, &wg)
	}

	wg.Wait()
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

	if opts.StrictAuth {
		if len(okAuthResult) < len(opts.AuthType) {
			exitErr(opts.Verbose)
		}
	}

	if opts.StrictURL {
		if len(okURLResult) < len(opts.URL) {
			exitErr(opts.Verbose)
		}
	} else {
		if len(okURLResult) == 0 {
			exitErr(opts.Verbose)
		}
	}
	exitOK(opts.Verbose)
}
