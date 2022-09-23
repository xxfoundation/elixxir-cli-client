////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"github.com/awesome-gocui/gocui"
)

func newSubView() subView {
	return subView{
		list:       []string{},
		activeArr:  []string{},
		cursorList: make(map[string]struct{}),
	}
}

// subView contains fields shared by all views.
type subView struct {
	// The index of the currently active element.
	active int

	// List of view names in the main view.
	list []string

	// List of view names that can be switch into view using <TAB>.
	activeArr []string

	// The list of views that when active, display the cursor
	cursorList map[string]struct{}
}

type mainView struct {
	subView

	// List of windows
	titleBox     *gocui.View
	infoButton   *gocui.View
	newButton    *gocui.View
	joinButton   *gocui.View
	leaveButton  *gocui.View
	chatList     *gocui.View
	channelFeed  *gocui.View
	messageInput *gocui.View
	sendButton   *gocui.View
	messageCount *gocui.View
}

type newBoxView struct {
	subView

	// List of windows
	newGroupBox          *gocui.View
	newGroupNameInput    *gocui.View
	newGroupDescInput    *gocui.View
	newGroupSubmitButton *gocui.View
	newGroupCancelButton *gocui.View
}

type joinBoxView struct {
	subView

	// List of windows
	joinGroupBox          *gocui.View
	joinGroupInput        *gocui.View
	joinGroupSubmitButton *gocui.View
	joinGroupCancelButton *gocui.View
}

type leaveBoxView struct {
	subView

	// List of windows
	leaveGroupBox          *gocui.View
	leaveGroupSubmitButton *gocui.View
	leaveGroupCancelButton *gocui.View
}

type channelInfoView struct {
	subView
	expanded bool

	// List of windows
	channelInfoBox          *gocui.View
	channelInfoBoxInside    *gocui.View
	channelInfoExpandButton *gocui.View
	channelInfoCopyButton   *gocui.View
	channelInfoCloseButton  *gocui.View
}

func (civ *channelInfoView) disableCopy() {
	for i, name := range civ.list {
		if name == channelInfoCopyButton {
			civ.list = append(civ.list[:i], civ.list[i+1:]...)
			break
		}
	}
	for i, name := range civ.activeArr {
		if name == channelInfoCopyButton {
			civ.activeArr = append(civ.activeArr[:i], civ.activeArr[i+1:]...)
			break
		}
	}
}
