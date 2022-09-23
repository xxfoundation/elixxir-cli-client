////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package client

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/channels"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/rounds"
	"gitlab.com/elixxir/client/xxdk"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
	"strconv"
	"strings"
	"time"
)

type Manager struct {
	chanMan       channels.Manager
	chans         map[id.ID]*ChannelIO
	username      string
	maxMessageLen int
	rng           *fastRNG.StreamGenerator
}

type SendMessage func(payload string) error

type ChannelIO struct {
	ChannelID       *id.ID
	ReceivedMsgChan chan string
	SendFn          SendMessage
	C               *crypto.Channel
}

func NewChannelManager(net *xxdk.Cmix, username string) *Manager {
	rng := net.GetRng().GetStream()
	nameService, err := channels.NewDummyNameService(username, rng)
	if err != nil {
		jww.FATAL.Panicf("Failed create dummy name service: %+v", err)
	}
	rng.Close()

	m := &Manager{
		chans:         make(map[id.ID]*ChannelIO),
		username:      username,
		maxMessageLen: net.GetCmix().GetMaxMessageLength(),
		rng:           net.GetRng(),
	}

	m.chanMan = channels.NewManager(
		net.GetStorage().GetKV(), net.GetCmix(), net.GetRng(), nameService, m)

	return m
}

func (m *Manager) Username() string {
	return m.username
}

func (m *Manager) ReplayChannels() ([]*ChannelIO, error) {
	channelList := m.chanMan.GetChannels()
	chans := make([]*ChannelIO, len(channelList))
	for i := range channelList {
		chanID := channelList[i]
		c, err := m.chanMan.GetChannel(chanID)
		if err != nil {
			return nil, errors.Errorf(
				"Failed to get the channel with ID %s: %+v", chanID, err)
		}
		receiveChan := make(chan string, 125)
		sendFn := m.newSendFunc(chanID)
		cio := &ChannelIO{
			ChannelID:       chanID,
			ReceivedMsgChan: receiveChan,
			SendFn:          sendFn,
			C:               c,
		}

		m.chans[*chanID] = cio
		chans[i] = cio

		err = m.chanMan.ReplayChannel(chanID)
		if err != nil {
			return nil, errors.Errorf(
				"Failed to replay channel %s: %+v", chanID, err)
		}
	}

	return chans, nil
}

func (m *Manager) newSendFunc(chanID *id.ID) SendMessage {
	return func(payload string) error {
		// TODO: Figure out correct value for validUntil.
		msgID, rnd, ephID, err := m.chanMan.SendMessage(
			chanID, payload, 30000, cmix.GetDefaultCMIXParams())
		if err != nil {
			return errors.Errorf(
				"Failed to send message to channel %s: %+v", chanID, err)
		}

		jww.INFO.Printf("Send message %s to channel %s on round %d",
			msgID, chanID, rnd, ephID)
		return nil
	}
}

func (m *Manager) GenerateChannel(name, description string) (*ChannelIO, error) {
	rng := m.rng.GetStream()
	defer rng.Close()
	c, _, err := crypto.NewChannel(name, description, m.maxMessageLen, rng)
	if err != nil {
		return nil, errors.Errorf(
			"Failed to create new channel with name %q: %+v", name, err)
	}

	return m.AddChannel(c.PrettyPrint())
}

func (m *Manager) AddChannel(prettyPrint string) (*ChannelIO, error) {
	c, err := crypto.NewChannelFromPrettyPrint(prettyPrint)
	if err != nil {
		return nil, errors.Errorf(
			"Failed to create new channel from pretty print: %+v", err)
	}

	err = m.chanMan.JoinChannel(c)
	if err != nil {
		return nil, errors.Errorf(
			"Failed to join channel %s: %+v", c.ReceptionID, err)
	}

	m.chans[*c.ReceptionID] = &ChannelIO{
		ChannelID:       c.ReceptionID,
		ReceivedMsgChan: make(chan string, 125),
		SendFn:          m.newSendFunc(c.ReceptionID),
		C:               c,
	}

	return m.chans[*c.ReceptionID], nil
}

func (m *Manager) RemoveChannel(channelID *id.ID) error {
	return m.chanMan.LeaveChannel(channelID)
}

func (m *Manager) JoinChannel(*crypto.Channel) {}
func (m *Manager) LeaveChannel(*id.ID)         {}

func (m *Manager) ReceiveMessage(channelID *id.ID, messageID channel.MessageID, senderUsername, text string, timestamp time.Time, lease time.Duration, round rounds.Round, status channels.SentStatus) {
	c, exists := m.chans[*channelID]
	if !exists {
		jww.FATAL.Printf("No channel with ID %s exists.", channelID)
		return
	}

	jww.INFO.Printf("Received message %s for channel %s from %s at %s on round %d: %s", messageID, channelID, senderUsername, timestamp, text)

	usernameField := "\x1b[38;5;255m" + senderUsername + "\x1b[0m"
	timestampField := "\x1b[38;5;245m[received " + netTime.Now().Format("3:04:05 pm") + " / sent " + timestamp.Format("3:04:05 pm") + " / round " + strconv.Itoa(int(round.ID)) + "]\x1b[0m"
	messageField := "\x1b[38;5;250m" + strings.TrimSpace(text) + "\x1b[0m"
	message := usernameField + " " + timestampField + "\n" + messageField

	c.ReceivedMsgChan <- message
}

func (m *Manager) ReceiveReply(channelID *id.ID, messageID channel.MessageID, reactionTo channel.MessageID, senderUsername, text string, timestamp time.Time, lease time.Duration, round rounds.Round, status channels.SentStatus) {
	jww.INFO.Printf("Received reply: channelID:%s messageID:%s reactionTo:%s senderUsername:%q text:%q timestamp:%s lease:%s round:%d round:%s status:%d", channelID, messageID, reactionTo, senderUsername, text, timestamp, lease, round.ID, status)
}

func (m *Manager) ReceiveReaction(channelID *id.ID, messageID channel.MessageID, reactionTo channel.MessageID, senderUsername, reaction string, timestamp time.Time, lease time.Duration, round rounds.Round, status channels.SentStatus) {
	jww.INFO.Printf("Received reaction: channelID:%s messageID:%s reactionTo:%s senderUsername:%q reaction:%q timestamp:%s lease:%s round:%d round:%s status:%d", channelID, messageID, reactionTo, senderUsername, reaction, timestamp, lease, round.ID, status)
}

func (m *Manager) UpdateSentStatus(messageID channel.MessageID, status channels.SentStatus) {
	jww.INFO.Printf("UpdateSentStatus: messageID:%s status:%d", messageID, status)
}
