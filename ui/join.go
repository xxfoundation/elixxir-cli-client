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
	joinGroupBox          = "joinGroupBox"
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

		chs.v.switchSubView(chs.v.joinBox.subView)

		g.Cursor = true

		maxX, maxY := g.Size()
		if v, err := g.SetView(joinGroupBox, maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupBox, err)
			}
			v.Title = " Join New Channel "
			v.Wrap = true
			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
			chs.v.joinBox.joinGroupBox = v
		}

		if v, err := g.SetView(joinGroupInput, maxX/2-40+2, maxY/2-8+5, maxX/2+40-2, maxY/2-8+11, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupInput, err)
			}
			v.Title = " Channel Pretty Print "
			v.Wrap = true
			v.Editable = true
			chs.v.joinBox.joinGroupInput = v

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
				v.Name(), gocui.KeyEnter, gocui.ModNone, chs.joinGroup())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(joinGroupCancelButton, maxX/2-40+12, maxY/2-8+12, maxX/2-40+28, maxY/2-8+14, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupCancelButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
			chs.v.joinBox.joinGroupCancelButton = v

			v.WriteString(centerView("Cancel", v))

			err = g.SetKeybinding(
				joinGroupCancelButton, gocui.MouseLeft, gocui.ModNone, chs.closeJoinBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				joinGroupCancelButton, gocui.KeyEnter, gocui.ModNone, chs.closeJoinBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(joinGroupSubmitButton, maxX/2+40-28, maxY/2-8+12, maxX/2+40-12, maxY/2-8+14, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupSubmitButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
			chs.v.joinBox.joinGroupSubmitButton = v

			v.WriteString(centerView("Join", v))

			err = g.SetKeybinding(
				joinGroupSubmitButton, gocui.MouseLeft, gocui.ModNone, chs.joinGroup())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				joinGroupSubmitButton, gocui.KeyEnter, gocui.ModNone, chs.joinGroup())
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

func (chs *Channels) closeJoinBox() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.switchSubView(chs.v.main.subView)
		g.Cursor = false
		err := g.DeleteView(joinGroupBox)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupBox, err)
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

func (chs *Channels) joinGroup() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		buff := strings.TrimSpace(chs.v.joinBox.joinGroupInput.Buffer())

		if len(buff) == 0 {
			return nil
		}

		chs.v.joinBox.joinGroupSubmitButton.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				chs.v.joinBox.joinGroupSubmitButton.Highlight = false
			}()
		}()

		chanIO, err := chs.m.AddChannel(buff)
		if err != nil {
			jww.ERROR.Printf("Failed to add channel %q: %+v", buff, err)
			chs.v.joinBox.joinGroupBox.Clear()
			chs.v.joinBox.joinGroupBox.WriteString(err.Error())
			return nil
			// return errors.Errorf("Failed to add channel %q: %+v", buff, err)
		}

		// TODO: fix message size
		chs.Add(chanIO.ReceivedMsgChan, chanIO.SendFn, "", 1000, chanIO.C)

		chs.v.switchSubView(chs.v.main.subView)
		g.Cursor = false
		err = g.DeleteView(joinGroupBox)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", joinGroupBox, err)
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
