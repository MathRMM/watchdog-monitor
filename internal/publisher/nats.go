package publisher

import (
	"fmt"

	natsgo "github.com/nats-io/nats.go"
)

// Publisher wraps a NATS connection and exposes a single Publish method.
type Publisher struct {
	conn *natsgo.Conn
}

// Connect establishes a connection to the NATS server at url.
// Automatic reconnection is enabled with no limit (nats.MaxReconnects(-1)).
// If the server is unavailable at connect time, returns an error immediately.
func Connect(url string) (*Publisher, error) {
	conn, err := natsgo.Connect(url, natsgo.MaxReconnects(-1))
	if err != nil {
		return nil, fmt.Errorf("nats connect %s: %w", url, err)
	}
	return &Publisher{conn: conn}, nil
}

// Publish serializes data to the given NATS subject.
// Returns an error if the publish fails — the caller decides whether to discard (RN03).
func (p *Publisher) Publish(subject string, data []byte) error {
	if err := p.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("nats publish %s: %w", subject, err)
	}
	return nil
}

// Close drains and closes the underlying NATS connection.
func (p *Publisher) Close() {
	p.conn.Close()
}
