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
	newGroupBox          = "newGroupBox"
	newGroupNameInput    = "newGroupNameInput"
	newGroupDescInput    = "newGroupDescInput"
	newGroupSubmitButton = "newGroupSubmitButton"
	newGroupCancelButton = "newGroupCancelButton"
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

		chs.v.switchSubView(chs.v.newBox.subView)

		g.Cursor = true

		maxX, maxY := g.Size()
		if v, err := g.SetView(newGroupBox, maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupBox, err)
			}
			v.Title = " Make New Channel "
			v.Wrap = true

			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}

			chs.v.newBox.newGroupBox = v
		}

		if v, err := g.SetView(newGroupNameInput, maxX/2-40+2, maxY/2-8+2, maxX/2+40-2, maxY/2-8+4, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupNameInput, err)
			}
			v.Title = " Channel Name "
			v.Wrap = true
			v.Editable = true

			chs.v.newBox.newGroupNameInput = v

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
				v.Name(), gocui.KeyEnter, gocui.ModNone, chs.newGroup())
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

			chs.v.newBox.newGroupDescInput = v

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
				v.Name(), gocui.KeyEnter, gocui.ModNone, chs.newGroup())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(newGroupSubmitButton, maxX/2-40+12, maxY/2-8+13, maxX/2-40+28, maxY/2-8+15, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupSubmitButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			chs.v.newBox.newGroupSubmitButton = v

			v.WriteString(centerView("New", v))

			err = g.SetKeybinding(
				newGroupSubmitButton, gocui.MouseLeft, gocui.ModNone, chs.newGroup())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				newGroupSubmitButton, gocui.KeyEnter, gocui.ModNone, chs.newGroup())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(newGroupCancelButton, maxX/2+40-28, maxY/2-8+13, maxX/2+40-12, maxY/2-8+15, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupCancelButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			chs.v.newBox.newGroupCancelButton = v

			v.WriteString(centerView("Cancel", v))

			err = g.SetKeybinding(
				newGroupCancelButton, gocui.MouseLeft, gocui.ModNone, chs.closeNewBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				newGroupCancelButton, gocui.KeyEnter, gocui.ModNone, chs.closeNewBox())
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

func (chs *Channels) closeNewBox() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		g.Cursor = false

		chs.v.switchSubView(chs.v.main.subView)
		err := g.DeleteView(newGroupBox)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupBox, err)
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

func (chs *Channels) newGroup() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		nameBuff := strings.TrimSpace(chs.v.newBox.newGroupNameInput.Buffer())
		if len(nameBuff) == 0 {
			return nil
		}

		descBuff := strings.TrimSpace(chs.v.newBox.newGroupDescInput.Buffer())
		if len(nameBuff) == 0 || len(descBuff) == 0 {
			return nil
		}

		chs.v.newBox.newGroupSubmitButton.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				chs.v.newBox.newGroupSubmitButton.Highlight = false
			}()
		}()

		chanIO, err := chs.m.GenerateChannel(nameBuff, descBuff)
		if err != nil {
			chs.v.newBox.newGroupBox.WriteString(err.Error())
			return errors.Errorf(
				"Failed to generate new channel %q: %+v", nameBuff, err)
		}

		// TODO: fix message size
		chs.Add(chanIO.ReceivedMsgChan, chanIO.SendFn, "", 1000, chanIO.C)

		chs.v.switchSubView(chs.v.main.subView)
		g.Cursor = false

		err = g.DeleteView(newGroupBox)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", newGroupBox, err)
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
