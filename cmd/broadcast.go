////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"git.xx.network/elixxir/cli-client/client"
	"git.xx.network/elixxir/cli-client/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gitlab.com/elixxir/client/api"
	"gitlab.com/elixxir/client/broadcast"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/netTime"
	"time"
)

func loadingDots(quit chan struct{}) {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-quit:
			fmt.Print("\n")
			ticker.Stop()
			return
		case <-ticker.C:
			fmt.Print(".")
		}
	}
}

var bCast = &cobra.Command{
	Use:   "broadcast {--new | --load} -o file [-n name -d description | -u username]",
	Short: "Create or join broadcast channels.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging and print version
		initLog(viper.GetString("logPath"), viper.GetInt("logLevel"))
		jww.INFO.Printf(Version())

		streamGen := fastRNG.NewStreamGenerator(12, 1024, csprng.NewSystemRNG)
		filePath := viper.GetString("open")

		// Print a usage error if neither new nor load flags are set
		if !viper.IsSet("new") && !viper.IsSet("load") {
			printUsageError(cmd,
				errors.Errorf("required flag %q or %q not set", "new", "load"))
		}

		// Generate new channel
		if viper.GetBool("new") {
			name := viper.GetString("name")
			description := viper.GetString("description")

			// Create and save new channel if no username is supplied
			rng := streamGen.GetStream()
			channel, rsaPrivKey, err := crypto.NewChannel(name, description, rng)
			if err != nil {
				jww.FATAL.Panicf("Could not make new channel: %+v", err)
			}
			rng.Close()

			jww.INFO.Printf(
				"Generated new channel %q: %+v", name, channel)

			err = client.WriteChannel(filePath, channel)
			if err != nil {
				jww.FATAL.Panicf("Could not write channel to file: %+v", err)
			}

			err = client.WriteRsaPrivateKey(
				viper.GetString("key"), name, rsaPrivKey)
			if err != nil {
				jww.FATAL.Panicf(
					"Could not write RSA private key to file: %+v", err)
			}

			return
		}

		// Join existing channel
		if viper.GetBool("load") {

			quit := make(chan struct{})
			go loadingDots(quit)

			// Initialise a new client
			var cMixClient *api.Client
			var broadcastClient broadcast.Client
			var err error
			if viper.GetBool("test") {
				// Initialise mock client for testing the UI
				broadcastClient = newMockCmix(newMockCmixHandler())
				jww.INFO.Printf("Initialised mock client for testing.")
			} else {
				// Initialise the real client
				cMixClient, err = client.InitClient(
					parsePassword(viper.GetString("password")),
					viper.GetString("session"),
					viper.GetString("ndf"),
				)
				if err != nil {
					jww.FATAL.Panicf("Failed to initialise client: %+v", err)
				}
				broadcastClient = cMixClient.GetNetworkInterface()
				jww.INFO.Printf("Initialised client.")
			}

			// Load channel from file
			channel, err := client.LoadChannel(filePath)
			if err != nil {
				jww.FATAL.Panicf("Could not load channel from file: %+v", err)
			}

			cb, cbChan := client.ReceptionCallback()

			symParams := broadcast.Param{Method: broadcast.Symmetric}
			symClient, err := broadcast.NewBroadcastChannel(
				*channel, cb, broadcastClient, streamGen, symParams)
			if err != nil {
				jww.FATAL.Panicf(
					"Failed to start new symmetric broadcast client: %+v", err)
			}

			asymParams := broadcast.Param{Method: broadcast.Asymmetric}
			asymClient, err := broadcast.NewBroadcastChannel(
				*channel, cb, broadcastClient, streamGen, asymParams)
			if err != nil {
				jww.FATAL.Panicf(
					"Failed to start new asymmetric broadcast client: %+v", err)
			}

			// Connect to the network
			if !viper.GetBool("test") {
				waitTimeout := viper.GetDuration("waitTimeout")
				err = client.ConnectToNetwork(cMixClient, waitTimeout)
				if err != nil {
					jww.FATAL.Panicf("Failed to connect to network: %+v", err)
				}
			}

			username := viper.GetString("username")

			symBroadcastFn, maxPayloadSize := client.SymmetricBroadcastFn(
				symClient, username)

			// Load RSA private key from file
			if viper.IsSet("admin") {
				message := viper.GetString("admin")
				privateKey, err := client.ReadRsaPrivateKey(
					viper.GetString("key"), channel.Name)
				if err != nil {
					jww.FATAL.Panicf("Cannot join channel as admin. Cannot "+
						"get RSA private key: %+v", err)
				}

				asymBroadcastFn, _ := client.AsymmetricBroadcastFn(
					asymClient, username, privateKey)

				err = asymBroadcastFn(client.Admin, netTime.Now(), []byte(message))
				quit <- struct{}{}
				if err != nil {
					jww.FATAL.Panicf("Failed to send message as admin on "+
						"asymmetric channel: %+v", err)
				}
			} else {
				quit <- struct{}{}
				ui.MakeUI(cbChan, symBroadcastFn, channel, username, maxPayloadSize)
			}

			if !viper.GetBool("test") {
				// Stop network follower
				if cMixClient != nil {
					err := cMixClient.StopNetworkFollower()
					if err != nil {
						jww.WARN.Printf(
							"Failed to stop network follower: %+v", err)
					}
				}
			}
		}
	},
}

// init is the initialization function for Cobra which defines commands and
// flags.
func init() {
	bCast.Flags().Bool("test", false,
		"Skips creating a client and connecting to network so that the UI "+
			"can be tested on its own.")
	bindPFlag(bCast.Flags(), "test", bCast.Use)
	hidePFlag(bCast.Flags(), "test", bCast.Use)

	bCast.Flags().Bool("new", false,
		"Creates a new broadcast channel with the specified name and "+
			"description.")
	bindPFlag(bCast.Flags(), "new", bCast.Use)

	bCast.Flags().Bool("load", false,
		"Joins an existing broadcast channel.")
	bindPFlag(bCast.Flags(), "load", bCast.Use)

	bCast.Flags().StringP("name", "n", "",
		"The name of the channel.")
	bindPFlag(bCast.Flags(), "name", bCast.Use)

	bCast.Flags().StringP("description", "d", "",
		"Description of the channel.")
	bindPFlag(bCast.Flags(), "description", bCast.Use)

	bCast.Flags().StringP("open", "o", "",
		"Location to output/open channel information file. Prints to stdout "+
			"if no path is supplied.")
	bindPFlag(bCast.Flags(), "open", bCast.Use)

	bCast.Flags().StringP("key", "k", "",
		"Location to save/load the RSA private key PEM file. Uses the name of "+
			"the channel if no path is supplied.")
	bindPFlag(bCast.Flags(), "key", bCast.Use)

	bCast.Flags().StringP("admin", "a", "",
		"Sends the given message as an admin. Either an RSA private key PEM "+
			"file exists in the default location or one must be specified with "+
			"the \"key\" flag.")
	bindPFlag(bCast.Flags(), "admin", bCast.Use)

	bCast.Flags().StringP("username", "u", "",
		"Join the channel with this username.")
	bindPFlag(bCast.Flags(), "username", bCast.Use)
}
