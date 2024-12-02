package etcd

import (
	"crypto/tls"
	"net/url"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// Option is a configuration option for the etcd client.
type Option func(conf *clientv3.Config)

// WithURI sets the connection from a URI, like: http://user:password@localhost:2379,localhost:2380,localhost:2381?timeout=5s&keepalive=5s&keepalivetimeout=5s&maxmsgsize=5&rejectoldcluster=true
func WithURI(uri string) Option {
	return func(conf *clientv3.Config) {
		urlObj, err := url.Parse(uri)
		if err != nil {
			panic(err)
		}

		schema := urlObj.Scheme
		switch schema {
		case "etcds":
			schema = "https"
		case "etcd":
			schema = "http"
		}

		if urlObj.User != nil {
			conf.Username = urlObj.User.Username()
			conf.Password, _ = urlObj.User.Password()
		}

		if urlObj.Host != "" {
			conf.Endpoints = xtypes.SliceApply(strings.Split(urlObj.Host, ","),
				func(s string) string { return schema + "://" + s })
		}

		query := urlObj.Query()
		_ = setIfNoEmpty(&conf.DialTimeout, query.Get("timeout"), time.ParseDuration)
		_ = setIfNoEmpty(&conf.DialKeepAliveTime, query.Get("keepalive"), time.ParseDuration)
		_ = setIfNoEmpty(&conf.DialKeepAliveTimeout, query.Get("keepalivetimeout"), time.ParseDuration)
		_ = setIfNoEmptyAny(&conf.MaxCallSendMsgSize, query.Get("maxmsgsize"), gocast.TryNumber[int])
		_ = setIfNoEmptyAny(&conf.MaxCallRecvMsgSize, query.Get("maxmsgsize"), gocast.TryNumber[int])
		_ = setIfNoEmpty(&conf.RejectOldCluster, query.Get("rejectoldcluster"), func(s string) (bool, error) {
			return gocast.Bool(s), nil
		})
	}
}

// WithEndpoints sets the endpoints.
func WithEndpoints(endpoints []string) Option {
	return func(conf *clientv3.Config) {
		conf.Endpoints = endpoints
	}
}

// WithDialTimeout sets the dial timeout.
func WithDialTimeout(timeout time.Duration) Option {
	return func(conf *clientv3.Config) {
		conf.DialTimeout = timeout
	}
}

// WithDialKeepAliveTime sets the dial keep alive time.
func WithDialKeepAliveTime(timeout time.Duration) Option {
	return func(conf *clientv3.Config) {
		conf.DialKeepAliveTime = timeout
	}
}

// WithDialKeepAliveTimeout sets the dial keep alive timeout.
func WithDialKeepAliveTimeout(timeout time.Duration) Option {
	return func(conf *clientv3.Config) {
		conf.DialKeepAliveTimeout = timeout
	}
}

// WithMaxCallSendMsgSize sets the max call send message size.
func WithMaxCallSendMsgSize(size int) Option {
	return func(conf *clientv3.Config) {
		conf.MaxCallSendMsgSize = size
	}
}

// WithMaxCallRecvMsgSize sets the max call receive message size.
func WithMaxCallRecvMsgSize(size int) Option {
	return func(conf *clientv3.Config) {
		conf.MaxCallRecvMsgSize = size
	}
}

// WithTLS sets the TLS configuration.
func WithTLS(tls *tls.Config) Option {
	return func(conf *clientv3.Config) {
		conf.TLS = tls
	}
}

// WithUsernameAndPassword sets the username and password.
func WithUsernameAndPassword(username, password string) Option {
	return func(conf *clientv3.Config) {
		conf.Username = username
		conf.Password = password
	}
}

// WithRejectOldCluster sets the reject old cluster flag.
func WithRejectOldCluster(reject bool) Option {
	return func(conf *clientv3.Config) {
		conf.RejectOldCluster = reject
	}
}

// WithDialOptions sets the dial options.
func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(conf *clientv3.Config) {
		conf.DialOptions = opts
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

func setIfNoEmptyAny[T any](dur *T, value string, cast func(any) (T, error)) error {
	if value != "" {
		if v, err := cast(value); err == nil {
			*dur = v
		} else {
			return err
		}
	}
	return nil
}
