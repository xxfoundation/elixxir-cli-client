////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"fmt"
	"git.xx.network/elixxir/cli-client/client"
	jww "github.com/spf13/jwalterweatherman"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Channels struct {
	m            *client.Manager
	v            *views
	channels     map[uint64]*channel
	channels2    []*channel
	currentIndex uint64
	nextIndex    uint64
	updateFeed   chan struct{}
	mux          sync.RWMutex
}

func NewChannels(m *client.Manager) *Channels {
	cs := &Channels{
		m:            m,
		v:            newViews(),
		channels:     make(map[uint64]*channel),
		currentIndex: 0,
		nextIndex:    0,
		updateFeed:   make(chan struct{}, 25),
		mux:          sync.RWMutex{},
	}

	go cs.UpdateChannelFeedThread()

	return cs
}

func (chs *Channels) Add(receivedMsgChan chan string, sendFn client.SendMessage,
	myUsername string, maxMessageLen int, c *crypto.Channel) {
	chs.mux.Lock()
	defer chs.mux.Unlock()
	chs.channels[chs.nextIndex] = &channel{
		chanBuff:        strings.Builder{},
		receivedMsgChan: receivedMsgChan,
		sendFn:          sendFn,
		myUsername:      myUsername,
		maxMessageLen:   maxMessageLen,
		unread:          true,
		c:               c,
	}

	chs.updateFeed <- struct{}{}
	go chs.UpdateChatList(chs.nextIndex)
	go chs.channels[chs.nextIndex].updateChatFeed(chs.updateFeed)
	chs.nextIndex++
}

func (chs *Channels) Remove(chanID uint64) {

	chs.mux.Lock()
	defer chs.mux.Unlock()

	// chs.channels = append(chs.channels[:chanID], chs.channels[chanID+1:]...)

	chs.nextIndex--
}

func (chs *Channels) Len() int {
	chs.mux.RLock()
	defer chs.mux.RUnlock()
	return len(chs.channels)
}

func (chs *Channels) Get(chatID uint64) *channel {
	chs.mux.RLock()
	defer chs.mux.RUnlock()
	return chs.channels[chatID]
}

func (chs *Channels) GetCurrent() *channel {
	chs.mux.RLock()
	defer chs.mux.RUnlock()
	return chs.channels[chs.currentIndex]
}

func (chs *Channels) UpdateChatList(selected uint64) {
	chs.mux.Lock()
	defer chs.mux.Unlock()

	numSp := strconv.Itoa(len(strconv.Itoa(len(chs.channels))))

	lines := make([]string, len(chs.channels))

	for cid, c := range chs.channels {
		unreadNotification := "  "
		if c.unread {
			unreadNotification = "* "
		}

		chatName := fmt.Sprintf("%"+numSp+"d. %s", cid, c.c.Name)
		line := unreadNotification + chatName
		if selected == cid {
			line = "\x1b[30;42m" + line + "\x1b[0m"

			chs.v.writeChannelInfo(chs.channels[selected].c)
		}
		lines[cid] = line
	}

	chs.v.chatList.Clear()
	_, err := fmt.Fprintf(chs.v.chatList, strings.Join(lines, "\n"))
	if err != nil {
		jww.ERROR.Print(err)
	}
}

func (chs *Channels) UpdateChannelFeed(chanID uint64) {
	chs.currentIndex = chanID
	chs.v.channelFeed.Clear()

	c := chs.channels[chanID]
	chs.v.channelFeed.Title = " Channel Feed for channel " + c.c.Name + " [F4] "
	chs.v.channelFeed.WriteString(chs.channels[chanID].getBuff())

	chs.UpdateChatList(chanID)
}

func (chs *Channels) UpdateChannelFeedThread() {
	for chs.v.channelFeed == nil {
		time.Sleep(250 * time.Millisecond)
	}

	go func() {
		cio, err := chs.m.ReplayChannels()
		if err != nil {
			jww.FATAL.Panicf("Failed to replay channels: %+v", err)
		}

		for _, c := range cio {
			// FIXME: message length
			chs.Add(c.ReceivedMsgChan, c.SendFn, "", 500, c.C)
		}
	}()

	for {
		select {
		case <-chs.updateFeed:
			chs.UpdateChannelFeed(chs.GetCurrentID())
		}
	}
}

func (chs *Channels) GetCurrentID() uint64 {
	chs.mux.RLock()
	defer chs.mux.RUnlock()
	return chs.currentIndex
}

func (chs *Channels) UpdateCurrentID(chanID uint64) {
	chs.mux.Lock()
	defer chs.mux.Unlock()
	chs.currentIndex = chanID
	chs.v.messageInput.Editable = true
}

func unmarshalChannelID(str string) uint64 {
	idString := strings.Split(strings.TrimSpace(str), "#")[1]
	idString = strings.Split(idString, "\x1B")[0]

	chanID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		jww.FATAL.Panicf("Failed to parse channel ID: %+v", err)
	}

	return chanID
}
