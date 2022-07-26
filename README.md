# squid-auth-healthcheck

This application checks the availability of the URL using a proxy server .
The application returns 0 or 1, depending on the input conditions and the availability of the URL.
If you use the `--verbose` flag, it returns json with the availability status and also the full response time for each URL

## Building

The service based on go-curl and required libcurl development files. Use your OS package manager to install libcurl-devel or something like this.

## Running

### Example

```
squid-auth-healthcheck --proxy-addr 127.0.0.1 --url https://google.com --strict-url --verbose --host-addr 127.0.0.1

```

## Parameters

```
Application Options:
  -u, --url=            url to check for availability (required)
      --proxy-addr=     proxy server address (required)
      --proxy-port=     proxy server port (default: 3128) (default: 3128)
      --host-addr=      host listen address (If healthz needed))
      --host-port=      host listen port (default: 3127)
      --timeout=        healthcheck connection timeout in seconds (default: 2) (default: 2)
      --strict-url      the check returns a positive result only if all URLs are available
  -v, --verbose         output verbose healthcheck information
```