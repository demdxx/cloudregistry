package zookeeper

import (
	"net/url"
	"strings"
	"time"

	"github.com/demdxx/xtypes"
)

// Option is a function that configures the ZooKeeper connection.
type Option func(*zkConfig)

// WithHosts sets the ZooKeeper hosts.
func WithHosts(hosts []string) Option {
	return func(conf *zkConfig) {
		conf.hosts = hosts
	}
}

// WithSessionTimeout sets the session timeout.
func WithSessionTimeout(timeout time.Duration) Option {
	return func(conf *zkConfig) {
		conf.sessionTimeout = timeout
	}
}

// WithBasePath sets the base path for all operations.
func WithBasePath(basePath string) Option {
	return func(conf *zkConfig) {
		conf.basePath = basePath
	}
}

// WithURI prepare configuration from ZooKeeper URI.
// The URI should be in the format: zookeeper://host1:port1,host2:port2/myapp?timeout=5s
func WithURI(uri string) Option {
	return func(conf *zkConfig) {
		urlObj, err := url.Parse(uri)
		if err != nil {
			panic("invalid ZooKeeper URI: " + err.Error())
		}

		if urlObj.Scheme != "" && urlObj.Scheme != "zookeeper" && urlObj.Scheme != "zk" {
			panic("invalid ZooKeeper URI scheme: " + urlObj.Scheme)
		}

		if urlObj.Host != "" {
			conf.hosts = xtypes.Slice[string](strings.Split(urlObj.Host, ",")).
				Filter(func(val string) bool { return val != "" })
		}

		query := urlObj.Query()
		_ = setVal(&conf.sessionTimeout, query.Get("timeout"), func(s string) (time.Duration, error) {
			if s == "" {
				return defaultSessionTimeout, nil
			}
			tm, err := time.ParseDuration(s)
			if err != nil || tm <= 0 {
				// If parsing fails, use default value
				return defaultSessionTimeout, nil
			}
			return tm, nil
		})
		_ = setVal(&conf.basePath, urlObj.Path, func(s string) (string, error) {
			if s == "" || s == `/` {
				return defaultBasePath, nil
			}
			return `/` + strings.Trim(s, `/`), nil
		})
	}
}

func setVal[T any](dur *T, value string, cast func(string) (T, error)) error {
	if v, err := cast(value); err == nil {
		*dur = v
	} else {
		return err
	}
	return nil
}
