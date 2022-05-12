package cmd

import (
	"fmt"
	"git.xx.network/elixxir/cli-client/client"
	"git.xx.network/elixxir/cli-client/ui"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gitlab.com/elixxir/client/api"
	"gitlab.com/elixxir/client/broadcast"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/xx_network/crypto/csprng"
	"os"
)

var bcast = &cobra.Command{
	Use:   "broadcast {--new | --load} -o file [-d description -m name | -u username]",
	Short: "Create or join broadcast channels.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging and print version
		initLog(viper.GetString("logPath"), viper.GetInt("logLevel"))
		jww.INFO.Printf(Version())

		rng := fastRNG.NewStreamGenerator(12, 1024, csprng.NewSystemRNG)
		filePath := viper.GetString("open")

		// Print a usage error if neither new nor load flags are set
		if !viper.IsSet("new") && !viper.IsSet("load") {
			err := fmt.Errorf("required flag %q or %q not set", "new", "load")
			cmd.PrintErrln("Error:", err.Error())
			cmd.Println(cmd.UsageString())
			fmt.Println(err)
			os.Exit(1)
		}

		// Generate new symmetric channel
		if viper.GetBool("new") {
			name := viper.GetString("name")
			description := viper.GetString("description")

			// Create and save new symmetric channel if no username is supplied
			symChannel, err := client.NewSymmetric(name, description, rng)
			if err != nil {
				jww.FATAL.Panicf(
					"Could not make new symmetric channel: %+v", err)
			}

			jww.INFO.Printf(
				"Generated new symmetric channel %q: %+v", name, symChannel)

			err = client.WriteSymmetric(filePath, symChannel)
			if err != nil {
				jww.FATAL.Panicf(
					"Could not write symmetric channel to file: %+v", err)
			}

			jww.INFO.Printf(
				"Saved new symmetric channel %q to file %q.", name, filePath)

			return
		}

		// Join existing symmetric channel
		if viper.GetBool("load") {

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

			// Load symmetric channel from file
			symChannel, err := client.LoadSymmetricChannel(filePath)
			if err != nil {
				jww.FATAL.Panicf(
					"Could not load symmetric channel from file: %+v", err)
			}
			jww.DEBUG.Printf("Loaded symmetric channel %q from file: %+v",
				symChannel.Name, symChannel)

			cb, cbChan := client.ReceptionCallback()

			symClient := broadcast.NewSymmetricClient(*symChannel, cb,
				broadcastClient, rng)

			// Connect to the network
			if !viper.GetBool("test") {
				waitTimeout := viper.GetDuration("waitTimeout")
				err = client.ConnectToNetwork(cMixClient, waitTimeout)
				if err != nil {
					jww.FATAL.Panicf("Failed to connect to network: %+v", err)
				}
			}

			username := viper.GetString("username")
			broadcastFn, maxPayloadSize := client.BroadcastFn(
				symClient, username, broadcastClient)

			ui.MakeUI(cbChan, broadcastFn, symChannel.Name, username,
				symChannel.Description, symChannel.ReceptionID, maxPayloadSize)

			if !viper.GetBool("test") {
				// Stop network follower
				if cMixClient != nil {
					err := cMixClient.StopNetworkFollower()
					if err != nil {
						jww.WARN.Printf("Failed to stop network follower: %+v", err)
					}
				}
			}
		}
	},
}

// init is the initialization function for Cobra which defines commands and
// flags.
func init() {
	bcast.Flags().Bool("test", false,
		"Skips creating a client and connecting to network so that the UI "+
			"can be tested on its own.")
	bindPFlag(bcast.Flags(), "test", bcast.Use)

	bcast.Flags().Bool("new", false,
		"Creates a new symmetric broadcast channel with the specified name "+
			"and description.")
	bindPFlag(bcast.Flags(), "new", bcast.Use)

	bcast.Flags().Bool("load", false,
		"Joins an existing symmetric broadcast channel .")
	bindPFlag(bcast.Flags(), "load", bcast.Use)

	bcast.Flags().StringP("name", "m", "",
		"The name of the channel.")
	bindPFlag(bcast.Flags(), "name", bcast.Use)

	bcast.Flags().StringP("description", "d", "",
		"Description of the channel.")
	bindPFlag(bcast.Flags(), "description", bcast.Use)

	bcast.Flags().StringP("open", "o", "",
		"Location to output/open channel information file. Prints to stdout "+
			"if not path is supplied.")
	bindPFlag(bcast.Flags(), "open", bcast.Use)

	bcast.Flags().StringP("username", "u", "",
		"Join the channel with this username.")
	bindPFlag(bcast.Flags(), "username", bcast.Use)
}
