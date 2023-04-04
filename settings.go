package evo

import (
	"time"
)

type DatabaseConfig struct {
	Enabled            bool          `description:"Enabled database" default:"false" json:"enabled" yaml:"enabled"`
	Type               string        `description:"Database engine" default:"sqlite" json:"type" yaml:"type"`
	Username           string        `description:"Username" default:"root" json:"username" yaml:"username"`
	Password           string        `description:"Password" default:"" json:"password" yaml:"password"`
	Server             string        `description:"Server" default:"127.0.0.1:3306" json:"server" yaml:"server"`
	Cache              string        `description:"Enabled query cache" default:"false" json:"cache" yaml:"cache"`
	Debug              int           `description:"Debug level (1-4)" default:"3" params:"{\"min\":1,\"max\":4}" json:"debug" yaml:"debug"`
	Database           string        `description:"Database Name" default:"" json:"database" yaml:"database"`
	SSLMode            string        `description:"SSL Mode (required by some DBMS)" default:"false" json:"ssl-mode" yaml:"ssl-mode"`
	Params             string        `description:"Extra connection string parameters" default:"" json:"params" yaml:"params"`
	MaxOpenConns       int           `description:"Max pool connections" default:"100" json:"max-open-connections" yaml:"max-open-connections"`
	MaxIdleConns       int           `description:"Max idle connections in pool" default:"10" json:"max-idle-connections" yaml:"max-idle-connections"`
	ConnMaxLifTime     time.Duration `description:"Max connection lifetime" default:"1h" json:"connection-max-lifetime" yaml:"connection-max-lifetime"`
	SlowQueryThreshold time.Duration `description:"Slow query threshold" default:"500ms" json:"slow_query_threshold" yaml:"slow-query-threshold"`
}

