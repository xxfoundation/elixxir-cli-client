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
)

// WriteChannel serialises and write the channel to the given file path. If no
// path is supplied, it is printed to stdout.
func WriteChannel(path string, s *crypto.Channel) error {
	data, err := s.Marshal()
	if err != nil {
		return errors.Errorf("failed to marshal channel: %+v", err)
	}

	if path != "" {
		err = utils.WriteFileDef(path, data)
		if err != nil {
			return errors.Errorf("could not write file: %+v", err)
		}
	} else {
		_, err = fmt.Printf("%s", data)
		if err != nil {
			return errors.Errorf("could not write to stdout: %+v", err)
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
		return nil, errors.Errorf(
			"failed to read channel data from file: %+v", err)
	}

	c, err := crypto.UnmarshalChannel(data)
	if err != nil {
		return nil, errors.Errorf(
			"failed to unmarshal channel: %+v", err)
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
		return errors.Errorf("could not write file: %+v", err)
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
		return nil, errors.Errorf(
			"failed to read RSA private key data from file: %+v", err)
	}

	pk, err := rsa.LoadPrivateKeyFromPem(data)
	if err != nil {
		return nil, errors.Errorf(
			"could not load RSA private key from PEM: %+v", err)
	}

	jww.INFO.Printf("Loaded channel RSA private key to file %q.", path)

	return pk, nil
}

func ReceptionCallback() (broadcast.ListenerFunc, chan []byte) {
	cbChan := make(chan []byte, 100)
	cb := func(payload []byte, ephID receptionID.EphemeralIdentity,
		round rounds.Round) {
		jww.INFO.Printf("Received broadcast message from %d on round %d: %q",
			ephID.EphId.Int64(), round.ID, payload)
		decodedPayload, err := broadcast.DecodeSizedBroadcast(payload)
		if err != nil {
			jww.ERROR.Printf("Failed to decode sized broadcast: %+v", err)
		}
		cbChan <- decodedPayload
	}
	return cb, cbChan
}

func SymmetricBroadcastFn(c broadcast.Channel, username string) (
	func(message []byte) error, int) {
	usernameTag := username + ":\xb1"

	// Get the maximum payload size; dependent on symmetric or asymmetric
	maxPayloadSize := broadcast.MaxSizedBroadcastPayloadSize(
		c.MaxSymmetricPayloadSize()) - len(usernameTag)

	broadcastFn := func(message []byte) error {
		payload, err := broadcast.NewSizedBroadcast(
			c.MaxSymmetricPayloadSize(), []byte(usernameTag+string(message)))
		if err != nil {
			return errors.Errorf(
				"failed to make new sized broadcast message: %+v", err)
		}

		cMixParams := cmix.GetDefaultCMIXParams()

		round, ephID, err := c.Broadcast(payload, cMixParams)
		if err != nil {
			return errors.Errorf(
				"failed to broadcast symmetric payload: %+v", err)
		}

		jww.INFO.Printf(
			"Broadcasted symmetric payload on round %s to ephemeral ID %d",
			round.String(), ephID.Int64())

		return nil
	}

	return broadcastFn, maxPayloadSize
}

func AsymmetricBroadcastFn(c broadcast.Channel, username string,
	pk *rsa.PrivateKey) (func(message []byte) error, int) {
	usernameTag := username + ":\xb1"

	// Get the maximum payload size; dependent on symmetric or asymmetric
	maxPayloadSize := broadcast.MaxSizedBroadcastPayloadSize(
		c.MaxAsymmetricPayloadSize()) - len(usernameTag)

	broadcastFn := func(message []byte) error {
		payload, err := broadcast.NewSizedBroadcast(
			c.MaxAsymmetricPayloadSize(), []byte(usernameTag+string(message)))
		if err != nil {
			return errors.Errorf(
				"failed to make new sized broadcast message: %+v", err)
		}

		cMixParams := cmix.GetDefaultCMIXParams()
		round, ephID, err := c.BroadcastAsymmetric(pk, payload, cMixParams)
		if err != nil {
			return errors.Errorf(
				"failed to broadcast asymmetric payload: %+v", err)
		}

		jww.INFO.Printf(
			"Broadcasted asymmetric payload on round %s to ephemeral ID %d",
			round.String(), ephID.Int64())

		return nil
	}

	return broadcastFn, maxPayloadSize
}
