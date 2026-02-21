package nats

// Params holds parsed options for NATS driver calls.
type Params struct {
	Bucket       string
	IgnorePrefix bool
	QueueGroup   string
	UseJetStream bool
}

// typed wrappers so callers get compile-time safety
type bucket string
type queue string

// Named integer constants accepted by Parse.
const (
	IgnorePrefix = 1 // skip the configured topic prefix
	WithJetStream = 2 // use JetStream publish (at-least-once delivery with ack)
)

// Bucket returns a typed bucket name to be passed as a variadic param.
func Bucket(s string) bucket { return bucket(s) }

// Queue returns a typed queue-group name for queue-subscribe load balancing.
func Queue(s string) queue { return queue(s) }

// Parse extracts a Params value from a heterogeneous variadic list.
func Parse(params []any) Params {
	var p Params
	for idx := range params {
		switch v := params[idx].(type) {
		case int:
			if v == IgnorePrefix {
				p.IgnorePrefix = true
			}
			if v == WithJetStream {
				p.UseJetStream = true
			}
		case Params:
			return v
		case string:
			p.Bucket = v
		case bucket:
			p.Bucket = string(v)
		case queue:
			p.QueueGroup = string(v)
		}
	}
	return p
}
