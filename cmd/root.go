////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.com/elixxir/client/api"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Execute adds all child commands to the root command and sets the flags
// appropriately. This is called by main.main(). It only needs to happen once to
// the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "cli-client",
	Short: "Command-line interface for client.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logging and print version
		initLog(viper.GetString("logPath"), viper.GetInt("logLevel"))
		jww.INFO.Printf(Version())
	},
}

func initClient() *api.Client {
	pass := parsePassword(viper.GetString("password"))
	storeDir := viper.GetString("session")

	// create a new client if none exist
	if _, err := os.Stat(storeDir); errors.Is(err, fs.ErrNotExist) {
		// Load NDF
		ndfJSON, err := ioutil.ReadFile(viper.GetString("ndf"))
		if err != nil {
			jww.FATAL.Panicf("Failed to read NDF file: %+v", err)
		}

		err = api.NewClient(string(ndfJSON), storeDir, pass, "")
		if err != nil {
			jww.FATAL.Panicf("Failed to create new client: %+v", err)
		}
	}

	// Load the client
	client, err := api.Login(storeDir, pass, nil, api.GetDefaultParams())
	if err != nil {
		jww.FATAL.Panicf("Failed to log in into new client: %+v", err)
	}

	return client
}

// waitUntilConnected waits until the network is connected.
func waitUntilConnected(connected chan bool) {
	waitTimeout := viper.GetDuration("waitTimeout")
	timeoutTimer := time.NewTimer(waitTimeout)
	// Wait until connected or panic after time out is reached
	for isConnected := false; !isConnected; {
		select {
		case isConnected = <-connected:
			jww.INFO.Printf("Network status: %t", isConnected)
		case <-timeoutTimer.C:
			jww.FATAL.Panicf("Timed out after %s while waiting for network "+
				"connection.", waitTimeout)
		}
	}
}

func parsePassword(pwStr string) []byte {
	if strings.HasPrefix(pwStr, "0x") {
		return getPwFromHexString(pwStr[2:])
	} else if strings.HasPrefix(pwStr, "b64:") {
		return getPwFromB64String(pwStr[4:])
	} else {
		return []byte(pwStr)
	}
}

func getPwFromHexString(pwStr string) []byte {
	pwBytes, err := hex.DecodeString(
		fmt.Sprintf("%0*d%s", 66-len(pwStr), 0, pwStr))
	if err != nil {
		jww.FATAL.Panicf("Failed to get password from hex string: %+v", err)
	}
	return pwBytes
}

func getPwFromB64String(pwStr string) []byte {
	pwBytes, err := base64.StdEncoding.DecodeString(pwStr)
	if err != nil {
		jww.FATAL.Panicf("Failed to get password from base 64 string: %+v", err)
	}
	return pwBytes
}

// initLog initializes logging thresholds and the log path. If not path is
// provided, the log output is not set. Possible values for logLevel:
//  0  = info
//  1  = debug
//  2+ = trace
func initLog(logPath string, logLevel int) {
	// Set log level to highest verbosity while setting up log files
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelTrace)

	// Set log file output
	if logPath != "" {
		logFile, err := os.OpenFile(
			logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			jww.ERROR.Printf("Could not open log file %q: %+v\n", logPath, err)
			jww.SetStdoutThreshold(jww.LevelFatal)
		} else {
			jww.INFO.Printf("Setting log output to %q", logPath)
			// jww.SetLogOutput(logFile)
			jww.SetStdoutOutput(logFile)
		}
	} else {
		jww.INFO.Printf("No log output set: no log path provided")
	}

	// Select the level of logs to display
	var threshold jww.Threshold
	if logLevel > 1 {
		// Turn on trace logs
		threshold = jww.LevelTrace
	} else if logLevel == 1 {
		// Turn on debugging logs
		threshold = jww.LevelDebug
	} else {
		// Turn on info logs
		threshold = jww.LevelInfo
	}

	// Set logging thresholds
	jww.SetLogThreshold(threshold)
	jww.SetStdoutThreshold(threshold)
	jww.INFO.Printf("Log level set to: %s", threshold)
}

// init is the initialization function for Cobra which defines commands and
// flags.
func init() {
	rootCmd.Flags()
	rootCmd.PersistentFlags().StringP("logPath", "l", "",
		"File path to save log file to.")
	bindPFlag(rootCmd.PersistentFlags(), "logPath", rootCmd.Use)
	rootCmd.PersistentFlags().IntP("logLevel", "v", 0,
		"Verbosity level for log printing (2+ = Trace, 1 = Debug, 0 = Info).")
	bindPFlag(rootCmd.PersistentFlags(), "logLevel", rootCmd.Use)
	rootCmd.PersistentFlags().StringP("session", "s", "session",
		"Sets the initial storage directory for client session data.")
	bindPFlag(rootCmd.PersistentFlags(), "session", rootCmd.Use)
	rootCmd.PersistentFlags().StringP("password", "p", "",
		"Password to the session file.")
	bindPFlag(rootCmd.PersistentFlags(), "password", rootCmd.Use)
	rootCmd.PersistentFlags().StringP("ndf", "n", "ndf.json",
		"Path to the network definition JSON file.")
	bindPFlag(rootCmd.PersistentFlags(), "ndf", rootCmd.Use)
	rootCmd.PersistentFlags().Duration("waitTimeout", 15*time.Second,
		"Duration to wait for messages to arrive.")
	bindPFlag(rootCmd.PersistentFlags(), "waitTimeout", rootCmd.Use)

	rootCmd.AddCommand(bcast)
}

// bindPFlag binds the key to a pflag.Flag. Panics on error.
func bindPFlag(flagSet *pflag.FlagSet, key, use string) {
	err := viper.BindPFlag(key, flagSet.Lookup(key))
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to bind key %q to a pflag on %s: %+v", key, use, err)
	}
}
