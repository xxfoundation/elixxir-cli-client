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
	channels     []*channel
	currentIndex int
	nextIndex    int
	updateFeed   chan struct{}
	username     string
	mux          sync.RWMutex
}

func NewChannels(m *client.Manager) *Channels {
	cs := &Channels{
		m:            m,
		v:            newViews(),
		channels:     []*channel{},
		currentIndex: 0,
		nextIndex:    0,
		updateFeed:   make(chan struct{}, 25),
		username:     m.Username(),
	}

	go cs.UpdateChannelFeedThread()

	return cs
}

func (chs *Channels) Add(receivedMsgChan chan string, sendFn client.SendMessage,
	myUsername string, maxMessageLen int, c *crypto.Channel) {
	chs.mux.Lock()
	defer chs.mux.Unlock()
	chs.channels = append(chs.channels, &channel{
		chanBuff:        strings.Builder{},
		receivedMsgChan: receivedMsgChan,
		sendFn:          sendFn,
		myUsername:      myUsername,
		maxMessageLen:   maxMessageLen,
		unread:          false,
		c:               c,
	})

	chs.updateFeed <- struct{}{}
	go chs.UpdateChatList(chs.nextIndex)
	go chs.channels[chs.nextIndex].updateChatFeed(chs.updateFeed)
	chs.nextIndex++
}

func (chs *Channels) Remove(chanID int) {
	chs.mux.Lock()
	defer chs.mux.Unlock()

	chs.channels = append(chs.channels[:chanID], chs.channels[chanID+1:]...)

	if chanID > chs.currentIndex {
		chs.currentIndex--
	}

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

func (chs *Channels) UpdateChatList(selected int) {
	chs.mux.Lock()
	defer chs.mux.Unlock()

	numSp := strconv.Itoa(len(strconv.Itoa(len(chs.channels))))

	lines := make([]string, len(chs.channels))

	x0, _, x1, _ := chs.v.main.chatList.Dimensions()
	width := strconv.Itoa(x1 - x0 - 3)

	for cid, c := range chs.channels {
		unreadNotification := "  "
		if c.unread {
			unreadNotification = "* "
		}

		chatName := fmt.Sprintf("%"+numSp+"d. %s", cid, c.c.Name)
		chatName = fmt.Sprintf("%-"+width+"s", chatName)
		line := unreadNotification + chatName
		if selected == cid {
			line = "\x1b[30;42m" + line + "\x1b[0m"

			chs.v.writeChannelInfo(chs.channels[selected].c)
		}
		lines[cid] = line
	}

	chs.v.main.chatList.Clear()
	_, err := fmt.Fprintf(chs.v.main.chatList, strings.Join(lines, "\n"))
	if err != nil {
		jww.ERROR.Print(err)
	}
}

func (chs *Channels) UpdateChannelFeed(chanID int) {
	chs.currentIndex = chanID
	chs.v.main.channelFeed.Clear()

	c := chs.channels[chanID]
	chs.v.main.channelFeed.Title = " Channel Feed for channel " + c.c.Name + " [F4] "
	chs.v.main.channelFeed.WriteString(chs.channels[chanID].getBuff())

	chs.UpdateChatList(chanID)
}

func (chs *Channels) UpdateChannelFeedThread() {

	cio, err := chs.m.ReplayChannels()
	if err != nil {
		jww.FATAL.Panicf("Failed to replay channels: %+v", err)
	}

	for chs.v.main.channelFeed == nil {
		time.Sleep(250 * time.Millisecond)
	}

	for _, c := range cio {
		// FIXME: message length
		chs.Add(c.ReceivedMsgChan, c.SendFn, "", 500, c.C)
	}

	for {
		select {
		case <-chs.updateFeed:
			chs.UpdateChannelFeed(chs.GetCurrentID())
		}
	}
}

func (chs *Channels) GetCurrentID() int {
	chs.mux.RLock()
	defer chs.mux.RUnlock()
	return chs.currentIndex
}

func (chs *Channels) UpdateCurrentID(chanID int) {
	chs.mux.Lock()
	defer chs.mux.Unlock()
	chs.currentIndex = chanID
	chs.v.main.messageInput.Editable = true
}
