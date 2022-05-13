package client

import (
	"bytes"
	"gitlab.com/xx_network/primitives/netTime"
	"testing"
)

// Tests that the payload and metadata encoded with NewMessage can be decoded
// with UnmarshalMessage.
func TestNewMessageUnmarshalMessage(t *testing.T) {
	tag := Join
	timestamp := netTime.Now()
	username := "myUsername"
	payload := []byte("This is my payload.")

	message, err := NewMessage(1024, tag, timestamp, username, payload)
	if err != nil {
		t.Errorf("Failed to create new message: %+v", err)
	}

	receivedTag, receivedTimestamp, receivedUsername, receivedPayload :=
		UnmarshalMessage(message)

	if tag != receivedTag {
		t.Errorf("Received tag does not match expected."+
			"\nexpected: %d\nreceived: %d", tag, receivedTag)
	}

	if !timestamp.Equal(receivedTimestamp) {
		t.Errorf("Received timestamp does not match expected."+
			"\nexpected: %s\nreceived: %s", timestamp, receivedTimestamp)
	}

	if username != receivedUsername {
		t.Errorf("Received username does not match expected."+
			"\nexpected: %q\nreceived: %q", username, receivedUsername)
	}

	if !bytes.Equal(payload, receivedPayload) {
		t.Errorf("Received payload does not match expected."+
			"\nexpected: %q\nreceived: %q", payload, receivedPayload)
	}
}
