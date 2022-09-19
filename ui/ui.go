////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"fmt"
	"git.xx.network/elixxir/cli-client/client"
	"github.com/awesome-gocui/gocui"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/xx_network/primitives/netTime"
	"strings"
	"time"
)

const (
	titleBox     = "titleBox"
	channelFeed  = "channelFeed"
	messageInput = "messageInputBox"
	sendButton   = "sendButtonBox"
	messageCount = "messageCountBox"
	adminBtn     = "adminButtonView"
)

const charCountFmt = "%4d/\n%4d"

var (
	viewArr = []string{channelFeed, messageInput, sendButton, titleBox}
)

func (m *Manager) MakeUI() {
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		jww.FATAL.Panicf("Failed to make new GUI: %+v", err)
	}
	defer g.Close()

	g.Cursor = true
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen
	g.SelFrameColor = gocui.ColorGreen
	g.Highlight = true

	g.SetManagerFunc(m.makeLayout())

	err = m.initKeybindings(g)
	if err != nil {
		jww.FATAL.Panicf("Failed to generate key bindings: %+v", err)
	}
	go func() {
		for {
			select {
			case r := <-m.receivedBroadcastCh:

				for m.v.channelFeed == nil {
					time.Sleep(250 * time.Millisecond)
				}

				jww.INFO.Printf("Got broadcast: %+v", r)

				var message string
				tsFmt := "\u001B[38;5;242m["
				unFmt := "\u001B[38;5;255m"

				switch r.Tag {
				case client.Default:
					usernameField := unFmt + r.Username + "\x1b[0m"
					timestampField := tsFmt + "sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"
					messageField := "\x1b[38;5;250m" + strings.TrimSpace(string(r.Message)) + "\x1b[0m"

					message = usernameField + " " + timestampField + "\n" + messageField
				case client.Join:
					usernameField := unFmt + r.Username + "\x1b[0m \x1B[38;5;250mhas joined the channel.\x1B[0m"
					timestampField := tsFmt + "sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"

					message = usernameField + " " + timestampField
				case client.Exit:
					usernameField := unFmt + r.Username + "\x1b[0m \x1B[38;5;250mhas left the channel.\x1B[0m"
					timestampField := tsFmt + "sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"

					message = usernameField + " " + timestampField
				case client.Admin:
					usernameField := "\x1b[41m[ADMIN]\x1b[0m"
					timestampField := tsFmt + "sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"
					messageField := "\x1b[31m" + strings.TrimSpace(string(r.Message)) + "\x1b[0m"

					message = usernameField + " " + timestampField + "\n" + messageField
				}

				message = message + "\n\n"

				_, err = fmt.Fprintf(m.v.channelFeed, message)
				if err != nil {
					jww.ERROR.Print(err)
					continue
				}

				m.v.channelFeed.Autoscroll = true
			}
		}
	}()

	err = m.symBroadcastFunc(client.Join, netTime.Now(), []byte{})
	if err != nil {
		jww.FATAL.Panicf("Failed to send initial join message: %+v", err)
	}

	if err = g.MainLoop(); err != nil && err != gocui.ErrQuit {
		jww.FATAL.Panicf("Error in main loop: %+v", err)
	}
}

