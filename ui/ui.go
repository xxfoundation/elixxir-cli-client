////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"strings"
	"time"
)

const (
	titleBox     = "titleBox"
	infoButton   = "infoButton"
	newButton    = "newButton"
	joinButton   = "joinButton"
	leaveButton  = "leaveButton"
	chatList     = "chatList"
	channelFeed  = "channelFeed"
	messageInput = "messageInputBox"
	sendButton   = "sendButtonBox"
	messageCount = "messageCountBox"
)

const charCountFmt = "%4d/\n%4d"

func (chs *Channels) MakeUI() {
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		jww.FATAL.Panicf("Failed to make new GUI: %+v", err)
	}
	defer g.Close()

	g.Cursor = false
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen
	g.SelFrameColor = gocui.ColorGreen
	g.Highlight = true

	chs.v.switchSubView(chs.v.main.subView)
	// g.SetCurrentView(chs.v.activeArr[chs.v.active])

	g.SetManagerFunc(chs.makeLayout(chs.username, 1024))

	err = chs.initKeybindings(g)
	if err != nil {
		jww.FATAL.Panicf("Failed to generate key bindings: %+v", err)
	}

	if err = g.MainLoop(); err != nil && err != gocui.ErrQuit {
		jww.FATAL.Printf("Error in main loop: %+v", err)
		jww.FATAL.Panicf("Error in main loop: %+v", err)
	}
}

func (chs *Channels) makeLayout(username string, maxMessageLen int) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		if v, err := g.SetView(titleBox, maxX-25, 0, maxX-1, maxY-10, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf("Failed to set view %q: %+v", titleBox, err)
			}
			v.Title = " Anonymous Chat "
			v.Wrap = true
			v.Autoscroll = true

			_, err = fmt.Fprintf(v, "Controls:\n"+
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
				return errors.Errorf(
					"Failed to write to view %q: %+v", v.Name(), err)
			}

			chs.v.main.titleBox = v
		}

		if v, err := g.SetView(infoButton, maxX-25, maxY-9, maxX-1, maxY-7, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf("Failed to set view %q: %+v", infoButton, err)
			}
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("Channel Info", v))

			chs.v.main.infoButton = v
		}

		if v, err := g.SetView(newButton, 0, 0, 8, 2, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("New", v))

			chs.v.main.newButton = v
		}

		if v, err := g.SetView(joinButton, 9, 0, 17, 2, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("Join", v))

			chs.v.main.joinButton = v
		}

		if v, err := g.SetView(leaveButton, 18, 0, 26, 2, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("Leave", v))

			chs.v.main.leaveButton = v
		}

		if v, err := g.SetView(chatList, 0, 3, 26, maxY-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", chatList, err)
			}
			v.Title = " Chat List [F3] "
			v.Wrap = false
			v.Autoscroll = false

			if _, err = g.SetCurrentView(v.Name()); err != nil {
				return errors.Errorf(
					"Failed to set the current view to %q: %+v", v.Name(), err)
			}

			err := g.SetKeybinding(
				v.Name(), gocui.KeyArrowUp, gocui.ModNone, chs.moveView(-1))
			if err != nil {
				return errors.Errorf(
					"failed to bind arrow up to %s: %+v", v.Name(), err)
			}

			err = g.SetKeybinding(
				v.Name(), gocui.KeyArrowDown, gocui.ModNone, chs.moveView(1))
			if err != nil {
				return errors.Errorf(
					"failed to bind arrow down to %s: %+v", v.Name(), err)
			}

			chs.v.main.chatList = v
		}

		if v, err := g.SetView(channelFeed, 27, 0, maxX-26, maxY-7, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", channelFeed, err)
			}
			v.Title = " Channel Feed [F4] "
			v.Wrap = true
			v.Autoscroll = true

			chs.v.main.channelFeed = v
		}

		if v, err := g.SetView(messageInput, 27, maxY-6, maxX-9, maxY-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", messageInput, err)
			}
			v.Title = " Sending Message as \"" + username + "\" [F5] "
			v.Editable = true
			v.Wrap = true

			chs.v.main.messageInput = v
		}

		if v, err := g.SetView(messageCount, maxX-8, maxY-6, maxX-1, maxY-3, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", messageCount, err)
			}
			v.Frame = false
			v.Wrap = true

			_, err = fmt.Fprintf(v, charCountFmt, 0, maxMessageLen)
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}

			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				for {
					select {
					case <-ticker.C:
						g.Update(func(gui *gocui.Gui) error {
							messageInputView, err := g.View(messageInput)
							if err != nil {
								return errors.Errorf(
									"Failed to get view %q: %+v", messageInput, err)
							}
							messageCountView, err := g.View(messageCount)
							if err != nil {
								return errors.Errorf(
									"Failed to get view %q: %+v", messageCount, err)
							}

							buff := strings.TrimSpace(messageInputView.Buffer())
							n := len(buff)

							var color string
							if n >= maxMessageLen {
								messageInputView.Editable = false
								color = "\x1b[0;31m"
							} else {
								messageInputView.Editable = true
							}

							messageCountView.Clear()
							_, err = fmt.Fprintf(messageCountView, color+charCountFmt+"\x1b[0m", n, maxMessageLen)
							if err != nil {
								return errors.Errorf("Failed to write to view: %+v", err)
							}
							return nil
						})
					}
				}
			}()

			chs.v.main.messageCount = v
		}

		if v, err := g.SetView(sendButton, maxX-8, maxY-3, maxX-1, maxY-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", sendButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			_, err = v.Write([]byte(" Send "))
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}
			chs.v.main.sendButton = v
		}

		return nil
	}
}

