////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package client

import (
	"git.xx.network/elixxir/cli-client/username"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/channels"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/rounds"
	"gitlab.com/elixxir/client/storage/versioned"
	"gitlab.com/elixxir/client/xxdk"
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fastRNG"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
	"os"
	"strconv"
	"strings"
	"time"
)

type Manager struct {
	chanMan       channels.Manager
	chans         map[id.ID]*ChannelIO
	username      string
	nickname      string
	maxMessageLen int
	rng           *fastRNG.StreamGenerator
	counter       uint64
}

type SendMessage func(payload string) error

type ChannelIO struct {
	ChannelID       *id.ID
	ReceivedMsgChan chan string
	SendFn          SendMessage
	C               *crypto.Channel
}

const identityKey = "identityKey"

func NewChannelManager(net *xxdk.Cmix, nickname string) *Manager {
	m := &Manager{
		chans:         make(map[id.ID]*ChannelIO),
		nickname:      nickname,
		maxMessageLen: net.GetCmix().GetMaxMessageLength(),
		rng:           net.GetRng(),
		counter:       0,
	}

	f := func(path string) (channels.EventModel, error) {
		return m, nil
	}

	obj, err := net.GetStorage().GetKV().Get(identityKey, 0)
	if err == nil {
		m.chanMan, err = channels.LoadManager(string(obj.Data),
			net.GetStorage().GetKV(), net.GetCmix(), net.GetRng(), f)
		if err != nil {
			jww.FATAL.Panicf("Failed create new manager: %+v", err)
		}

		m.username = m.chanMan.GetIdentity().Codename

		return m
	} else if err != nil && net.GetStorage().GetKV().Exists(err) {
		jww.FATAL.Panicf("Failed to store storage tag: %+v", err)
	}
	var usernames []string
	usernamesMap := make(map[string]channel.PrivateIdentity)

	rng := net.GetRng().GetStream()
	for i := 0; i < 100; i++ {
		pi, err := channel.GenerateIdentity(rng)
		if err != nil {
			jww.FATAL.Panicf("Failed to generate identity: %+v", err)
		}

		usernamesMap[pi.Codename] = pi
		usernames = append(usernames, pi.Codename)
	}
	rng.Close()

	selected := make(chan string)
	quit := make(chan struct{})

	go username.MakeUI(usernames, selected, quit)

	var chosenUsername string
	select {
	case chosenUsername = <-selected:
	case <-quit:
		os.Exit(0)
	}

	chosenIdentity := usernamesMap[chosenUsername]

	m.username = chosenUsername

	m.chanMan, err = channels.NewManager(chosenIdentity,
		net.GetStorage().GetKV(), net.GetCmix(), net.GetRng(), f)
	if err != nil {
		jww.FATAL.Panicf("Failed create new manager: %+v", err)
	}

	err = net.GetStorage().GetKV().Set(identityKey, &versioned.Object{
		Version:   0,
		Timestamp: netTime.Now(),
		Data:      []byte(m.chanMan.GetStorageTag()),
	})
	if err != nil {
		jww.FATAL.Panicf("Failed to store storage tag: %+v", err)
	}

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

	err = m.chanMan.SetNickname(m.nickname, c.ReceptionID)
	if err != nil {
		return nil, err
	}

	return m.chans[*c.ReceptionID], nil
}

func (m *Manager) RemoveChannel(channelID *id.ID) error {
	return m.chanMan.LeaveChannel(channelID)
}

func (m *Manager) JoinChannel(*crypto.Channel) {}
func (m *Manager) LeaveChannel(*id.ID)         {}

func (m *Manager) ReceiveMessage(channelID *id.ID, messageID channel.MessageID,
	nickname, text string, identity channel.Identity, timestamp time.Time,
	lease time.Duration, round rounds.Round, status channels.SentStatus) uint64 {
	c, exists := m.chans[*channelID]
	if !exists {
		jww.FATAL.Printf("No channel with ID %s exists.", channelID)
	}

	jww.INFO.Printf("Received message %s for channel %s from %s at %s on round %d: %s", messageID, channelID, nickname, timestamp, text)

	usernameField := "\x1b[38;1;255m" + nickname + "\x1b[0m\x1b[38;5;251m#" + identity.Codename + "\x1b[0m"
	timestampField := "\x1b[38;5;242m[" + timestamp.Format("3:04:05 pm 1/2/06") + " | round " + strconv.Itoa(int(round.ID)) + "]\x1b[0m"
	messageField := "\x1b[38;5;250m" + strings.TrimSpace(text) + "\x1b[0m"
	message := usernameField + " " + timestampField + "\n" + messageField

	c.ReceivedMsgChan <- message

	m.counter++
	return m.counter
}

func (m *Manager) ReceiveReply(channelID *id.ID, messageID channel.MessageID,
	reactionTo channel.MessageID, nickname, text string,
	identity channel.Identity, timestamp time.Time,
	lease time.Duration, round rounds.Round, status channels.SentStatus) uint64 {
	jww.INFO.Printf("Received reply: channelID:%s messageID:%s reactionTo:%s "+
		"nickname:%q text:%q timestamp:%s lease:%s round:%d "+
		"round:%s status:%d", channelID, messageID, reactionTo, nickname, text,
		timestamp, lease, round.ID, status)
	m.counter++
	return m.counter
}

func (m *Manager) ReceiveReaction(channelID *id.ID, messageID channel.MessageID,
	reactionTo channel.MessageID, nickname, reaction string,
	identity channel.Identity, timestamp time.Time, lease time.Duration,
	round rounds.Round, status channels.SentStatus) uint64 {
	jww.INFO.Printf("Received reaction: channelID:%s messageID:%s reactionTo:%s "+
		"nickname:%q reaction:%q timestamp:%s lease:%s round:%d round:%s "+
		"status:%d", channelID, messageID, reactionTo, nickname, reaction,
		timestamp, lease, round.ID, status)
	m.counter++
	return m.counter
}

func (m *Manager) UpdateSentStatus(uuid uint64, messageID channel.MessageID,
	timestamp time.Time, round rounds.Round, status channels.SentStatus) {
	jww.INFO.Printf("UpdateSentStatus: messageID:%s status:%d", messageID, status)
}
