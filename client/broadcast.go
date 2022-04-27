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
	cMixCrypto "gitlab.com/elixxir/crypto/cmix"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/xx_network/crypto/signature/rsa"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/utils"
)

// NewSymmetric generates a new crypto.Symmetric channel with the given name and
// description. The reception ID and, salt, and RSA key are randomly generated.
func NewSymmetric(name, description string,
	rngGen *fastRNG.StreamGenerator) (crypto.Symmetric, error) {
	if name == "" {
		return crypto.Symmetric{}, errors.New("name cannot be empty")
	} else if description == "" {
		return crypto.Symmetric{}, errors.New("description cannot be empty")
	}

	rng := rngGen.GetStream()
	defer rng.Close()

	rsaPrivKey, err := rsa.GenerateKey(rng, 512)
	if err != nil {
		return crypto.Symmetric{}, errors.Errorf(
			"failed to generate RSA key for new symmetric channel: %+v", err)
	}

	rid, err := id.NewRandomID(rng, id.User)
	if err != nil {
		return crypto.Symmetric{}, errors.Errorf(
			"failed to generate reception ID for new symmetric channel: %+v", err)
	}

	return crypto.Symmetric{
		ReceptionID: rid,
		Name:        name,
		Description: description,
		Salt:        cMixCrypto.NewSalt(rng, 32),
		RsaPubKey:   rsaPrivKey.GetPublic(),
	}, nil
}

// WriteSymmetric serialises and write the symmetric channel to the given file
// path. If no path is supplied, it is printed to stdout.
func WriteSymmetric(path string, s crypto.Symmetric) error {
	data, err := s.Marshal()
	if err != nil {
		return errors.Errorf("failed to marshal symmetric channel: %+v", err)
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

	return nil
}

// LoadSymmetricChannel loads the data from the given file path and deserializes
// it into a symmetric channel.
func LoadSymmetricChannel(path string) (*crypto.Symmetric, error) {
	data, err := utils.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf(
			"failed to read symmetric channel data from file: %+v", err)
	}

	s, err := crypto.UnmarshalSymmetric(data)
	if err != nil {
		return nil, errors.Errorf(
			"failed to unmarshal symmetric channel: %+v", err)
	}

	return s, nil
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

func BroadcastFn(s broadcast.Symmetric, username string, net broadcast.Client) (
	func(message []byte) error, int) {
	usernameTag := username + ":\xb1"
	maxPayloadSize := broadcast.MaxSizedBroadcastPayloadSize(
		net.GetMaxMessageLength()) - len(usernameTag)
	broadcastFn := func(message []byte) error {
		payload, err := broadcast.NewSizedBroadcast(
			net.GetMaxMessageLength(),
			[]byte(usernameTag+string(message)))
		if err != nil {
			return errors.Errorf("failed to make new sized "+
				"broadcast message: %+v", err)
		}

		params := cmix.GetDefaultCMIXParams()
		params.DebugTag = "SymmChannel"

		round, ephID, err := s.Broadcast(payload, params)
		if err != nil {
			return errors.Errorf(
				"failed to broadcast payload: %+v", err)
		}

		jww.INFO.Printf(
			"Broadcasted payload on round %s to ephemeral ID %d",
			round.String(), ephID.Int64())

		return nil
	}
	return broadcastFn, maxPayloadSize
}