// initKeybindings initializes all key bindings for the entire UI.
func (chs *Channels) initKeybindings(g *gocui.Gui) error {
	err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, chs.nextView)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for tab: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyEnter, gocui.ModNone, chs.readBuffs())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		sendButton, gocui.KeyEnter, gocui.ModNone, chs.readBuffs())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		sendButton, gocui.MouseLeft, gocui.ModNone, chs.readBuffs())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyCtrlJ, gocui.ModNone, addLine)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for Ctrl + J: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyBackspace, gocui.ModNone, backSpace)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyBackspace2, gocui.ModNone, backSpace)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyDelete, gocui.ModNone, backSpace)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, chs.quitWithMessage())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for Ctrl + C: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyF3, gocui.ModNone, chs.switchActiveTo(chatList))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for F3: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyF4, gocui.ModNone, chs.switchActiveTo(channelFeed))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for F4: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyF5, gocui.ModNone, chs.switchActiveTo(messageInput))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for F5: %+v", err)
	}

	err = g.SetKeybinding(
		newButton, gocui.MouseLeft, gocui.ModNone, chs.newChannel())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	err = g.SetKeybinding(
		newButton, gocui.KeyEnter, gocui.ModNone, chs.newChannel())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	err = g.SetKeybinding(
		joinButton, gocui.MouseLeft, gocui.ModNone, chs.joinChannel())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	err = g.SetKeybinding(
		joinButton, gocui.KeyEnter, gocui.ModNone, chs.joinChannel())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	err = g.SetKeybinding(
		leaveButton, gocui.MouseLeft, gocui.ModNone, chs.leaveChannel())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	err = g.SetKeybinding(
		leaveButton, gocui.KeyEnter, gocui.ModNone, chs.leaveChannel())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	err = g.SetKeybinding(
		infoButton, gocui.MouseLeft, gocui.ModNone, chs.channelInfoBox(false))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	err = g.SetKeybinding(
		infoButton, gocui.KeyEnter, gocui.ModNone, chs.channelInfoBox(false))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	for _, v := range chs.v.activeArr {
		err = chs.addScrolling(g, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (chs *Channels) switchActive() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		jww.TRACE.Printf("Set current view to %s", v.Name())
		if _, err := g.SetCurrentView(v.Name()); err != nil {
			return errors.Errorf(
				"failed to set %s as current view: %+v", v.Name(), err)
		}

		g.Cursor = chs.v.inCursorList(v.Name())

		for i, vName := range chs.v.activeArr {
			if vName == v.Name() {
				chs.v.active = i
			}
		}

		if v.Name() == chatList {
			ox, oy := v.Origin()
			_, cy := v.Cursor()

			chs.UpdateChannelFeed(cy + oy)
			v.SetOrigin(ox, oy)

			// cx, cy := v.Cursor()
			// if (cy >= 0) && (cy < len(v.ViewBufferLines())) {
			// 	chs.UpdateChannelFeed(cy)
			// } else if cy >= len(v.ViewBufferLines()) {
			// 	err := v.SetCursor(cx, len(v.ViewBufferLines()))
			// 	if err != nil {
			// 		return errors.Errorf(
			// 			"Failed to set the cursor in %q to %d, %d: %+v",
			// 			v.Name(), cx, len(v.ViewBufferLines()), err)
			// 	}
			// }
		}

		return nil
	}
}

