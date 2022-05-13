package client

import (
	"bytes"
	"encoding/binary"
	"github.com/graph-gophers/graphql-go/errors"
	"math"
	"time"
)

// Size constants.
const (
	tagSize         = 1
	timestampSize   = 8
	usernameLenSize = 1
	minSize         = tagSize + timestampSize + usernameLenSize
)

// Error messages.
const (
	// NewMessage
	errNewMessageSize = "max size of payload (%d) must be greater than %d"
	errUsernameLen    = "length of username (%d) cannot exceed %d"
	errPayloadLen     = "combined size of payload (%d) cannot exceed %d"
)

/*
+------------------------------------------------------------------------+
|                            Broadcast Payload                           |
+--------+-----------+-------------+-------------------+-----------------+
|  tag   | timestamp | usernameLen |     username      |     payload     |
| 1 byte |  8 bytes  |   1 byte    | usernameLen bytes | remaining bytes |
+--------+-----------+-------------+-------------------+-----------------+
*/

// MaxMessagePayloadSize returns the maximum size of a payload for the given
// max payload size and username.
func MaxMessagePayloadSize(maxPayloadSize int, username string) int {
	return maxPayloadSize - (minSize + len(username))
}

// NewMessage generates a new message containing the payload and the metadata.
func NewMessage(maxPayloadSize int, tag Tag, timestamp time.Time,
	username string, payload []byte) ([]byte, error) {
	if maxPayloadSize < minSize {
		return nil, errors.Errorf(errNewMessageSize, maxPayloadSize, minSize)
	}

	if len(username) > math.MaxUint8 {
		return nil, errors.Errorf(errUsernameLen, len(username), math.MaxUint8)
	}

	payloadSize := minSize + len(username) + len(payload)
	if payloadSize > maxPayloadSize {
		return nil, errors.Errorf(errPayloadLen, payloadSize, maxPayloadSize)
	}

	buff := bytes.NewBuffer(nil)
	buff.Grow(minSize + len(username) + len(payload))

	buff.WriteByte(byte(tag))

	b := make([]byte, timestampSize)
	binary.LittleEndian.PutUint64(b, uint64(timestamp.UnixNano()))
	buff.Write(b)

	buff.WriteByte(uint8(len(username)))
	buff.WriteString(username)
	buff.Write(payload)

	return buff.Bytes(), nil
}

// UnmarshalMessage decodes the data into a payload and its metadata.
func UnmarshalMessage(data []byte) (tag Tag, timestamp time.Time,
	username string, payload []byte) {
	buff := bytes.NewBuffer(data)

	tag = Tag(buff.Next(tagSize)[0])
	timestamp = time.Unix(
		0, int64(binary.LittleEndian.Uint64(buff.Next(timestampSize))))
	usernameLen := int(buff.Next(usernameLenSize)[0])
	username = string(buff.Next(usernameLen))
	payload = buff.Bytes()

	return tag, timestamp, username, payload
}
