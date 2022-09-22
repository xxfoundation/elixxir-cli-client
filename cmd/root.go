////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.com/xx_network/primitives/utils"
	"log"
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
		// Initiate config file
		initConfig(viper.GetString("config"))

		// Initialize logging and print version
		initLog(viper.GetString("logPath"), viper.GetInt("logLevel"))
		jww.INFO.Printf(Version())
	},
}

// parsePassword parses the client password string.
func parsePassword(pwStr string) []byte {
	if strings.HasPrefix(pwStr, "0x") {
		return getPwFromHexString(pwStr[2:])
	} else if strings.HasPrefix(pwStr, "b64:") {
		return getPwFromB64String(pwStr[4:])
	} else {
		return []byte(pwStr)
	}
}

// getPwFromHexString decodes the hex-encoded password.
func getPwFromHexString(pwStr string) []byte {
	pwBytes, err := hex.DecodeString(
		fmt.Sprintf("%0*d%s", 66-len(pwStr), 0, pwStr))
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to get password %q from hex string: %+v", pwStr, err)
	}
	return pwBytes
}

// getPwFromB64String decodes the base 64 encoded password.
func getPwFromB64String(pwStr string) []byte {
	pwBytes, err := base64.StdEncoding.DecodeString(pwStr)
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to get password %q from base 64 string: %+v", pwStr, err)
	}
	return pwBytes
}

// initConfig reads in config file and ENV variables if set.
func initConfig(configPath string) {
	jww.INFO.Printf("Getting config file %s", configPath)
	// Use default config location if none is passed
	var err error
	if configPath == "" {
		configPath, err = utils.SearchDefaultLocations(
			"cli-client.yaml", "xxnetwork")
		if err != nil {
			jww.DEBUG.Printf("Failed to find config file: %+v", err)
		}
	} else {
		configPath, err = utils.ExpandPath(configPath)
		if err != nil {
			jww.DEBUG.Printf("Failed to expand config file path: %+v", err)
		}
	}

	jww.INFO.Printf("Setting config file %s", configPath)
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv() // Read in environment variables that match

	// If a config file is found, read it in
	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			jww.FATAL.Panicf(
				"Unable to read config file (%s): %+v", configPath, err)
		}
	}
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
			jww.SetStdoutOutput(logFile)
			jww.INFO.Printf("Setting log output to %q", logPath)
		}
	} else {
		jww.INFO.Printf("No log output set: no log path provided")
	}

	// Select the level of logs to display
	var threshold jww.Threshold
	if logLevel > 1 {
		// Turn on trace logs
		threshold = jww.LevelTrace
		jww.SetFlags(log.LstdFlags | log.Lmicroseconds)
	} else if logLevel == 1 {
		// Turn on debugging logs
		threshold = jww.LevelDebug
		jww.SetFlags(log.LstdFlags | log.Lmicroseconds)
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
	rootCmd.PersistentFlags().StringP("logPath", "l", "cli-client.log",
		"File path to save log file to.")
	bindPFlag(rootCmd.PersistentFlags(), "logPath", rootCmd.Use)

	rootCmd.PersistentFlags().IntP("logLevel", "v", 0,
		"Verbosity level for log printing (2+ = Trace, 1 = Debug, 0 = Info).")
	bindPFlag(rootCmd.PersistentFlags(), "logLevel", rootCmd.Use)

	rootCmd.PersistentFlags().StringP("config", "c", "",
		"Path to YAML file with custom configuration..")
	bindPFlag(rootCmd.PersistentFlags(), "config", rootCmd.Use)

	rootCmd.PersistentFlags().StringP("session", "s", "session",
		"Sets the initial storage directory for client session data.")
	bindPFlag(rootCmd.PersistentFlags(), "session", rootCmd.Use)

	rootCmd.PersistentFlags().StringP("password", "p", "",
		"Password to the session file.")
	bindPFlag(rootCmd.PersistentFlags(), "password", rootCmd.Use)

	rootCmd.PersistentFlags().String("ndf", "",
		"Path to the network definition JSON file. By default, the "+
			"prepacked NDF is used.")
	bindPFlag(rootCmd.PersistentFlags(), "ndf", rootCmd.Use)

	rootCmd.PersistentFlags().Duration("waitTimeout", 15*time.Second,
		"Duration to wait for messages to arrive.")
	bindPFlag(rootCmd.PersistentFlags(), "waitTimeout", rootCmd.Use)

	rootCmd.AddCommand(bCast)
}

// bindPFlag binds the key to a pflag.Flag. Panics on error.
func bindPFlag(flagSet *pflag.FlagSet, key, use string) {
	err := viper.BindPFlag(key, flagSet.Lookup(key))
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to bind key %q to a pflag on %s: %+v", key, use, err)
	}
}

// bindPFlag binds the key to a pflag.Flag. Panics on error.
func markFlagRequired(cmd *cobra.Command, name string) {
	err := cmd.MarkFlagRequired(name)
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to mark flag %q on command %s as required: %+v",
			name, cmd.Name(), err)
	}
}

// hidePFlag hides the key to a pflag.Flag in the help text. Panics on error.
func hidePFlag(flagSet *pflag.FlagSet, key, use string) {
	err := flagSet.MarkHidden(key)
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to hide key %q to a pflag on %s: %+v", key, use, err)
	}
}

// printUsageError prints the provided error with the commands' usage/help
// message and exits.
func printUsageError(cmd *cobra.Command, err error) {
	cmd.PrintErrln("Error:", err.Error())
	cmd.Println(cmd.UsageString())
	fmt.Println(err)
	os.Exit(1)
}
