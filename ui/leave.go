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
	"time"
)

const (
	leaveGroupBox          = "leaveGroupBox"
	leaveGroupSubmitButton = "leaveGroupSubmitButton"
	leaveGroupCancelButton = "leaveGroupCancelButton"
)

func (chs *Channels) leaveChannel() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {

		if len(chs.channels) == 0 {
			return nil
		}

		v.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				v.Highlight = false
			}()
		}()

		chs.v.switchSubView(chs.v.leaveBox.subView)

		c := chs.channels[chs.currentIndex].c

		maxX, maxY := g.Size()
		x0, y0, x1, y1 := fixDimensions(maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, maxX, maxY)
		if v, err := g.SetView(leaveGroupBox, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveGroupBox, err)
			}
			v.Title = " Leave Channel "
			v.Wrap = true
			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
			// v.BgColor = gocui.Get256Color(236)

			v.WriteString("\n\n" +
				CenterView("Are you sure you would like to leave the following group?", v) +
				"\n\n" +
				CenterView(c.Name, v) +
				"\n\n" +
				CenterView(c.ReceptionID.String(), v))
		}

		if v, err := g.SetView(leaveGroupSubmitButton, maxX/2-40+12, y1-3, maxX/2-40+28, y1-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveGroupSubmitButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("Yes", v))

			err = g.SetKeybinding(
				leaveGroupSubmitButton, gocui.MouseLeft, gocui.ModNone, chs.leaveGroup())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				leaveGroupSubmitButton, gocui.KeyEnter, gocui.ModNone, chs.leaveGroup())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(leaveGroupCancelButton, maxX/2+40-28, y1-3, maxX/2+40-12, y1-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveGroupCancelButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("No", v))

			err = g.SetKeybinding(
				leaveGroupCancelButton, gocui.MouseLeft, gocui.ModNone, chs.closeLeaveBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				leaveGroupCancelButton, gocui.KeyEnter, gocui.ModNone, chs.closeLeaveBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if _, err := g.SetCurrentView(leaveGroupCancelButton); err != nil {
			return errors.Errorf(
				"Failed to set the current view to %q: %+v", leaveGroupCancelButton, err)
		}

		return nil
	}
}

func (chs *Channels) closeLeaveBox() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {

		chs.v.switchSubView(chs.v.main.subView)
		g.Cursor = false
		err := g.DeleteView(leaveGroupBox)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroupBox, err)
		}
		err = g.DeleteView(leaveGroupCancelButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroupCancelButton, err)
		}
		err = g.DeleteView(leaveGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroupSubmitButton, err)
		}
		return nil
	}
}

func (chs *Channels) leaveGroup() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {

		submitButton, err := g.View(leaveGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", leaveGroupSubmitButton, err)
		}

		submitButton.Highlight = true

		c := chs.channels[chs.currentIndex]
		err = chs.m.RemoveChannel(c.c.ReceptionID)
		if err != nil {
			return errors.Errorf("Failed to leave channel %s (index %d): %+v",
				c.c.ReceptionID, chs.currentIndex, err)
		}

		chs.v.switchSubView(chs.v.main.subView)

		jww.ERROR.Printf("Removing %s", c.c.ReceptionID)

		// Delete channel fro UI
		chs.Remove(chs.currentIndex)
		chs.UpdateChannelFeed(chs.currentIndex)

		err = g.DeleteView(leaveGroupBox)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroupBox, err)
		}
		err = g.DeleteView(leaveGroupCancelButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroupCancelButton, err)
		}
		err = g.DeleteView(leaveGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroupSubmitButton, err)
		}

		return nil
	}
}
