////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package client

import (
	"fmt"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/broadcast"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/rounds"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/xx_network/crypto/signature/rsa"
	"gitlab.com/xx_network/primitives/utils"
	"regexp"
	"time"
)

// Error messages.
const (
	// WriteChannel
	errMarshalChannel     = "failed to marshal channel: %+v"
	errWriteChannelFile   = "could not write file: %+v"
	errWriteChannelStdout = "could not write to stdout: %+v"

	// LoadChannel
	errReadChannelFile  = "failed to read channel data from file: %+v"
	errUnmarshalChannel = "failed to unmarshal channel: %+v"

	// WriteRsaPrivateKey
	errWriteRsaPrivKeyFile = "could not write file: %+v"

	// ReadRsaPrivateKey
	errReadRsaPrivKeyFile    = "failed to read RSA private key data from file: %+v"
	errLoadPrivateKeyFromPem = "could not load RSA private key from PEM: %+v"

	// SymmetricBroadcastFn
	errNewSymmetricMessage = "failed to create new symmetric message with payload and metadata: %+v"
	errNewSymmetricSized   = "failed to make new symmetric sized broadcast message: %+v"
	errSymmetricBroadcast  = "failed to broadcast symmetric payload: %+v"

	// AsymmetricBroadcastFn
	errNewAsymmetricMessage = "failed to create new asymmetric message with payload and metadata: %+v"
	errNewAsymmetricSized   = "failed to make new asymmetric sized broadcast message: %+v"
	errAsymmetricBroadcast  = "failed to broadcast asymmetric payload: %+v"
)

// WriteChannel serialises and write the channel to the given file path. If no
// path is supplied, it is printed to stdout.
func WriteChannel(path string, s *crypto.Channel) error {
	data, err := s.Marshal()
	if err != nil {
		return errors.Errorf(errMarshalChannel, err)
	}

	if path != "" {
		err = utils.WriteFileDef(path, data)
		if err != nil {
			return errors.Errorf(errWriteChannelFile, err)
		}
	} else {
		_, err = fmt.Printf("%s", data)
		if err != nil {
			return errors.Errorf(errWriteChannelStdout, err)
		}
	}

	jww.INFO.Printf("Saved new channel %q to file %q.", s.Name, path)

	return nil
}

// LoadChannel loads the data from the given file path and deserializes it into
// a channel.
func LoadChannel(path string) (*crypto.Channel, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf(errReadChannelFile, err)
	}

	c, err := crypto.UnmarshalChannel(data)
	if err != nil {
		return nil, errors.Errorf(errUnmarshalChannel, err)
	}

	jww.DEBUG.Printf(
		"Loaded channel %q from file: %+v", c.Name, c)

	return c, nil
}

// WriteRsaPrivateKey writes the RSA private key PEM to the given file path. If
// not file path is supplied, the channel name is used instead.
func WriteRsaPrivateKey(path, channelName string, pk *rsa.PrivateKey) error {
	if path == "" {
		match := regexp.MustCompile("\\s")
		path = match.ReplaceAllString(channelName, "") + "-privateKey.pem"
	}

	pem := rsa.CreatePrivateKeyPem(pk)

	err := utils.WriteFileDef(path, pem)
	if err != nil {
		return errors.Errorf(errWriteRsaPrivKeyFile, err)
	}

	jww.INFO.Printf("Saved new channel RSA private key to file %q.", path)

	return nil
}

// ReadRsaPrivateKey reads the RSA private key PEM from the given file path. If
// not file path is supplied, the channel name is used instead.
func ReadRsaPrivateKey(path, channelName string) (*rsa.PrivateKey, error) {
	if path == "" {
		match := regexp.MustCompile("\\s")
		path = match.ReplaceAllString(channelName, "") + "-privateKey.pem"
	}

	data, err := utils.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf(errReadRsaPrivKeyFile, err)
	}

	pk, err := rsa.LoadPrivateKeyFromPem(data)
	if err != nil {
		return nil, errors.Errorf(errLoadPrivateKeyFromPem, err)
	}

	jww.INFO.Printf("Loaded channel RSA private key to file %q.", path)

	return pk, nil
}

