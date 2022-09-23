////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package ui

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/pkg/errors"
	"gitlab.com/xx_network/primitives/id"
	"time"
)

const (
	leaveGroupBox          = "leaveGroupBox"
	leaveGroupSubmitButton = "leaveGroupSubmitButton"
	leaveGroupCancelButton = "leaveGroupCancelButton"
)

func (chs *Channels) leaveChannel() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
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
		if v, err := g.SetView(leaveGroupBox, maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveGroupBox, err)
			}
			v.Title = " Leave Channel "
			v.Wrap = true
			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
			// v.BgColor = gocui.Get256Color(236)

			v.WriteString(fmt.Sprintf("\n\n      Are you sure you would like to leave the group %s?", c.Name))
			v.WriteString(fmt.Sprintf("\n\n          ID: %s", c.ReceptionID))
		}

		if v, err := g.SetView(leaveGroupSubmitButton, maxX/2-40+12, maxY/2-8+10, maxX/2-40+28, maxY/2-8+12, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveGroupSubmitButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(centerView("Yes", v))

			err = g.SetKeybinding(
				leaveGroupSubmitButton, gocui.MouseLeft, gocui.ModNone, chs.leaveGroup(chs.currentIndex, c.ReceptionID))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				leaveGroupSubmitButton, gocui.KeyEnter, gocui.ModNone, chs.leaveGroup(chs.currentIndex, c.ReceptionID))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if v, err := g.SetView(leaveGroupCancelButton, maxX/2+40-28, maxY/2-8+10, maxX/2+40-12, maxY/2-8+12, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveGroupCancelButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(centerView("No", v))

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

func (chs *Channels) leaveGroup(chanIndex int, chanID *id.ID) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {

		submitButton, err := g.View(leaveGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", leaveGroupSubmitButton, err)
		}

		submitButton.Highlight = true

		err = chs.m.RemoveChannel(chanID)
		if err != nil {
			return errors.Errorf("Failed to leave channel %s (index %d): %+v",
				chanID, chanIndex, err)
		}

		chs.v.switchSubView(chs.v.main.subView)

		// Delete channel fro UI
		chs.Remove(chanIndex)
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
