////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"github.com/awesome-gocui/gocui"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	joinGroup             = "joinGroup"
	joinGroupInput        = "joinGroupInput"
	joinGroupCancelButton = "joinGroupCancelButton"
	joinGroupSubmitButton = "joinGroupSubmitButton"
)

func (chs *Channels) joinChannel() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				v.Highlight = false
			}()
		}()

		maxX, maxY := g.Size()
		savedActiveArr := make([]string, len(chs.v.activeArr))
		copy(savedActiveArr, chs.v.activeArr)
		copy(savedActiveArr, chs.v.activeArr)

		chs.v.activeArr = []string{
			joinGroupInput, joinGroupCancelButton, joinGroupSubmitButton,
		}
		chs.v.active = 0
		chs.v.cursorList[joinGroupInput] = struct{}{}

		g.Cursor = true

		if v, err := g.SetView(joinGroup, maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroup, err)
			}
			v.Title = " Join New Channel "
			v.Wrap = true
			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
		}

		if v, err := g.SetView(joinGroupInput, maxX/2-40+2, maxY/2-8+2, maxX/2+40-2, maxY/2-8+8, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupInput, err)
			}
			v.Title = " Channel Pretty Print "
			v.Wrap = true
			v.Editable = true

			err = g.SetKeybinding(v.Name(), gocui.KeyArrowUp, gocui.ModNone, chs.scrollView(-1))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for arrow up: %+v", err)
			}

			err = g.SetKeybinding(v.Name(), gocui.MouseWheelUp, gocui.ModNone, chs.scrollView(-1))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for wheel up: %+v", err)
			}

			err = g.SetKeybinding(v.Name(), gocui.KeyArrowDown, gocui.ModNone, chs.scrollView(1))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for arrow down: %+v", err)
			}

			err = g.SetKeybinding(v.Name(), gocui.MouseWheelDown, gocui.ModNone, chs.scrollView(1))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for wheel down: %+v", err)
			}

			err = g.SetKeybinding(v.Name(), gocui.MouseLeft, gocui.ModNone, chs.switchActive())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				v.Name(), gocui.KeyEnter, gocui.ModNone, chs.joinGroup(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(joinGroupCancelButton, maxX/2-40+12, maxY/2-8+10, maxX/2-40+28, maxY/2-8+12, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupCancelButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			_, err = v.Write([]byte("    Cancel    "))
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}

			err = g.SetKeybinding(
				joinGroupCancelButton, gocui.MouseLeft, gocui.ModNone, chs.closeJoinBox(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				joinGroupCancelButton, gocui.KeyEnter, gocui.ModNone, chs.closeJoinBox(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(joinGroupSubmitButton, maxX/2+40-28, maxY/2-8+10, maxX/2+40-12, maxY/2-8+12, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupSubmitButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			_, err = v.Write([]byte("      Join      "))
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}

			err = g.SetKeybinding(
				joinGroupSubmitButton, gocui.MouseLeft, gocui.ModNone, chs.joinGroup(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				joinGroupSubmitButton, gocui.KeyEnter, gocui.ModNone, chs.joinGroup(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if _, err := g.SetCurrentView(joinGroupInput); err != nil {
			return errors.Errorf(
				"Failed to set the current view to %q: %+v", joinGroupInput, err)
		}

		return nil
	}
}

func (chs *Channels) closeJoinBox(savedActiveArr []string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.activeArr = savedActiveArr
		chs.v.active = 0
		g.Cursor = false
		err := g.DeleteView(joinGroup)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroup, err)
		}
		err = g.DeleteView(joinGroupInput)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupInput, err)
		}
		err = g.DeleteView(joinGroupCancelButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupCancelButton, err)
		}
		err = g.DeleteView(joinGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupSubmitButton, err)
		}
		return nil
	}
}

func (chs *Channels) joinGroup(savedActiveArr []string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v, err := g.View(joinGroupInput)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", joinGroupInput, err)

		}
		buff := strings.TrimSpace(v.Buffer())

		if len(buff) == 0 {
			return nil
		}

		submitButton, err := g.View(joinGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", joinGroupSubmitButton, err)
		}

		submitButton.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				submitButton.Highlight = false
			}()
		}()

		chanIO, err := chs.m.AddChannel(buff)
		if err != nil {
			return errors.Errorf("Failed to add channel %q: %+v", buff, err)
		}

		// TODO: fix message size
		chs.Add(chanIO.ReceivedMsgChan, chanIO.SendFn, "", 1000, chanIO.C)

		chs.v.activeArr = savedActiveArr
		chs.v.active = 0
		g.Cursor = false
		err = g.DeleteView(joinGroup)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroup, err)
		}
		err = g.DeleteView(joinGroupInput)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupInput, err)
		}
		err = g.DeleteView(joinGroupCancelButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupCancelButton, err)
		}
		err = g.DeleteView(joinGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupSubmitButton, err)
		}

		return nil
	}
}