func (m *Manager) makeLayout() func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		deltaY := 7
		if m.asymBroadcastFunc != nil {
			deltaY = 10
		}

		adminControl := "\n"
		if m.asymBroadcastFunc != nil {
			adminControl = " F6      Admin toggle\n\n"
		}

		if v, err := g.SetView(titleBox, maxX-25, 0, maxX-1, maxY-deltaY, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " xx Channel Chat "
			v.Wrap = true
			v.Autoscroll = true

			_, err = fmt.Fprintf(v, "Controls:\n"+
				"\u001B[38;5;250m"+
				" Ctrl+C  exit\n"+
				" Tab     Switch view\n"+
				" ↑ ↓     Seek input\n"+
				" Enter   Send message\n"+
				" Ctrl+J  New line\n"+
				" F4      Channel feed\n"+
				" F5      Message field\n"+
				adminControl+
				"\x1b[0m"+
				"Channel Info:\n"+
				"\x1b[38;5;252mName:\n\x1b[38;5;248m"+m.ch.Name+"\x1b[0m\n\n"+
				"\x1b[38;5;252mDescription:\n\x1b[38;5;248m"+m.ch.Description+"\x1b[0m\n\n"+
				"\x1b[38;5;252mID:\n\x1b[38;5;248m"+m.ch.ReceptionID.String()+"\x1b[0m")
			if err != nil {
				return err
			}
			m.v.titleBox = v
		}

		if v, err := g.SetView(channelFeed, 0, 0, maxX-26, maxY-7, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Channel Feed for \"" + m.ch.Name + "\" [F4] "
			v.Wrap = true
			v.Autoscroll = true
			m.v.channelFeed = v
		}

		if v, err := g.SetView(messageInput, 0, maxY-6, maxX-9, maxY-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Sending Message as \"" + m.username + "\" [F5] "
			v.Editable = true
			v.Wrap = true

			if _, err = g.SetCurrentView(messageInput); err != nil {
				return err
			}
			m.v.messageInput = v
		}

		if v, err := g.SetView(messageCount, maxX-8, maxY-6, maxX-1, maxY-3, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = false
			v.Wrap = true

			_, err = fmt.Fprintf(v, charCountFmt, 0, m.symMaxMessageLen)
			if err != nil {
				return err
			}

			go func() {
				ticker := time.NewTicker(50 * time.Millisecond)
				for {
					select {
					case <-ticker.C:
						g.Update(func(gui *gocui.Gui) error {
							buff := strings.TrimSpace(m.v.messageInput.Buffer())
							n := len(buff)

							var color string
							if (!m.isAdminMode() && n >= m.symMaxMessageLen) ||
								(m.isAdminMode() && n >= m.asymMaxMessageLen) {
								m.v.messageInput.Editable = false
								color = "\x1b[0;31m"
							} else {
								m.v.messageInput.Editable = true
							}

							max := m.symMaxMessageLen
							if m.isAdminMode() {
								max = m.asymMaxMessageLen
							}

							m.v.messageCount.Clear()
							_, err = fmt.Fprintf(m.v.messageCount, color+charCountFmt+"\x1b[0m", n, max)
							if err != nil {
								return errors.Errorf("Failed to write to view: %+v", err)
							}
							return nil
						})
					}
				}
			}()
			m.v.messageCount = v
		}

		if v, err := g.SetView(sendButton, maxX-8, maxY-3, maxX-1, maxY-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			v.Highlight = false
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			_, err = v.Write([]byte(" Send "))
			if err != nil {
				return err
			}
			m.v.sendButton = v
		}

		if m.asymBroadcastFunc != nil {
			if v, err := g.SetView(adminBtn, maxX-25, maxY-9, maxX-1, maxY-7, 0); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Title = " [F6] "
				v.Highlight = false
				v.SelBgColor = gocui.ColorRed
				v.SelFgColor = gocui.ColorBlack

				_, err = fmt.Fprintf(v, "    ☐ Send as Admin    ")
				if err != nil {
					return err
				}
				m.v.adminBtn = v
			}
		}

		m.v.makeList()

		return nil
	}
}

// initKeybindings initializes all key bindings for the entire UI.
func (m *Manager) initKeybindings(g *gocui.Gui) error {
	err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, m.nextView)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for tab: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyEnter, gocui.ModNone, m.readBuffs())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		sendButton, gocui.KeyEnter, gocui.ModNone, m.readBuffs())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		sendButton, gocui.MouseLeft, gocui.ModNone, m.readBuffs())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyCtrlJ, gocui.ModNone, addLine)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for Ctrl+J: %+v", err)
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

	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, m.quitWithMessage())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for Ctrl + C: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyF4, gocui.ModNone, switchActiveTo(channelFeed))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for F4: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyF5, gocui.ModNone, switchActiveTo(messageInput))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for F5: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyF6, gocui.ModNone, switchActiveTo(adminBtn))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for F6: %+v", err)
	}

	for _, v := range viewArr {
		err = g.SetKeybinding(v, gocui.KeyArrowUp, gocui.ModNone, scrollView(-1))
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for arrow up: %+v", err)
		}

		err = g.SetKeybinding(v, gocui.MouseWheelUp, gocui.ModNone, scrollView(-1))
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for wheel up: %+v", err)
		}

		err = g.SetKeybinding(v, gocui.KeyArrowDown, gocui.ModNone, scrollView(1))
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for arrow down: %+v", err)
		}

		err = g.SetKeybinding(v, gocui.MouseWheelDown, gocui.ModNone, scrollView(1))
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for wheel down: %+v", err)
		}

		err = g.SetKeybinding(v, gocui.MouseLeft, gocui.ModNone, switchActive)
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for left mouse button: %+v", err)
		}
	}

	err = g.SetKeybinding(adminBtn, gocui.MouseLeft, gocui.ModNone, m.toggleAdmin())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	err = g.SetKeybinding(adminBtn, gocui.KeyEnter, gocui.ModNone, m.toggleAdmin())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	err = g.SetKeybinding(adminBtn, gocui.KeySpace, gocui.ModNone, m.toggleAdmin())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	return nil
}

