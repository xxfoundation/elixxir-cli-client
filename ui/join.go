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
	joinGroupSubmitButton = "joinGroupSubmitButton"
	joinGroupCancelButton = "joinGroupCancelButton"

	joinGroupErrorBox         = "joinGroupErrorBox"
	joinGroupErrorBoxMessage  = "joinGroupErrorBoxMessage"
	joinGroupErrorBackButton  = "joinGroupErrorBackButton"
	joinGroupErrorCloseButton = "joinGroupErrorCloseButton"
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
		x0, y0, x1, y1 := maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8
		if v, err := g.SetView(joinGroupBox, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", joinGroupBox, err)
			}
			v.Title = " Join New Channel "
			v.Wrap = true
			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
			chs.v.joinBox.joinGroupBox = v
		}

		// if v, err := g.SetView(joinGroupInput, maxX/2-40+2, maxY/2-8+2, maxX/2+40-2, maxY/2-8+12, 0); err != nil {
		if v, err := g.SetView(joinGroupInput, x0+2, y0+2, x1-2, y1-4, 0); err != nil {
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

		if v, err := g.SetView(joinGroupSubmitButton, maxX/2-40+12, y1-3, maxX/2-40+28, y1-1, 0); err != nil {
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

		if v, err := g.SetView(joinGroupCancelButton, maxX/2+40-28, y1-3, maxX/2+40-12, y1-1, 0); err != nil {
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

		for _, name := range chs.v.joinBox.list {
			err := g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
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
			err = chs.showJoinChannelError(g, err.Error())
			if err != nil {
				return err
			}

			return nil
		}

		// TODO: fix message size
		chs.Add(chanIO.ReceivedMsgChan, chanIO.SendFn, "", 1000, chanIO.C)

		chs.v.switchSubView(chs.v.main.subView)
		g.Cursor = false
		for _, name := range chs.v.joinBox.list {
			err = g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
		}

		return nil
	}
}

func (chs *Channels) showJoinChannelError(g *gocui.Gui, message string) error {
	chs.v.switchSubView(subView{
		active: 0,
		list: []string{joinGroupErrorBox, joinGroupErrorBoxMessage,
			joinGroupErrorBackButton, joinGroupErrorCloseButton},
		activeArr: []string{joinGroupErrorBoxMessage,
			joinGroupErrorBackButton, joinGroupErrorCloseButton},
		cursorList: nil,
	})

	g.Cursor = false

	x0, y0, x1, y1 := chs.v.joinBox.joinGroupBox.Dimensions()
	if v, err := g.SetView(joinGroupErrorBox, x0, y0, x1, y1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", joinGroupErrorBox, err)
		}
		v.Title = " Join New Channel "
		v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
	}

	if v, err := g.SetView(joinGroupErrorBoxMessage, x0+2, y0+1, x1-2, y1-4, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", joinGroupErrorBoxMessage, err)
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
	if v, err := g.SetView(joinGroupErrorBackButton, xOffset-28, y1-3, xOffset-12, y1-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", joinGroupErrorBackButton, err)
		}

		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		chs.v.joinBox.joinGroupCancelButton = v

		v.WriteString(centerView("Back", v))

		err = g.SetKeybinding(
			joinGroupErrorBackButton, gocui.MouseLeft, gocui.ModNone, chs.closeJoinErrorBox())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for left mouse button: %+v", err)
		}

		err = g.SetKeybinding(
			joinGroupErrorBackButton, gocui.KeyEnter, gocui.ModNone, chs.closeJoinErrorBox())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for enter key: %+v", err)
		}
	}

	if v, err := g.SetView(joinGroupErrorCloseButton, xOffset+12, y1-3, xOffset+28, y1-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf(
				"Failed to set view %q: %+v", joinGroupErrorCloseButton, err)
		}

		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		chs.v.joinBox.joinGroupCancelButton = v

		v.WriteString(centerView("Close", v))

		err = g.SetKeybinding(
			joinGroupErrorCloseButton, gocui.MouseLeft, gocui.ModNone, chs.closeAllJoinBoxes())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for left mouse button: %+v", err)
		}

		err = g.SetKeybinding(
			joinGroupErrorCloseButton, gocui.KeyEnter, gocui.ModNone, chs.closeAllJoinBoxes())
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for enter key: %+v", err)
		}
	}

	return nil
}

func (chs *Channels) closeJoinErrorBox() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.switchSubView(chs.v.joinBox.subView)

		for _, name := range []string{joinGroupErrorBox, joinGroupErrorBoxMessage,
			joinGroupErrorBackButton, joinGroupErrorCloseButton} {
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

func (chs *Channels) closeAllJoinBoxes() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.switchSubView(chs.v.main.subView)

		for _, name := range chs.v.joinBox.list {
			v2, _ := g.View(name)
			v2.Clear()
			err := g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
		}

		for _, name := range []string{joinGroupErrorBox, joinGroupErrorBoxMessage,
			joinGroupErrorBackButton, joinGroupErrorCloseButton} {
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