func (chs *Channels) moveView(dy int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if v != nil && g.CurrentView() == v {
			v.Autoscroll = false
			ox, oy := v.Origin()
			_, height := v.Size()

			if (chs.currentIndex+dy < 0) || (chs.currentIndex+dy >= len(chs.channels)) {
				return nil
			}

			chs.UpdateChannelFeed(chs.currentIndex + dy)

			if chs.currentIndex-oy == height {
				if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
					if err := v.SetOrigin(ox, oy+dy); err != nil {
						return errors.Errorf("Failed to set origin of %q to %d, %d",
							v.Name(), ox, oy+dy)
					}
				}
			} else if oy == chs.currentIndex+1 {
				if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
					if err := v.SetOrigin(ox, oy+dy); err != nil {
						return errors.Errorf("Failed to set origin of %q to %d, %d",
							v.Name(), ox, oy+dy)
					}
				}
			} else if chs.currentIndex < oy {
				if err := v.SetOrigin(ox, chs.currentIndex); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, chs.currentIndex)
				}
			} else if chs.currentIndex > oy+height {
				if err := v.SetOrigin(ox, chs.currentIndex-height+1); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, chs.currentIndex)
				}
			} else {
				if err := v.SetOrigin(ox, oy); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, oy)
				}
			}

			// if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
			// 	if err := v.SetOrigin(ox, oy+dy); err != nil {
			// 		return errors.Errorf("Failed to set origin of %q to %d, %d",
			// 			v.Name(), ox, oy+dy)
			// 	}
			// }
		}
		return nil
	}
}

func (chs *Channels) switchActiveTo(name string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		jww.TRACE.Printf("Set current view to %s", name)
		if _, err := g.SetCurrentView(name); err != nil {
			return errors.Errorf(
				"failed to set %s as current view: %+v", name, err)
		}

		for i, vName := range chs.v.activeArr {
			if vName == name {
				chs.v.active = i
			}
		}

		g.Cursor = chs.v.inCursorList(v.Name())

		return nil
	}
}

func backSpace(_ *gocui.Gui, v *gocui.View) error {
	v.EditDelete(true)
	v.Editable = true

	return nil
}

