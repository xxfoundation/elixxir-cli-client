# CLI Client Repository

The `cli-client` repository provides a commandline interface for interacting
with the client api on the xx network. Currently, it only supports the
`broadcast` functionality.

## Building

```shell
$ git clone https://gitlab.com/elixxir/cli-client.git cli-client
$ cd cli-client
$ go mod vendor
$ go mod tidy

# Linux 64-bit binary
$ GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o cli-client.linux64 main.go

# Windows 64-bit binary
GOOS=windows GOARCH=amd64 go build -ldflags '-w -s' -o cli-client.win64 main.go

# Mac OSX 64-bit binary (Intel)
$ GOOS=darwin GOARCH=amd64 go build -ldflags '-w -s' -o cli-client.darwin64 main.go
```

## Commandline Usage

### Broadcast Channels

Using the `broadcast` subcommand, you can create or join a broadcast channel. It
demonstrates the use of broadcast channels to create a simple TUI chat client.

#### Creating a Channel

To create a broadcast channel, use the following command. `-o` sets the name of
the channel file that will be shared with others to join your channel. It should
end with the `.xxchan` prefix.

```shell
$ ./cli-client broadcast --new -o test.xxchan -n "<Channel Name>" -d "<Channel description>"
```

When creating a new channel a private RSA `.pem` file will be saved. This can be
used to send an admin message as described in the next section.

#### Joining a Channel

To join a channel, use the following command. `-o` specified the channel file to
open and `-u` specifies your chosen username.

```shell
$ ./cli-client broadcast --load -o test.xxchan -u <username> 
```

#### Sending an Admin Message

If you are the creator/admin of the channel or have the channels RSA private
key, then you can send an admin message using the following command. `-a`
specifies an admin message that will be sent and `-k` the private key location.
If no key is specified, then it searches for the default private key file
`<Channel Name>-privateKey.pem`.

```shell
$ ./cli-client broadcast --load -o test.xxchan -a "<Admin message>" -k privateKey.pem
```

#### More Help

For more help on broadcast flags, use the `-h` flag.

```shell
$ ./cli-client broadcast -h
Create or join broadcast channels.

Usage:
  cli-client broadcast {--new | --load} -o file [-n name -d description | -u username] [flags]

Flags:
  -a, --admin string         Sends the given message as an admin. Either an RSA private key PEM file exists in the default location or one must be specified with the "key" flag.
  -d, --description string   Description of the channel.
  -h, --help                 help for broadcast
  -k, --key string           Location to save/load the RSA private key PEM file. Uses the name of the channel if no path is supplied.
      --load                 Joins an existing broadcast channel.
  -n, --name string          The name of the channel.
      --new                  Creates a new broadcast channel with the specified name and description.
  -o, --open string          Location to output/open channel information file. Prints to stdout if no path is supplied.
  -u, --username string      Join the channel with this username.

Global Flags:
  -c, --config string          Path to YAML file with custom configuration..
  -v, --logLevel int           Verbosity level for log printing (2+ = Trace, 1 = Debug, 0 = Info).
  -l, --logPath string         File path to save log file to. (default "cli-client.log")
      --ndf string             Path to the network definition JSON file. By default, the prepacked NDF is used.
  -p, --password string        Password to the session file.
  -s, --session string         Sets the initial storage directory for client session data. (default "session")
      --waitTimeout duration   Duration to wait for messages to arrive. (default 15s)
```