////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"git.xx.network/elixxir/cli-client/client"
	jww "github.com/spf13/jwalterweatherman"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"strings"
	"sync"
)

type channel struct {
	chanID          uint64
	chanBuff        strings.Builder
	receivedMsgChan chan string
	sendFn          client.SendMessage
	myUsername      string
	maxMessageLen   int
	unread          bool
	mux             sync.RWMutex

	c *crypto.Channel
}

func (ch *channel) getBuff() string {
	ch.mux.Lock()
	defer ch.mux.Unlock()
	ch.unread = false
	return ch.chanBuff.String()
}

func (ch *channel) updateChatFeed(updateFeed chan struct{}, updateChatList func(uint64)) {
	for {
		select {
		case r := <-ch.receivedMsgChan:
			jww.INFO.Printf("Got broadcast for channel %d: %q", ch.chanID, r)

			ch.mux.Lock()
			ch.chanBuff.WriteString(string(r) + "\n\n")
			ch.unread = true
			ch.mux.Unlock()

			updateFeed <- struct{}{}
			go updateChatList(ch.chanID)
		}
	}
}