type HTTPConfig struct {
	Host string `description:"host listening interface" default:"0.0.0.0" json:"host" yaml:"host"`
	Port string `description:"port" default:"8080" json:"port" yaml:"port"`

	// When set to true, this will spawn multiple Go processes listening on the same port.
	//
	// Default: false
	Prefork bool `description:"When set to true, this will spawn multiple Go processes listening on the same port." default:"false" yaml:"prefork" json:"prefork"`

	// Enables the "Server: value" HTTP header.
	//
	// Default: ""
	ServerHeader string `description:"Enables the \"Server: value\" HTTP header." default:"EVO NG" yaml:"server_header" json:"server_header"`

	// When set to true, the router treats "/foo" and "/foo/" as different.
	// By default this is disabled and both "/foo" and "/foo/" will execute the same handler.
	//
	// Default: false
	StrictRouting bool `description:"When set to true, the router treats /foo and /foo/ as different." default:"false" yaml:"strict_routing" json:"strict_routing"`

	// When set to true, enables case sensitive routing.
	// E.g. "/FoO" and "/foo" are treated as different routes.
	// By default this is disabled and both "/FoO" and "/foo" will execute the same handler.
	//
	// Default: false
	CaseSensitive bool `description:"When set to true, enables case sensitive routing." default:"false" yaml:"case_sensitive" json:"case_sensitive"`

	// When set to true, this relinquishes the 0-allocation promise in certain
	// cases in order to access the handler values (e.g. request bodies) in an
	// immutable fashion so that these values are available even if you return
	// from handler.
	//
	// Default: false
	Immutable bool `description:"When set to true, this relinquishes the 0-allocation promise in certain cases in order to access the handler values" default:"false" yaml:"immutable" json:"immutable"`

	// When set to true, converts all encoded characters in the route back
	// before setting the path for the context, so that the routing,
	// the returning of the current url from the context `ctx.Path()`
	// and the paramters `ctx.Params(%key%)` with decoded characters will work
	//
	// Default: false
	UnescapePath bool `description:"When set to true, converts all encoded characters in the route back before setting the path for the context" default:"false" yaml:"unescape_path" json:"unescape_path"`

	// Enable or disable ETag header generation, since both weak and strong etags are generated
	// using the same hashing method (CRC-32). Weak ETags are the default when enabled.
	//
	// Default: false
	ETag bool `description:"Enable or disable ETag header generation" default:"false" yaml:"etag" json:"etag"`

	// Max body size that the server accepts.
	// -1 will decline any body size
	//
	// Default: 4 * 1024 * 1024
	BodyLimit int `description:"Max body size that the server accepts. -1 will decline any body size." default:"128kb" yaml:"body_limit" json:"body_limit"`

	// Maximum number of concurrent connections.
	//
	// Default: 256 * 1024
	Concurrency int `description:"Maximum number of concurrent connections." default:"1024" yaml:"concurrency" json:"concurrency"`

	// The amount of time allowed to read the full request including body.
	// It is reset after the request handler has returned.
	// The connection's read deadline is reset when the connection opens.
	//
	// Default: unlimited
	ReadTimeout time.Duration `description:"The amount of time allowed to read the full request including body." default:"1s" yaml:"read_timeout" json:"read_timeout"`

	// The maximum duration before timing out writes of the response.
	// It is reset after the request handler has returned.
	//
	// Default: unlimited
	WriteTimeout time.Duration `description:"The maximum duration before timing out writes of the response." default:"5s" yaml:"write_timeout" json:"write_timeout"`

	// The maximum amount of time to wait for the next request when keep-alive is enabled.
	// If IdleTimeout is zero, the value of ReadTimeout is used.
	//
	// Default: unlimited
	IdleTimeout time.Duration `description:"The maximum amount of time to wait for the next request when keep-alive is enabled." default:"0" yaml:"idle_timeout" json:"idle_timeout"`

	// Per-connection buffer size for requests' reading.
	// This also limits the maximum header size.
	// Increase this buffer if your clients send multi-KB RequestURIs
	// and/or multi-KB headers (for example, BIG cookies).
	//
	// Default: 4096
	ReadBufferSize int `description:"Per-connection buffer size for requests' reading." default:"8kb" yaml:"read_buffer_size" json:"read_buffer_size"`

	// Per-connection buffer size for responses' writing.
	//
	// Default: 4096
	WriteBufferSize int `description:"Per-connection buffer size for responses' writing." default:"4kb" yaml:"write_buffer_size" json:"write_buffer_size"`

	// CompressedFileSuffix adds suffix to the original file name and
	// tries saving the resulting compressed file under the new file name.
	//
	// Default: ".fiber.gz"
	CompressedFileSuffix string `description:"CompressedFileSuffix adds suffix to the original file name and tries saving the resulting compressed file under the new file name." default:".evo.gz" yaml:"compressed_file_suffix" json:"compressed_file_suffix"`

	// ProxyHeader will enable c.IP() to return the value of the given header key
	// By default c.IP() will return the Remote IP from the TCP connection
	// This property can be useful if you are behind a load balancer: X-Forwarded-*
	// NOTE: headers are easily spoofed and the detected IP addresses are unreliable.
	//
	// Default: ""
	ProxyHeader string `description:"ProxyHeader will enable c.IP() to return the value of the given header key" default:"X-Forwarded-For" yaml:"proxy_header" json:"proxy_header"`

	// GETOnly rejects all non-GET requests if set to true.
	// This option is useful as anti-DoS protection for servers
	// accepting only GET requests. The request size is limited
	// by ReadBufferSize if GETOnly is set.
	//
	// Default: false
	GETOnly bool `description:"GETOnly rejects all non-GET requests if set to true." default:"false" yaml:"get_only" json:"get_only"`

	// When set to true, disables keep-alive connections.
	// The server will close incoming connections after sending the first response to client.
	//
	// Default: false
	DisableKeepalive bool `description:"When set to true, disables keep-alive connections." default:"false" yaml:"disable_keepalive" json:"disable_keepalive"`

	// When set to true, causes the default date header to be excluded from the response.
	//
	// Default: false
	DisableDefaultDate bool `description:"When set to true, causes the default date header to be excluded from the response." default:"false" yaml:"disable_default_date" json:"disable_default_date"`

	// When set to true, causes the default Content-Type header to be excluded from the response.
	//
	// Default: false
	DisableDefaultContentType bool `description:"When set to true, causes the default Content-Type header to be excluded from the response." default:"false" yaml:"disable_default_content_type" json:"disable_default_content_type"`

	// When set to true, disables header normalization.
	// By default all header names are normalized: conteNT-tYPE -> Content-Type.
	//
	// Default: false
	DisableHeaderNormalizing bool `description:"When set to true, disables header normalization." default:"false" yaml:"disable_header_normalizing" json:"disable_header_normalizing"`

	// Aggressively reduces memory usage at the cost of higher CPU usage
	// if set to true.
	//
	// Try enabling this option only if the server consumes too much memory
	// serving mostly idle keep-alive connections. This may reduce memory
	// usage by more than 50%.
	//
	// Default: false
	ReduceMemoryUsage bool `description:"Aggressively reduces memory usage at the cost of higher CPU usage" default:"false" yaml:"reduce_memory_usage" json:"reduce_memory_usage"`

	// FEATURE: v2.3.x
	// The router executes the same handler by default if StrictRouting or CaseSensitive is disabled.
	// Enabling RedirectFixedPath will change this behaviour into a client redirect to the original route path.
	// Using the status code 301 for GET requests and 308 for all other request methods.
	//
	// Default: false
	// RedirectFixedPath bool

	// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only)
	// WARNING: When prefork is set to true, only "tcp4" and "tcp6" can be chose.
	//
	// Default: NetworkTCP4
	Network string `description:"Known networks are tcp, tcp4 (IPv4-only), tcp6 (IPv6-only)" default:"tcp4" yaml:"network" json:"network"`

	// If set to true, will print all routes with their method, path and handler.
	// Default: false
	EnablePrintRoutes bool `description:"If set to true, will print all routes with their method, path and handler." default:"false" yaml:"enable_print_routes" json:"enable_print_routes"`
}
