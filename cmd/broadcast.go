////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"git.xx.network/elixxir/cli-client/client"
	"git.xx.network/elixxir/cli-client/ui"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"os"
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
	Use: "broadcast -u username",
	Short: "Start the channel TUI that allows for joining and creating " +
		"channels to send and receive messages on.",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Initiate config file
		initConfig(viper.GetString("config"))

		// Initialize logging and print version
		initLog(viper.GetString("logPath"), viper.GetInt("logLevel"))
		jww.INFO.Printf(Version())

		// Start printing loading dots to stdout
		quit := make(chan struct{})
		go loadingDots(quit)

		// Initialise a new client
		net, err := client.InitClient(
			parsePassword(viper.GetString("password")),
			viper.GetString("session"), viper.GetString("ndf"),
		)
		if err != nil {
			jww.FATAL.Panicf("Failed to initialise client: %+v", err)
		} else {
			jww.INFO.Printf("Initialised client.")
		}

		// Connect to the network
		err = client.ConnectToNetwork(net, viper.GetDuration("waitTimeout"))
		if err != nil {
			jww.FATAL.Panicf("Failed to connect to network: %+v", err)
		} else {
			jww.INFO.Printf("Connected to healthy network.")
		}

		username := viper.GetString("nickname")

		quit <- struct{}{}

		m := client.NewChannelManager(net, username)

		chs := ui.NewChannels(m)

		chs.MakeUI()

		err = net.StopNetworkFollower()
		if err != nil {
			jww.ERROR.Printf("Failed to stop network follower: %+v", err)
		}

		jww.INFO.Printf("Stopped network follower.")

		os.Exit(0)
	},
}

// init is the initialization function for Cobra which defines commands and
// flags.
func init() {
	bCast.Flags().StringP("nickname", "n", "",
		"Join the channel with this nickname.")
	bindPFlag(bCast.Flags(), "nickname", bCast.Use)
	markFlagRequired(bCast, "nickname")
}
