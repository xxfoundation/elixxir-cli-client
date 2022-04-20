package cmd

import (
	"fmt"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gitlab.com/elixxir/client/broadcast"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/identity/receptionID"
	"gitlab.com/elixxir/client/cmix/rounds"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	cmixCrypto "gitlab.com/elixxir/crypto/cmix"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/crypto/signature/rsa"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
	"gitlab.com/xx_network/primitives/utils"
	"strconv"
	"time"
)

var bcast = &cobra.Command{
	Use:   "broadcast",
	Short: "Create or join broadcast channels.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize logging and print version
		initLog(logPath, logLevel)
		jww.INFO.Printf(Version())

		// Initialise a new client
		client := initClient()

		// Start the network follower
		err := client.StartNetworkFollower(5 * time.Second)
		if err != nil {
			jww.FATAL.Panicf("Failed to start the network follower: %+v", err)
		}

		// Wait until connected or crash on timeout
		connected := make(chan bool, 10)
		client.GetHealth().AddChannel(connected)
		waitUntilConnected(connected)

		// After connection, wait until registered with at least 85% of nodes
		for numReg, total := 1, 100; numReg < (total*3)/4; {
			time.Sleep(1 * time.Second)

			numReg, total, err = client.GetNodeRegistrationStatus()
			if err != nil {
				jww.FATAL.Panicf("Failed to get node registration status: %+v",
					err)
			}

			jww.INFO.Printf("Registering with nodes (%d/%d)...", numReg, total)
		}

		rngStreamGen := fastRNG.NewStreamGenerator(12, 1024,
			csprng.NewSystemRNG)

		name := viper.GetString("name")
		description := viper.GetString("description")
		filePath := viper.GetString("open")
		username := viper.GetString("username")

		if viper.GetBool("symmetric") {
			if username == "" {
				// Create and save new symmetric channel if no username is
				// supplied
				rng := rngStreamGen.GetStream()
				symChannel, err := newSymmetricChannel(name, description, rng)
				if err != nil {
					jww.FATAL.Panicf(
						"Could not make new symmetric channel: %+v", err)
				}
				rng.Close()

				err = writeSymmetricChannel(filePath, symChannel)
				if err != nil {
					jww.FATAL.Panicf(
						"Could not write symmetric channel to file: %+v", err)
				}
			} else {
				symChannel, err := loadSymmetricChannel(filePath)
				if err != nil {
					jww.FATAL.Panicf(
						"Could not load symmetric channel from file: %+v", err)
				}

				cb := func(payload []byte, _ receptionID.EphemeralIdentity,
					_ rounds.Round) {
					jww.DEBUG.Printf("Received broadcast message: %q", payload)
				}

				symClient := broadcast.NewSymmetricClient(*symChannel, cb,
					client.GetNetworkInterface(), rngStreamGen)

				round, ephemeralID, err := symClient.Broadcast(
					[]byte("hello"), cmix.GetDefaultCMIXParams())
				if err != nil {
					jww.FATAL.Panicf("Failed to broadcast payload: %+v", err)
					return
				}

				jww.DEBUG.Printf("Round: %+v", round)
				jww.DEBUG.Printf("ephemeral ID: %+v", ephemeralID)
			}

		}

		// Stop network follower
		err = client.StopNetworkFollower()
		if err != nil {
			jww.WARN.Printf("Failed to stop network follower: %+v", err)
		}
	},
}

// newSymmetricChannel generates a new crypto.Symmetric channel with the given
// name and description. The reception ID and, salt, and RSA key are randomly
// generation.
func newSymmetricChannel(
	name, description string, csprng csprng.Source) (crypto.Symmetric, error) {
	rsaPrivKey, err := rsa.GenerateKey(csprng, 512)
	if err != nil {
		return crypto.Symmetric{}, errors.Errorf(
			"Failed to generate RSA key for new symmetric channel: %+v", err)
	}

	rid, err := id.NewRandomID(csprng, id.User)
	if err != nil {
		return crypto.Symmetric{}, errors.Errorf(
			"Failed to generate reception ID for new symmetric channel: %+v", err)
	}

	return crypto.Symmetric{
		ReceptionID: rid,
		Name:        name,
		Description: description,
		Salt:        cmixCrypto.NewSalt(csprng, 32),
		RsaPubKey:   rsaPrivKey.GetPublic(),
	}, nil
}

// writeSymmetricChannel serialises and write the symmetric channel to the given
// file path. If no path is supplied, it is printed to stdout.
func writeSymmetricChannel(path string, s crypto.Symmetric) error {
	data, err := s.Marshal()
	if err != nil {
		return errors.Errorf("failed to marshal symmetric channel: %+v", err)
	}

	if path != "" {
		err = utils.WriteFileDef(path, data)
		if err != nil {
			return errors.Errorf(
				"failed to write symmetric channel to file: %+v", err)
		}
	} else {
		fmt.Printf("%s", data)
	}

	return nil
}

// writeSymmetricChannel serialises and write the symmetric channel to the given
// file path. If no path is supplied, it is printed to stdout.
func loadSymmetricChannel(path string) (*crypto.Symmetric, error) {
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

// init is the initialization function for Cobra which defines commands and
// flags.
func init() {
	rootCmd.PersistentFlags().StringVarP(&logPath, "logPath", "l", "",
		"File path to save log file to.")
	rootCmd.PersistentFlags().IntVarP(&logLevel, "logLevel", "v", 0,
		"Verbosity level for log printing (2+ = Trace, 1 = Debug, 0 = Info).")

	timeNow := strconv.Itoa(int(netTime.Now().UnixNano()))
	bcast.Flags().Bool("symmetric", false,
		"Creates/loads a symmetric broadcast channel")
	bcast.Flags().String("name",
		"Channel-"+timeNow,
		"The name of the channel.")
	bcast.Flags().String("description ",
		"Description of channel made at "+timeNow,
		"Description of the channel.")
	bcast.Flags().StringP("open", "o", "",
		"Location to output/open channel information file. Prints to stdout "+
			"if not path is supplied.")
	bcast.Flags().StringP("username", "u",
		"AnonymousUser",
		"Join the channel with this username.")
}
