Dynflare
========

Detects your external IP address and updates a DNS record in Cloudflare.

## Install

From source: `go get -u github.com/trapped/dynflare`

Docker image: planned

## Setup

Both environment variables and CLI flags are supported.

- `CLOUDFLARE_EMAIL` or `-cloudflare-email`: Cloudflare email used for API authentication
- `CLOUDFLARE_API_KEY` or `-cloudflare-api-key`: Cloudflare API key
- `CLOUDFLARE_ZONE` or `-cloudflare-zone`: your second-level domain name (like `example.com`)
- `CLOUDFLARE_RECORD` or `-cloudflare-record`: Cloudflare DNS record type and name separated by space (like `A dynamic.example.com`)
- `-print-records`: pretty-print all Cloudflare DNS records for the specified zone in a tab-separated table
- `-dry`: fake the actual DNS update

After setting everything up, you should be able to just run `./dynflare`; depending on your use-case, you might want to set up a cronjob or a Kubernetes CronJob.

## Supported IP detection mechanisms

- [x] [IPify](https://www.ipify.org)
- [x] [MyIP](https://www.myip.com)
- [ ] More public services (planned)
- [ ] Self-connect using [ngrok](https://ngrok.com) (planned)
- [x] Fetch from multiple sources and pick most seen

## Contributing

### Adding a new IP provider

IP providers must implement the `Provider` interface:

```go
type Provider interface {
	Name() string
	Fetch() (string, error)
}
```

Anyone is free to submit a pull request implementing a new IP provider.
