package publisher_test

import (
	"testing"
	"time"

	natsgo "github.com/nats-io/nats.go"
	natsserver "github.com/nats-io/nats-server/v2/server"

	"github.com/mathrmm/watchdog-monitor/internal/publisher"
)

// TestConnect_NATSUnavailable verifies that Connect() returns an error (not panic)
// when no NATS server is listening.
func TestConnect_NATSUnavailable(t *testing.T) {
	_, err := publisher.Connect("nats://127.0.0.1:54321")
	if err == nil {
		t.Error("expected error connecting to unavailable NATS, got nil")
	}
}

// TestPublish_NATSAvailable verifies that Publish() delivers a message to a subscriber.
func TestPublish_NATSAvailable(t *testing.T) {
	// Start embedded NATS server on a random port.
	opts := &natsserver.Options{Port: -1}
	s, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("failed to create embedded NATS: %v", err)
	}
	go s.Start()
	if !s.ReadyForConnections(2 * time.Second) {
		t.Fatal("embedded NATS server not ready")
	}
	defer s.Shutdown()

	url := s.ClientURL()

	pub, err := publisher.Connect(url)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer pub.Close()

	// Independent subscriber to verify delivery.
	nc, err := natsgo.Connect(url)
	if err != nil {
		t.Fatalf("subscriber connect failed: %v", err)
	}
	defer nc.Close()

	received := make(chan []byte, 1)
	if _, err := nc.Subscribe("test.subject", func(msg *natsgo.Msg) {
		received <- msg.Data
	}); err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	// Flush ensures the subscription is registered on the server before publishing.
	if err := nc.Flush(); err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	want := []byte("hello-watchdog")
	if err := pub.Publish("test.subject", want); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case got := <-received:
		if string(got) != string(want) {
			t.Errorf("expected %q, got %q", want, got)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message delivery")
	}
}
