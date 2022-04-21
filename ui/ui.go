package ui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"strings"
)

const (
	channelFeed  = "channelFeed"
	messageInput = "messageInputBox"
)

var (
	viewArr = []string{channelFeed, messageInput}
	active  = 0
)

func MakeUI(payloadChan chan []byte, broadcastFn func(message []byte) error, channelName string) {
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		jww.FATAL.Panicf("Failed to make new GUI: %+v", err)
	}
	defer g.Close()

	g.Cursor = true
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen
	g.Highlight = true

	g.SetManagerFunc(makeLayout(channelName))

	err = initKeybindings(g, broadcastFn)
	if err != nil {
		jww.FATAL.Panicf("Failed to generate key bindings: %+v", err)
	}
	go func() {
		for {
			select {
			case r := <-payloadChan:
				channelFeedView, err := g.View(channelFeed)
				if err != nil {
					jww.ERROR.Print(err)
				}

				payloadParts := strings.SplitN(string(r), ": ", 2)
				username := "\x1b[38;5;255m" + string(payloadParts[0]) + "\x1b[0m"
				message := "\x1b[38;5;250m" + strings.TrimSpace(payloadParts[1]) + "\x1b[0m"

				str := username + "\n" + message + "\n\n"

				_, err = fmt.Fprintf(channelFeedView, str)
				if err != nil {
					jww.ERROR.Print(err)
					return
				}

				// _, err = channelFeedView.Write([]byte(str))
				// if err != nil {
				// 	jww.ERROR.Print(err)
				// 	return
				// }

				channelFeedView.Autoscroll = true
			}
		}
	}()

	if err = g.MainLoop(); err != nil && err != gocui.ErrQuit {
		jww.FATAL.Panicf("Error in main loop: %+v", err)
	}
}

func makeLayout(channelName string) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView(messageInput, 0, maxY-6, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Enter your message [F5] "
			v.Editable = true
			v.Wrap = true

			if _, err = g.SetCurrentView(messageInput); err != nil {
				return err
			}
		}

		if v, err := g.SetView(channelFeed, 0, 2, maxX-1, maxY-7); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Channel Feed for " + channelName + " [F4] "
			v.Wrap = true
			v.Autoscroll = true
		}
		return nil
	}
}

// initKeybindings initializes all key bindings for the entire UI.
func initKeybindings(g *gocui.Gui, broadcastFn func(message []byte) error) error {
	err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for tab: %+v", err)
	}

	err = g.SetKeybinding(
		messageInput, gocui.KeyEnter, gocui.ModNone, readBuffs(broadcastFn))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
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
	if v.Name() == channelFeed {
		g.Cursor = false
	} else {
		g.Cursor = true
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
		if name == channelFeed {
			g.Cursor = false
		} else {
			g.Cursor = true
		}
		return nil
	}
}

func readBuffs(broadcastFn func(message []byte) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		buff := v.ViewBuffer()

		if len(buff) == 0 || len(strings.TrimSpace(buff)) == 0 {
			return nil
		}

		err := broadcastFn([]byte(buff))
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

		return nil
	}
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

	if v.Name() == channelFeed {
		g.Cursor = false
	} else {
		g.Cursor = true
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

func quit(*gocui.Gui, *gocui.View) error {
	return gocui.ErrQuit
}
