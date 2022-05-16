package ui

import (
	"fmt"
	"git.xx.network/elixxir/cli-client/client"
	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	crypto "gitlab.com/elixxir/crypto/broadcast"
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
)

const charCountFmt = "%4d/\n%4d"

var (
	viewArr = []string{channelFeed, messageInput, sendButton, titleBox}
	active  = 0
)

func MakeUI(payloadChan chan client.ReceivedBroadcast,
	broadcastFn client.BroadcastFn, c *crypto.Channel, username string,
	maxMessageLen int) {
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		jww.FATAL.Panicf("Failed to make new GUI: %+v", err)
	}
	defer g.Close()

	g.Cursor = true
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen
	g.Highlight = true

	g.SetManagerFunc(makeLayout(c, username, maxMessageLen))

	err = initKeybindings(g, broadcastFn, maxMessageLen)
	if err != nil {
		jww.FATAL.Panicf("Failed to generate key bindings: %+v", err)
	}
	go func() {
		for {
			select {
			case r := <-payloadChan:
				jww.INFO.Printf("Got broadcast: %+v", r)
				channelFeedView, err := g.View(channelFeed)
				if err != nil {
					jww.ERROR.Printf("Failed to get view %q: %+v", channelFeed, err)
					continue
				}

				var message string

				switch r.Tag {
				case client.Default:
					usernameField := "\x1b[38;5;255m" + r.Username + "\x1b[0m"
					timestampField := "\x1b[38;5;245m[sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"
					messageField := "\x1b[38;5;250m" + strings.TrimSpace(string(r.Message)) + "\x1b[0m"

					message = usernameField + " " + timestampField + "\n" + messageField
				case client.Join:
					usernameField := "\x1b[38;5;255m" + r.Username + "\x1b[0m \x1B[38;5;250mhas joined the channel.\x1B[0m"
					timestampField := "\x1b[38;5;245m[sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"

					message = usernameField + " " + timestampField
				case client.Exit:
					usernameField := "\x1b[38;5;255m" + r.Username + "\x1b[0m \x1B[38;5;250mhas left the channel.\x1B[0m"
					timestampField := "\x1b[38;5;245m[sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"

					message = usernameField + " " + timestampField
				case client.Admin:
					usernameField := "\x1b[38;5;160m" + r.Username + " [ADMIN]\x1b[0m"
					timestampField := "\x1b[38;5;245m[sent " + r.Timestamp.Format("3:04:05 pm") + " / received " + netTime.Now().Format("3:04:05 pm") + "]\x1b[0m"
					messageField := "\x1b[38;5;124m" + strings.TrimSpace(string(r.Message)) + "\x1b[0m"

					message = usernameField + " " + timestampField + "\n" + messageField
				}

				message = message + "\n\n"

				_, err = fmt.Fprintf(channelFeedView, message)
				if err != nil {
					jww.ERROR.Print(err)
					continue
				}

				channelFeedView.Autoscroll = true
			}
		}
	}()

	err = broadcastFn(client.Join, netTime.Now(), []byte{})
	if err != nil {
		jww.FATAL.Panicf("Failed to send initial join message: %+v", err)
	}

	if err = g.MainLoop(); err != nil && err != gocui.ErrQuit {
		jww.FATAL.Panicf("Error in main loop: %+v", err)
	}
}

func makeLayout(c *crypto.Channel, username string, maxMessageLen int) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		if v, err := g.SetView(titleBox, maxX-25, 0, maxX-1, maxY-7); err != nil {
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
				" Ctrl+J  New Line\n"+
				" F4      Channel feed\n"+
				" F5      Message field\n\n"+
				"\x1b[0m"+
				"Channel Info:\n"+
				"\x1b[38;5;252mName:\n\x1b[38;5;248m"+c.Name+"\x1b[0m\n\n"+
				"\x1b[38;5;252mDescription:\n\x1b[38;5;248m"+c.Description+"\x1b[0m\n\n"+
				"\x1b[38;5;252mID:\n\x1b[38;5;248m"+c.ReceptionID.String()+"\x1b[0m")
			if err != nil {
				return err
			}
		}

		if v, err := g.SetView(channelFeed, 0, 0, maxX-26, maxY-7); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Channel Feed for \"" + c.Name + "\" [F4] "
			v.Wrap = true
			v.Autoscroll = true
		}

		if v, err := g.SetView(messageInput, 0, maxY-6, maxX-9, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Sending Message as \"" + username + "\" [F5] "
			v.Editable = true
			v.Wrap = true

			if _, err = g.SetCurrentView(messageInput); err != nil {
				return err
			}
		}

		if v, err := g.SetView(messageCount, maxX-8, maxY-6, maxX-1, maxY-3); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = false
			v.Wrap = true

			_, err = fmt.Fprintf(v, charCountFmt, 0, maxMessageLen)
			if err != nil {
				return err
			}

			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				for {
					select {
					case <-ticker.C:
						g.Update(func(gui *gocui.Gui) error {
							messageInputView, err := g.View(messageInput)
							if err != nil {
								return errors.Errorf("Failed to get view: %+v", err)
							}
							messageCountView, err := g.View(messageCount)
							if err != nil {
								return errors.Errorf("Failed to get view: %+v", err)
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
		}

		if v, err := g.SetView(sendButton, maxX-8, maxY-3, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			v.Highlight = true

			_, err = v.Write([]byte("\n Send "))
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// initKeybindings initializes all key bindings for the entire UI.
func initKeybindings(g *gocui.Gui, broadcastFn client.BroadcastFn, maxMessageLen int) error {
	err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for tab: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyEnter, gocui.ModNone, readBuffs(broadcastFn, maxMessageLen))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		sendButton, gocui.KeyEnter, gocui.ModNone, readBuffs(broadcastFn, maxMessageLen))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		sendButton, gocui.MouseLeft, gocui.ModNone, readBuffs(broadcastFn, maxMessageLen))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyCtrlJ, gocui.ModNone, addLine)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
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

	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitWithMessage(broadcastFn))
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

	return nil
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

func readBuffs(broadcastFn client.BroadcastFn, maxMessageLen int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		v, err := g.View(messageInput)
		if err != nil {
			return err
		}
		buff := strings.TrimSpace(v.Buffer())

		if len(buff) == 0 || len(buff) > maxMessageLen {
			return nil
		}

		err = broadcastFn(client.Default, netTime.Now(), []byte(buff))
		if err != nil {
			return err
		}

		v.Clear()
		err = v.SetOrigin(0, 0)
		if err != nil {
			return errors.Errorf("Failed to set origin back to (0, 0): %+v", err)
		}
		err = v.SetCursor(0, 0)
		if err != nil {
			return errors.Errorf("Failed to set cursor back to (0, 0): %+v", err)
		}

		messageCountView, err := g.View(messageCount)
		if err != nil {
			return errors.Errorf("Failed to get view: %+v", err)
		}

		messageCountView.Clear()
		_, err = fmt.Fprintf(messageCountView, charCountFmt, 0, maxMessageLen)
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

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	name := viewArr[nextIndex]

	_, err := g.View(name)
	if err != nil {
		return err
	}

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}

	if v.Name() == messageInput {
		g.Cursor = true
	} else {
		g.Cursor = false
	}

	active = nextIndex
	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func quitWithMessage(broadcastFn client.BroadcastFn) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		err := broadcastFn(client.Exit, netTime.Now(), []byte{})
		if err != nil {
			jww.ERROR.Printf("Failed to send exit message: %+v", err)
		}
		return gocui.ErrQuit
	}
}
