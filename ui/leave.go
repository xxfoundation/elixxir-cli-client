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
	crypto "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/xx_network/primitives/id"
	"time"
)

const (
	leaveGroup             = "leaveGroup"
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

		_, cy := chs.v.chatList.Cursor()
		var c *crypto.Channel
		if (cy >= 0) && (cy < len(v.ViewBufferLines())) {
			c = chs.channels[uint64(cy)].c
		} else {
			return nil
		}

		maxX, maxY := g.Size()
		savedActiveArr := make([]string, len(chs.v.activeArr))
		copy(savedActiveArr, chs.v.activeArr)
		copy(savedActiveArr, chs.v.activeArr)

		chs.v.activeArr = []string{
			leaveGroupCancelButton, leaveGroupSubmitButton,
		}
		chs.v.active = 0

		if v, err := g.SetView(leaveGroup, maxX/2-40, maxY/2-8, maxX/2+40, maxY/2+8, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", leaveGroup, err)
			}
			v.Title = " Leave Channel "
			v.Wrap = true
			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
			v.BgColor = gocui.NewRGBColor(41, 40, 52)

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

			_, err = v.Write([]byte("      Yes      "))
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}

			err = g.SetKeybinding(
				leaveGroupSubmitButton, gocui.MouseLeft, gocui.ModNone, chs.leaveGroup(c.ReceptionID, savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				leaveGroupSubmitButton, gocui.KeyEnter, gocui.ModNone, chs.leaveGroup(c.ReceptionID, savedActiveArr))
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

			_, err = v.Write([]byte("      No       "))
			if err != nil {
				return errors.Errorf(
					"Failed to write to %q: %+v", v.Name(), err)
			}

			err = g.SetKeybinding(
				leaveGroupCancelButton, gocui.MouseLeft, gocui.ModNone, chs.closeLeaveBox(savedActiveArr))
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				leaveGroupCancelButton, gocui.KeyEnter, gocui.ModNone, chs.closeLeaveBox(savedActiveArr))
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

func (chs *Channels) closeLeaveBox(savedActiveArr []string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		chs.v.activeArr = savedActiveArr
		chs.v.active = 0
		g.Cursor = false
		err := g.DeleteView(leaveGroup)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroup, err)
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

func (chs *Channels) leaveGroup(chanID *id.ID, savedActiveArr []string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {

		submitButton, err := g.View(leaveGroupSubmitButton)
		if err != nil {
			return errors.Errorf(
				"Failed to get view %q: %+v", leaveGroupSubmitButton, err)
		}

		submitButton.Highlight = true

		chs.m.LeaveChannel(chanID)

		// TODO: delete channel from UI

		chs.v.activeArr = savedActiveArr
		chs.v.active = 0
		g.Cursor = false
		err = g.DeleteView(leaveGroup)
		if err != nil {
			return errors.Errorf(
				"Failed to delete view %q: %+v", leaveGroup, err)
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
