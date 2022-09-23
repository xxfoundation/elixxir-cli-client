package ui

import (
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	crypto "gitlab.com/elixxir/crypto/broadcast"
)

type views struct {
	subView

	// List of windows
	main           *mainView
	newBox         *newBoxView
	joinBox        *joinBoxView
	leaveBox       *leaveBoxView
	channelInfoBox *channelInfoView
}

func newViews() *views {
	v := &views{
		subView: newSubView(),
		main: &mainView{
			subView: subView{
				active: 5,
				list: []string{titleBox, infoButton, newButton, joinButton,
					leaveButton, chatList, channelFeed, messageInput,
					sendButton, messageCount},
				activeArr: []string{titleBox, infoButton, newButton, joinButton,
					leaveButton, chatList, channelFeed, messageInput,
					sendButton},
				cursorList: map[string]struct{}{
					messageInput: {},
				},
			},
		},
		newBox: &newBoxView{
			subView: subView{
				active: 0,
				list: []string{newGroupBox, newGroupNameInput,
					newGroupDescInput, newGroupCancelButton,
					newGroupSubmitButton},
				activeArr: []string{newGroupNameInput,
					newGroupDescInput, newGroupCancelButton,
					newGroupSubmitButton},
				cursorList: map[string]struct{}{
					newGroupNameInput: {},
					newGroupDescInput: {},
				},
			},
		},
		joinBox: &joinBoxView{
			subView: subView{
				active: 0,
				list: []string{joinGroupBox, joinGroupInput,
					joinGroupCancelButton, joinGroupSubmitButton},
				activeArr: []string{joinGroupInput,
					joinGroupCancelButton, joinGroupSubmitButton},
				cursorList: map[string]struct{}{
					joinGroupInput: {},
				},
			},
		},
		leaveBox: &leaveBoxView{
			subView: subView{
				active: 1,
				list: []string{leaveGroupBox, leaveGroupSubmitButton,
					leaveGroupCancelButton},
				activeArr: []string{leaveGroupSubmitButton,
					leaveGroupCancelButton},
				cursorList: map[string]struct{}{},
			},
		},
		channelInfoBox: &channelInfoView{
			subView: subView{
				active: 0,
				list: []string{channelInfoBox, channelInfoBoxInside,
					channelInfoExpandButton, channelInfoCloseButton},
				activeArr: []string{channelInfoBoxInside, channelInfoExpandButton,
					channelInfoCloseButton},
				cursorList: map[string]struct{}{},
			},
		},
	}

	return v
}

func (v *views) inCursorList(name string) bool {
	_, exists := v.cursorList[name]
	return exists
}

func (v *views) switchSubView(sb subView) {
	v.subView.active = sb.active
	v.subView.list = sb.list
	v.subView.activeArr = sb.activeArr
	v.subView.cursorList = sb.cursorList
}

func (v *views) writeChannelInfo(c *crypto.Channel) bool {
	v.main.titleBox.Clear()
	_, err := fmt.Fprintf(v.main.titleBox,
		"\x1B[38;5;255m"+"Name:"+"\x1B[38;5;250m\n"+
			" %s\n\n"+
			"\x1B[38;5;255m"+"Description:"+"\x1B[38;5;250m\n"+
			" %s\n\n"+
			"\x1B[38;5;255m"+"Channel ID:"+"\x1B[38;5;250m\n"+
			" %s\n\n"+
			"\x1B[38;5;255m"+"Pretty Print:"+"\x1B[38;5;250m\n"+
			" %s"+"\x1b[0m"+"\n\n\n",
		c.Name, c.Description, c.ReceptionID, c.PrettyPrint())
	if err != nil {
		jww.FATAL.Panicf("%+v", err)
	}

	_, err = fmt.Fprintf(v.main.titleBox, "Controls:\n"+
		"\x1B[38;5;250m"+
		" Ctrl+C  exit\n"+
		" Tab     Switch view\n"+
		" ↑ ↓     Seek input\n"+
		" Enter   Send message\n"+
		" Ctrl+J  New line\n"+
		" F3      Chat list\n"+
		" F4      Channel feed\n"+
		" F5      Message field")
	if err != nil {
		jww.FATAL.Panicf("%+v", err)
	}

	return true
}
