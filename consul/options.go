package consul

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

// Option defines a function type for configuring the Consul client.
type Option func(*api.Config)

// WithURI sets the Consul agent URI.
// URI format: [scheme://][user:password@]host[:port][/path][?query]
// URI examples:
//   - consul://localhost:8500
//   - consul+http://localhost:8500?dc=dc1&token=token&wait=10s
func WithURI(uri string) Option {
	return func(conf *api.Config) {
		urlObj, err := url.Parse(uri)
		if err != nil {
			panic(err)
		}
		schema := urlObj.Scheme
		if strings.HasPrefix(schema, "consul+") {
			schema = strings.TrimPrefix(schema, "consul+")
		} else if schema == "consul" {
			schema = "http"
		} else if schema == "consuls" {
			schema = "https"
		}

		if urlObj.User != nil {
			pass, _ := urlObj.User.Password()
			conf.HttpAuth = &api.HttpBasicAuth{
				Username: urlObj.User.Username(),
				Password: pass,
			}
		}

		query := urlObj.Query()
		conf.Address = urlObj.Host
		conf.Scheme = schema
		conf.PathPrefix = urlObj.Path
		_ = setIfNoEmpty(&conf.WaitTime, query.Get("wait"), time.ParseDuration)
		setIfNoEmptyStr(&conf.Datacenter, query.Get("dc"))
		setIfNoEmptyStr(&conf.Token, query.Get("token"))
		setIfNoEmptyStr(&conf.TokenFile, query.Get("token_file"))
		setIfNoEmptyStr(&conf.Partition, query.Get("partition"))
	}
}

// WithAddress sets the Consul agent address.
func WithAddress(addr string) Option {
	return func(conf *api.Config) {
		conf.Address = addr
	}
}

// WithDatacenter sets the Consul datacenter.
func WithDatacenter(dc string) Option {
	return func(conf *api.Config) {
		conf.Datacenter = dc
	}
}

// WithToken sets the Consul token.
func WithToken(token string) Option {
	return func(conf *api.Config) {
		conf.Token = token
	}
}

// WithTokenFile sets the Consul token file.
func WithHttpClient(client *http.Client) Option {
	return func(conf *api.Config) {
		conf.HttpClient = client
	}
}

// WithTokenFile sets the Consul token file.
func WithWaitTime(waitTime time.Duration) Option {
	return func(conf *api.Config) {
		conf.WaitTime = waitTime
	}
}

func setIfNoEmpty[T any](dur *T, value string, cast func(string) (T, error)) error {
	if value != "" {
		if v, err := cast(value); err == nil {
			*dur = v
		} else {
			return err
		}
	}
	return nil
}

func setIfNoEmptyStr(dur *string, value string) {
	if value != "" {
		*dur = value
	}
}
