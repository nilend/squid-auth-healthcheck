package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	flags "github.com/jessevdk/go-flags"
	"github.com/nilend/squid-auth-healthcheck/app/checker"
	"github.com/nilend/squid-auth-healthcheck/app/slice"
)

const (
	version = "0.0.6"
)

var opts struct {
	URL               []string `short:"u" long:"url" description:"url to check for availability (required)" required:"true"`
	ProxyAddr         string   `long:"proxy-addr" description:"proxy server address (required)" required:"true"`
	ProxyPort         int      `long:"proxy-port" description:"proxy server port (default: 3128)" default:"3128"`
	HostAddr          string   `long:"host-addr" description:"host listen address (If healthz needed))"`
	HostPort          int      `long:"host-port" description:"host listen port (default: 3127)" default:"3127"`
	ConnectionTimeout int      `long:"timeout" description:"healthcheck connection timeout in seconds (default: 2)" default:"2"`
	StrictURL         bool     `long:"strict-url" description:"the check returns a positive result only if all URLs are available"`
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

	if opts.HostAddr != "" {
		http.HandleFunc("/healthz", handler)
		hostUrl := fmt.Sprintf("%s:%d", opts.HostAddr, opts.HostPort)
		fmt.Printf("Start to listening on http://%s:\n", hostUrl)
		log.Fatal(http.ListenAndServe(hostUrl, nil))
	} else if do() == 0 {
		exitOK(opts.Verbose)
	} else {
		exitErr(opts.Verbose)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if do() == 0 {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(400)
	}
}

func do() int {

	var wg sync.WaitGroup

	ch := make(chan checker.HealthResponse, len(opts.URL))
	wg.Add(1)
	var no checker.Interface = checker.NewAuthNo(opts.ProxyAddr, opts.ProxyPort, opts.ConnectionTimeout)
	go no.Check(opts.URL, ch, &wg)

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

	if opts.StrictURL {
		if len(okURLResult) < len(opts.URL) {
			return 1
		}
	} else {
		if len(okURLResult) == 0 {
			return 1
		}
	}

	return 0
}
