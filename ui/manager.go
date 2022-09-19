////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"git.xx.network/elixxir/cli-client/client"
	"github.com/awesome-gocui/gocui"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"sync"
)

type Manager struct {
	v                   *views
	ch                  *crypto.Channel
	receivedBroadcastCh chan client.ReceivedBroadcast
	symBroadcastFunc    client.BroadcastFn
	asymBroadcastFunc   client.BroadcastFn
	username            string
	symMaxMessageLen    int
	asymMaxMessageLen   int
	adminMode           bool
	adminModeMux        sync.RWMutex
}

func NewManager(ch *crypto.Channel,
	receivedBroadcastCh chan client.ReceivedBroadcast,
	symBroadcastFunc, asymBroadcastFunc client.BroadcastFn, username string,
	symMaxMessageLen, asymMaxMessageLen int) *Manager {
	m := &Manager{
		v:                   newViews(),
		ch:                  ch,
		receivedBroadcastCh: receivedBroadcastCh,
		symBroadcastFunc:    symBroadcastFunc,
		asymBroadcastFunc:   asymBroadcastFunc,
		username:            username,
		symMaxMessageLen:    symMaxMessageLen,
		asymMaxMessageLen:   asymMaxMessageLen,
		adminMode:           false,
	}

	return m
}

func (m *Manager) toggleAdminMode() {
	m.adminModeMux.Lock()
	defer m.adminModeMux.Unlock()
	m.adminMode = !m.adminMode
}

func (m *Manager) isAdminMode() bool {
	m.adminModeMux.RLock()
	defer m.adminModeMux.RUnlock()
	return m.adminMode
}

type views struct {
	list   []*gocui.View
	active int

	channelFeed  *gocui.View
	messageInput *gocui.View
	sendButton   *gocui.View
	messageCount *gocui.View
	adminBtn     *gocui.View
	titleBox     *gocui.View
}

func newViews() *views {
	vs := &views{
		list:   make([]*gocui.View, 0, 5),
		active: 0,
	}
	return vs
}

func (vs *views) makeList() {
	list := []*gocui.View{vs.channelFeed, vs.messageInput,
		vs.sendButton, vs.adminBtn, vs.titleBox}
	for i, v := range list {
		if v != nil {
			vs.list = append(vs.list, list[i])
		}
	}
}