func (m *Manager) toggleAdmin() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		m.toggleAdminMode()
		v.Highlight = !v.Highlight
		if m.isAdminMode() {
			m.v.messageInput.Title = " Sending Message as \"ADMIN\" [F5] "
			m.v.messageInput.FgColor = gocui.ColorRed
			m.v.messageInput.TitleColor = gocui.ColorRed

			v.Clear()
			_, err := fmt.Fprintf(v, "    ☑ Send as Admin    ")
			if err != nil {
				return err
			}
		} else {
			m.v.messageInput.Title = " Sending Message as \"" + m.username + "\" [F5] "
			m.v.messageInput.FgColor = gocui.ColorDefault
			m.v.messageInput.TitleColor = gocui.ColorDefault

			v.Clear()
			_, err := fmt.Fprintf(v, "    ☐ Send as Admin    ")
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func switchActive(g *gocui.Gui, v *gocui.View) error {
	jww.TRACE.Printf("Set current view to %s", v.Name())
	if _, err := g.SetCurrentView(v.Name()); err != nil {
		return errors.Errorf(
			"failed to set %s as current view: %+v", v.Name(), err)
	}
	if v.Name() == messageInput {
		g.Cursor = true
	} else {
		g.Cursor = false
	}
	return nil
}

func switchActiveTo(name string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		jww.TRACE.Printf("Set current view to %s", name)
		if _, err := g.SetCurrentView(name); err != nil {
			return errors.Errorf(
				"failed to set %s as current view: %+v", name, err)
		}
		if name == messageInput {
			g.Cursor = true
		} else {
			g.Cursor = false
		}
		return nil
	}
}

func backSpace(_ *gocui.Gui, v *gocui.View) error {
	v.EditDelete(true)
	v.Editable = true

	return nil
}

func (m *Manager) readBuffs() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {

		buff := strings.TrimSpace(m.v.messageInput.Buffer())

		if len(buff) == 0 ||
			(!m.isAdminMode() && len(buff) > m.symMaxMessageLen) ||
			(m.isAdminMode() && len(buff) > m.asymMaxMessageLen) {
			return nil
		}

		m.v.sendButton.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				m.v.sendButton.Highlight = false
			}()
		}()

		var err error
		if m.isAdminMode() {
			err = m.asymBroadcastFunc(client.Admin, netTime.Now(), []byte(buff))
		} else {
			err = m.symBroadcastFunc(client.Default, netTime.Now(), []byte(buff))
		}

		if err != nil {
			return err
		}

		m.v.messageInput.Clear()
		err = m.v.messageInput.SetOrigin(0, 0)
		if err != nil {
			return errors.Errorf("Failed to set origin back to (0, 0): %+v", err)
		}
		err = m.v.messageInput.SetCursor(0, 0)
		if err != nil {
			return errors.Errorf("Failed to set cursor back to (0, 0): %+v", err)
		}

		m.v.messageCount.Clear()
		_, err = fmt.Fprintf(m.v.messageCount, charCountFmt, 0, m.symMaxMessageLen)
		if err != nil {
			return errors.Errorf("Failed to write to view: %+v", err)
		}

		return nil
	}
}

func addLine(_ *gocui.Gui, v *gocui.View) error {
	v.EditNewLine()
	return nil
}

func scrollView(dy int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if v != nil && g.CurrentView() == v {
			v.Autoscroll = false
			ox, oy := v.Origin()
			_, height := v.Size()

			if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
				if err := v.SetOrigin(ox, oy+dy); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (m *Manager) nextView(g *gocui.Gui, _ *gocui.View) error {
	nextIndex := (m.v.active + 1) % len(m.v.list)
	v := m.v.list[nextIndex]

	if _, err := g.SetCurrentView(v.Name()); err != nil {
		return err
	}

	if v.Name() == messageInput {
		g.Cursor = true
	} else {
		g.Cursor = false
	}

	m.v.active = nextIndex
	return nil
}

func (m *Manager) quitWithMessage() func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		err := m.symBroadcastFunc(client.Exit, netTime.Now(), []byte{})
		if err != nil {
			jww.ERROR.Printf("Failed to send exit message: %+v", err)
		}
		return gocui.ErrQuit
	}
}
