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
	newGroup             = "newGroup"
	newGroupNameInput    = "newGroupNameInput"
	newGroupDescInput    = "newGroupDescInput"
	newGroupCancelButton = "newGroupCancelButton"
	newGroupSubmitButton = "newGroupSubmitButton"
)

func (chs *Channels) newChannel() func(*gocui.Gui, *gocui.View) error {
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
			newGroupNameInput, newGroupDescInput, newGroupCancelButton, newGroupSubmitButton,
		}
		chs.v.active = 0
		chs.v.cursorList[newGroupNameInput] = struct{}{}
		chs.v.cursorList[newGroupDescInput] = struct{}{}

		g.Cursor = true

		if v, err := g.SetView(newGroup, maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroup, err)
			}
			v.Title = " Make New Channel "
			v.Wrap = true

			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
		}

		if v, err := g.SetView(newGroupNameInput, maxX/2-40+2, maxY/2-8+2, maxX/2+40-2, maxY/2-8+4, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupNameInput, err)
			}
			v.Title = " Channel Name "
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
				v.Name(), gocui.KeyEnter, gocui.ModNone, chs.newGroup(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}
		if v, err := g.SetView(newGroupDescInput, maxX/2-40+2, maxY/2-8+5, maxX/2+40-2, maxY/2-8+12, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupDescInput, err)
			}
			v.Title = " Channel Description "
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
				v.Name(), gocui.KeyEnter, gocui.ModNone, chs.newGroup(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(newGroupCancelButton, maxX/2-40+12, maxY/2-8+13, maxX/2-40+28, maxY/2-8+15, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupCancelButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			_, err = v.Write([]byte("    Cancel    "))
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}

			err = g.SetKeybinding(
				newGroupCancelButton, gocui.MouseLeft, gocui.ModNone, chs.closeNewBox(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				newGroupCancelButton, gocui.KeyEnter, gocui.ModNone, chs.closeNewBox(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(newGroupSubmitButton, maxX/2+40-28, maxY/2-8+13, maxX/2+40-12, maxY/2-8+15, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupSubmitButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			_, err = v.Write([]byte("      New      "))
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}

			err = g.SetKeybinding(
				newGroupSubmitButton, gocui.MouseLeft, gocui.ModNone, chs.newGroup(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				newGroupSubmitButton, gocui.KeyEnter, gocui.ModNone, chs.newGroup(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if _, err := g.SetCurrentView(newGroupNameInput); err != nil {
			return errors.Errorf(
				"Failed to set the current view to %q: %+v", newGroupNameInput, err)
		}

		return nil
	}
}

func (chs *Channels) closeNewBox(savedActiveArr []string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		g.Cursor = false

		chs.v.activeArr = savedActiveArr
		chs.v.active = 0
		err := g.DeleteView(newGroup)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroup, err)
		}
		err = g.DeleteView(newGroupNameInput)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupNameInput, err)
		}
		err = g.DeleteView(newGroupDescInput)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupDescInput, err)
		}
		err = g.DeleteView(newGroupCancelButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupCancelButton, err)
		}
		err = g.DeleteView(newGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupSubmitButton, err)
		}
		return nil
	}
}

func (chs *Channels) newGroup(savedActiveArr []string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		nameV, err := g.View(newGroupNameInput)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", newGroupNameInput, err)
		}
		nameBuff := strings.TrimSpace(nameV.Buffer())

		if len(nameBuff) == 0 {
			return nil
		}
		descV, err := g.View(newGroupDescInput)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", newGroupDescInput, err)
		}
		descBuff := strings.TrimSpace(descV.Buffer())

		if len(nameBuff) == 0 || len(descBuff) == 0 {
			return nil
		}

		submitButton, err := g.View(newGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", newGroupSubmitButton, err)
		}

		submitButton.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				submitButton.Highlight = false
			}()
		}()
		newGroupBox, err := g.View(newGroup)
		if err != nil {
			return errors.Errorf("Failed to get view %q: %+v", newGroup, err)
		}

		chanIO, err := chs.m.GenerateChannel(nameBuff, descBuff)
		if err != nil {
			newGroupBox.WriteString(err.Error())
			return errors.Errorf(
				"Failed to generate new channel %q: %+v", nameBuff, err)
		}

		// TODO: fix message size
		chs.Add(chanIO.ReceivedMsgChan, chanIO.SendFn, "", 1000, chanIO.C)

		chs.v.activeArr = savedActiveArr
		chs.v.active = 0
		g.Cursor = false
		err = g.DeleteView(newGroup)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroup, err)
		}
		err = g.DeleteView(newGroupNameInput)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupNameInput, err)
		}
		err = g.DeleteView(newGroupDescInput)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupDescInput, err)
		}
		err = g.DeleteView(newGroupCancelButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupCancelButton, err)
		}
		err = g.DeleteView(newGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupSubmitButton, err)
		}

		return nil
	}
}