// ReceivedBroadcast contains the broadcast message and its metadata.
type ReceivedBroadcast struct {
	Tag       Tag
	Timestamp time.Time
	Username  string
	Message   []byte
}

// ReceptionCallback generates the listener callback function that the broadcast
// channel delivers new payloads on. Also returns a channel that receives all
// received broadcast messages for the UI to use to print messages.
func ReceptionCallback() (broadcast.ListenerFunc, chan ReceivedBroadcast) {
	cbChan := make(chan ReceivedBroadcast, 100)
	cb := func(payload []byte, ephID receptionID.EphemeralIdentity,
		round rounds.Round) {
		jww.INFO.Printf("Received broadcast message from %s (%d) on round %d: %q",
			ephID.Source, ephID.EphId.Int64(), round.ID, payload)

		decodedPayload, err := broadcast.DecodeSizedBroadcast(payload)
		if err != nil {
			jww.ERROR.Printf("Failed to decode sized broadcast: %+v", err)
		}

		tag, timestamp, username, payload := UnmarshalMessage(decodedPayload)

		cbChan <- ReceivedBroadcast{
			Tag:       tag,
			Timestamp: timestamp,
			Username:  username,
			Message:   payload,
		}
	}
	return cb, cbChan
}

// BroadcastFn allows the UI to pass the message and its metadata to the
// broadcast client.
type BroadcastFn func(tag Tag, timestamp time.Time, message []byte) error

// SymmetricBroadcastFn returns the BroadcastFn used to broadcast symmetric
// broadcast messages.
func SymmetricBroadcastFn(c broadcast.Channel, username string) (BroadcastFn, int) {
	// Get the maximum payload size; dependent on symmetric or asymmetric
	maxSymmetric := c.MaxPayloadSize()
	maxSized := broadcast.MaxSizedBroadcastPayloadSize(maxSymmetric)
	maxPayloadSize := MaxMessagePayloadSize(maxSized, username)

	broadcastFn := func(tag Tag, timestamp time.Time, message []byte) error {
		message, err := NewMessage(maxSized, tag, timestamp, username, message)
		if err != nil {
			return errors.Errorf(errNewSymmetricMessage, err)
		}

		payload, err := broadcast.NewSizedBroadcast(maxSymmetric, message)
		if err != nil {
			return errors.Errorf(errNewSymmetricSized, err)
		}

		cMixParams := cmix.GetDefaultCMIXParams()

		round, ephID, err := c.Broadcast(payload, cMixParams)
		if err != nil {
			return errors.Errorf(errSymmetricBroadcast, err)
		}

		jww.INFO.Printf(
			"Broadcasted symmetric payload on round %s to ephemeral ID %d",
			round.String(), ephID.Int64())

		return nil
	}

	return broadcastFn, maxPayloadSize
}

// AsymmetricBroadcastFn returns the BroadcastFn used to broadcast asymmetric
// broadcast messages.
func AsymmetricBroadcastFn(
	c broadcast.Channel, username string, pk *rsa.PrivateKey) (BroadcastFn, int) {
	// Get the maximum payload size; dependent on symmetric or asymmetric
	maxAsymmetric := c.MaxPayloadSize()
	maxSized := broadcast.MaxSizedBroadcastPayloadSize(maxAsymmetric)
	maxPayloadSize := MaxMessagePayloadSize(maxSized, username)

	broadcastFn := func(tag Tag, timestamp time.Time, message []byte) error {
		message, err := NewMessage(maxSized, tag, timestamp, username, message)
		if err != nil {
			return errors.Errorf(errNewAsymmetricMessage, err)
		}

		payload, err := broadcast.NewSizedBroadcast(maxAsymmetric, message)
		if err != nil {
			return errors.Errorf(errNewAsymmetricSized, err)
		}

		cMixParams := cmix.GetDefaultCMIXParams()

		round, ephID, err := c.BroadcastAsymmetric(pk, payload, cMixParams)
		if err != nil {
			return errors.Errorf(errAsymmetricBroadcast, err)
		}

		jww.INFO.Printf(
			"Broadcasted asymmetric payload on round %s to ephemeral ID %d",
			round.String(), ephID.Int64())

		return nil
	}

	return broadcastFn, maxPayloadSize
}
