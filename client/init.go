package client

import (
	_ "embed"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/api"
	"io/fs"
	"io/ioutil"
	"os"
	"time"
)

//go:embed ndf.json
var ndfJSON []byte

// InitClient initializes and returns a new api.Client. If a session folder
// already exists, then the client is loaded instead.
func InitClient(password []byte, storeDir, ndfPath string) (*api.Client, error) {
	// Create a new client if none exist
	if _, err := os.Stat(storeDir); errors.Is(err, fs.ErrNotExist) {
		// Load NDF
		if ndfPath != "" {
			ndfJSON, err = ioutil.ReadFile(ndfPath)
			if err != nil {
				return nil, errors.Errorf("failed to read NDF file: %+v", err)
			}
		}

		err = api.NewClient(string(ndfJSON), storeDir, password, "")
		if err != nil {
			return nil, errors.Errorf("failed to create new client: %+v", err)
		}
	}

	// Load the client
	client, err := api.Login(storeDir, password, nil, api.GetDefaultParams())
	if err != nil {
		return nil, errors.Errorf("failed to log in into client: %+v", err)
	}

	return client, nil
}

// ConnectToNetwork connects the client to the network.
func ConnectToNetwork(client *api.Client, timeout time.Duration) error {
	// Start the network follower
	err := client.StartNetworkFollower(5 * time.Second)
	if err != nil {
		return errors.Errorf("failed to start the network follower: %+v", err)
	}

	// Wait until connected or crash on timeout
	connected := make(chan bool, 10)
	client.GetNetworkInterface().AddHealthCallback(
		func(isConnected bool) { connected <- isConnected })
	waitUntilConnected(connected, timeout)

	// After connection, wait until registered with at least 85% of nodes
	for numReg, total := 1, 100; numReg < (total*3)/4; {
		time.Sleep(1 * time.Second)

		numReg, total, err = client.GetNodeRegistrationStatus()
		if err != nil {
			return errors.Errorf(
				"failed to get node registration status: %+v", err)
		}

		jww.INFO.Printf("Registering with nodes (%d/%d)...", numReg, total)
	}

	return nil
}

// waitUntilConnected waits until the network is connected.
func waitUntilConnected(connected chan bool, timeout time.Duration) {
	timeoutTimer := time.NewTimer(timeout)
	// Wait until connected or panic after time out is reached
	for isConnected := false; !isConnected; {
		select {
		case isConnected = <-connected:
			jww.INFO.Printf("Network status: %t", isConnected)
		case <-timeoutTimer.C:
			jww.FATAL.Panicf("Timed out after %s while waiting for network "+
				"connection.", timeout)
		}
	}
}
