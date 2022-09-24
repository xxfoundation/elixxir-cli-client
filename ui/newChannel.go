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
	jww "github.com/spf13/jwalterweatherman"
	"strings"
	"time"
)

const (
	newGroupBox          = "newGroupBox"
	newGroupNameInput    = "newGroupNameInput"
	newGroupDescInput    = "newGroupDescInput"
	newGroupSubmitButton = "newGroupSubmitButton"
	newGroupCancelButton = "newGroupCancelButton"

	newGroupErrorBox         = "newGroupErrorBox"
	newGroupErrorBoxMessage  = "newGroupErrorBoxMessage"
	newGroupErrorBackButton  = "newGroupErrorBackButton"
	newGroupErrorCloseButton = "newGroupErrorCloseButton"
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
		x0, y0, x1, y1 := fixDimensions(maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, maxX, maxY)
		if v, err := g.SetView(newGroupBox, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", newGroupBox, err)
			}
			v.Title = " Make New Channel "
			v.Wrap = true

			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}

			chs.v.newBox.newGroupBox = v
		}

		if v, err := g.SetView(newGroupNameInput, x0+2, y0+2, x1-2, y0+4, 0); err != nil {
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

		if v, err := g.SetView(newGroupDescInput, x0+2, y0+5, x1-2, y0+12, 0); err != nil {
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

			v.WriteString(CenterView("New", v))

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

			v.WriteString(CenterView("Cancel", v))

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

		chs.v.newBox.newGroupSubmitButton.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				chs.v.newBox.newGroupSubmitButton.Highlight = false
			}()
		}()

		var empty bool
		nameBuff := strings.TrimSpace(chs.v.newBox.newGroupNameInput.Buffer())
		if len(nameBuff) == 0 {
			chs.v.newBox.newGroupNameInput.BgColor = gocui.ColorRed
			defer func() {
				go func() {
					time.Sleep(500 * time.Millisecond)
					chs.v.newBox.newGroupNameInput.BgColor = gocui.ColorDefault
				}()
			}()

			empty = true
		}

		descBuff := strings.TrimSpace(chs.v.newBox.newGroupDescInput.Buffer())
		if len(nameBuff) == 0 || len(descBuff) == 0 {
			chs.v.newBox.newGroupDescInput.BgColor = gocui.ColorRed
			defer func() {
				go func() {
					time.Sleep(500 * time.Millisecond)
					chs.v.newBox.newGroupDescInput.BgColor = gocui.ColorDefault
				}()
			}()
			empty = true
		}

		if empty {
			return nil
		}

		chanIO, err := chs.m.GenerateChannel(nameBuff, descBuff)
		if err != nil {
			jww.ERROR.Printf("Failed to generate channel %q: %+v", nameBuff, err)
			err = chs.showNewChannelError(g, err.Error())
			if err != nil {
				return err
			}

			return nil
		}

		// TODO: fix message size
		chs.Add(chanIO.ReceivedMsgChan, chanIO.SendFn, "", 1000, chanIO.C)

		chs.v.switchSubView(chs.v.main.subView)
		g.Cursor = false
		for _, name := range chs.v.newBox.list {
			err := g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
		}

		chs.UpdateChannelFeed(len(chs.channels) - 1)

		return nil
	}
}

func (chs *Channels) showNewChannelError(g *gocui.Gui, message string) error {
	chs.v.switchSubView(subView{
		active: 0,
		list: []string{newGroupErrorBox, newGroupErrorBoxMessage,
			newGroupErrorBackButton, newGroupErrorCloseButton},
		activeArr: []string{newGroupErrorBoxMessage,
			newGroupErrorBackButton, newGroupErrorCloseButton},
		cursorList: nil,
	})

	g.Cursor = false

	x0, y0, x1, y1 := chs.v.newBox.newGroupBox.Dimensions()
	if v, err := g.SetView(newGroupErrorBox, x0, y0, x1, y1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", newGroupErrorBox, err)
		}
		v.Title = " Make New Channel "
		v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
	}

	if v, err := g.SetView(newGroupErrorBoxMessage, x0+2, y0+1, x1-2, y1-4, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", newGroupErrorBoxMessage, err)
		}

		v.Frame = false
		v.Wrap = true

		err = chs.addScrolling(g, v.Name())
		if err != nil {
			return err
		}

		v.Clear()

		v.WriteString("This error message can be found in the log\n\n")
		v.WriteString("\x1b[31m" + message + "\x1b[0m")
	}

	xOffset := (x1-x0)/2 + x0
	if v, err := g.SetView(newGroupErrorBackButton, xOffset-28, y1-3, xOffset-12, y1-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", newGroupErrorBackButton, err)
		}

		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		chs.v.newBox.newGroupCancelButton = v

		v.WriteString(CenterView("Back", v))

		err = g.SetKeybinding(
			newGroupErrorBackButton, gocui.MouseLeft, gocui.ModNone, chs.closeNewErrorBox())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for left mouse button: %+v", err)
		}

		err = g.SetKeybinding(
			newGroupErrorBackButton, gocui.KeyEnter, gocui.ModNone, chs.closeNewErrorBox())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for enter key: %+v", err)
		}
	}

	if v, err := g.SetView(newGroupErrorCloseButton, xOffset+12, y1-3, xOffset+28, y1-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", newGroupErrorCloseButton, err)
		}

		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		chs.v.newBox.newGroupCancelButton = v

		v.WriteString(CenterView("Close", v))

		err = g.SetKeybinding(
			newGroupErrorCloseButton, gocui.MouseLeft, gocui.ModNone, chs.closeAllNewBoxes())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for left mouse button: %+v", err)
		}

		err = g.SetKeybinding(
			newGroupErrorCloseButton, gocui.KeyEnter, gocui.ModNone, chs.closeAllNewBoxes())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for enter key: %+v", err)
		}
	}

	return nil
}

func (chs *Channels) closeNewErrorBox() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.switchSubView(chs.v.newBox.subView)

		for _, name := range []string{newGroupErrorBox, newGroupErrorBoxMessage,
			newGroupErrorBackButton, newGroupErrorCloseButton} {
			v2, _ := g.View(name)
			v2.Clear()
			err := g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
		}
		return nil
	}
}

func (chs *Channels) closeAllNewBoxes() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.switchSubView(chs.v.main.subView)

		for _, name := range chs.v.newBox.list {
			v2, _ := g.View(name)
			v2.Clear()
			err := g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
		}

		for _, name := range []string{newGroupErrorBox, newGroupErrorBoxMessage,
			newGroupErrorBackButton, newGroupErrorCloseButton} {
			v2, _ := g.View(name)
			v2.Clear()
			err := g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
		}
		return nil
	}
}
