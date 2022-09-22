package ui

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	jww "github.com/spf13/jwalterweatherman"
	crypto "gitlab.com/elixxir/crypto/broadcast"
)

type views struct {
	activeArr    []string
	active       int
	cursorList   map[string]struct{}
	titleBox     *gocui.View
	newButton    *gocui.View
	joinButton   *gocui.View
	leaveButton  *gocui.View
	chatList     *gocui.View
	channelFeed  *gocui.View
	messageInput *gocui.View
	sendButton   *gocui.View
	messageCount *gocui.View
}

func (v *views) inCursorList(name string) bool {
	_, exists := v.cursorList[name]
	return exists
}

func (v *views) writeChannelInfo(c *crypto.Channel) bool {
	v.titleBox.Clear()
	_, err := fmt.Fprintf(v.titleBox,
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

	_, err = fmt.Fprintf(v.titleBox, "Controls:\n"+
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
