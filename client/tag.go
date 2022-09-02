////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package client

import (
	"strconv"
)

// Tag specifies the type of broadcast message so that it can be handled
// appropriately on reception.
type Tag uint8

const (
	// Default indicates a normal chat messages.
	Default Tag = 0

	// Join indicates the user has joined the channel.
	Join Tag = 1

	// Exit indicates the user has left the channel.
	Exit Tag = 2

	// Admin indicates that the sender has a private key and the message is sent
	// asymmetrically.
	Admin Tag = 3
)

// tagStringMap correlates each Tag to a human-readable name.
var tagStringMap = map[Tag]string{
	Default: "default",
	Join:    "join",
	Exit:    "exit",
	Admin:   "admin",
}

// String returns a human-readable name for the Tag for debugging purposes.
// Adheres to the fmt.Stringer interface.
func (t Tag) String() string {
	str, exists := tagStringMap[t]
	if exists {
		return str
	}

	return "INVALID TAG: " + strconv.FormatUint(uint64(t), 10)
}