func (chs *Channels) readBuffs() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.main.messageInput.Editable = false

		chs.v.main.sendButton.Highlight = true
		defer func() {
			go func() {
				chs.v.main.messageInput.Editable = true
				time.Sleep(500 * time.Millisecond)
				chs.v.main.sendButton.Highlight = false
			}()
		}()

		buffSave := chs.v.main.messageInput.Buffer()
		chs.v.main.messageInput.Clear()

		if chs.Len() == 0 {
			chs.v.main.messageInput.WriteString(buffSave)
			return nil
		}
		buff := strings.TrimSpace(buffSave)
		c := chs.GetCurrent()

		if len(buff) == 0 || len(buff) > c.maxMessageLen {
			v.WriteString(buffSave)
			jww.CRITICAL.Printf("readBuffs not sending: invalid msg")
			return nil
		}

		go func() {
			err := c.sendFn(buff)
			if err != nil {
				jww.FATAL.Panicf("Failed to send message: %+v", err)
			}
		}()
		// err := c.sendFn(buff, timestamp)
		// if err != nil {
		// 	return errors.Errorf("Failed to send message: %+v", err)
		// }

		// usernameField := "\x1b[38;5;255m" + c.myUsername + " (you) \x1b[0m"
		// timestampField := "\x1b[38;5;245m[sent " + timestamp.Format("3:04:05 pm") + "]\x1b[0m"
		// messageField := "\x1b[38;5;250m" + strings.TrimSpace(buff) + "\x1b[0m"
		// message := usernameField + " " + timestampField + "\n" + messageField

		// c.receivedMsgChan <- message

		messageCountView, err := g.View(messageCount)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", messageCount, err)
		}

		messageCountView.Clear()
		_, err = fmt.Fprintf(messageCountView, charCountFmt, 0, c.maxMessageLen)
		if err != nil {
			return errors.Errorf(
				"Failed to write to view %q: %+v", messageCount, err)
		}

		return nil
	}
}

func addLine(_ *gocui.Gui, v *gocui.View) error {
	v.EditNewLine()
	return nil
}

func (chs *Channels) scrollView(dy int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if v != nil && g.CurrentView() == v {
			v.Autoscroll = false
			ox, oy := v.Origin()
			_, height := v.Size()

			if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
				if err := v.SetOrigin(ox, oy+dy); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, oy+dy)
				}
			}
		}
		return nil
	}
}

func (chs *Channels) nextView(g *gocui.Gui, _ *gocui.View) error {
	// TODO: get current view first
	nextIndex := (chs.v.active + 1) % len(chs.v.activeArr)
	name := chs.v.activeArr[nextIndex]

	_, err := g.View(name)
	if err != nil {
		return errors.Errorf("Failed to get view %q: %+v", name, err)
	}

	if _, err = setCurrentViewOnTop(g, name); err != nil {
		return errors.Errorf("Failed to set %q to top: %+v", name, err)
	}

	g.Cursor = chs.v.inCursorList(name)

	chs.v.active = nextIndex
	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, errors.Errorf(
			"Failed to set the current view to %q: %+v", name, err)
	}
	return g.SetViewOnTop(name)
}

// TODO: close channel window and send message
func (chs *Channels) quitWithMessage() func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		return gocui.ErrQuit
	}
}

func (chs *Channels) addScrolling(g *gocui.Gui, name string) error {

	if name != chatList {
		err := g.SetKeybinding(
			name, gocui.KeyArrowUp, gocui.ModNone, chs.scrollView(-1))
		if err != nil {
			return errors.Errorf(
				"failed to bind arrow up to %s: %+v", name, err)
		}
	}

	err := g.SetKeybinding(
		name, gocui.MouseWheelUp, gocui.ModNone, chs.scrollView(-1))
	if err != nil {
		return errors.Errorf(
			"failed to bind mouse wheel up to %s: %+v", name, err)
	}

	if name != chatList {
		err = g.SetKeybinding(
			name, gocui.KeyArrowDown, gocui.ModNone, chs.scrollView(1))
		if err != nil {
			return errors.Errorf(
				"failed to bind arrow down to %s: %+v", name, err)
		}
	}

	err = g.SetKeybinding(
		name, gocui.MouseWheelDown, gocui.ModNone, chs.scrollView(1))
	if err != nil {
		return errors.Errorf(
			"failed to bind mouse wheel down to %s: %+v", name, err)
	}

	err = g.SetKeybinding(
		name, gocui.MouseLeft, gocui.ModNone, chs.switchActive())
	if err != nil {
		return errors.Errorf(
			"failed to bind mouse left click to %s: %+v", name, err)
	}

	return nil
}
