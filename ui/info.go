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
	"golang.design/x/clipboard"
	"strconv"
	"strings"
	"time"
)

const (
	channelInfoBox          = "channelInfoBox"
	channelInfoBoxInside    = "channelInfoBoxInside"
	channelInfoExpandButton = "channelInfoExpandButton"
	channelInfoCopyButton   = "channelInfoCopyButton"
	channelInfoCloseButton  = "channelInfoCloseButton"
)

func (chs *Channels) channelInfoBox(expand bool) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		if len(chs.channels) == 0 {
			return nil
		}

		chs.v.main.infoButton.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				chs.v.main.infoButton.Highlight = false
			}()
		}()

		chs.v.channelInfoBox.expanded = expand
		chs.v.switchSubView(chs.v.channelInfoBox.subView)

		c := chs.channels[chs.currentIndex].c

		maxX, maxY := g.Size()

		var xo0, yo0, xo1, yo1 int
		if expand {
			xo0, yo0, xo1, yo1 = -1, 0, maxX, maxY-1
		} else {
			xo0, yo0, xo1, yo1 = fixDimensions(maxX/2-70, maxY/2-15, maxX/2+70, maxY/2+15, maxX, maxY)
		}
		if v, err := g.SetView(channelInfoBox, xo0, yo0, xo1, yo1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", channelInfoBox, err)
			}
			v.Title = " Channel Information "
			v.Wrap = true
			v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬'}
			chs.v.channelInfoBox.channelInfoBox = v
		}

		var x0, y0, x1, y1 int
		if expand {
			x0, y0, x1, y1 = -1, 0, maxX, maxY-5
		} else {
			x0, y0, x1, y1 = fixDimensions(xo0+2, yo0+1, xo1-2, yo1-5, maxX, maxY)
		}
		if v, err := g.SetView(channelInfoBoxInside, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", channelInfoBoxInside, err)
			}
			v.Wrap = true
			v.Frame = false

			chs.v.channelInfoBox.channelInfoBoxInside = v

			lines := []string{
				fmt.Sprintf("%s:\n%s", Bold("Name"), Dim(c.Name)),
				fmt.Sprintf("%s:\n%s", Bold("Description"), Dim(c.Description)),
				fmt.Sprintf("%s:\n%s", Bold("ReceptionID"), Dim(c.ReceptionID.String())),
				fmt.Sprintf("%s:\n%s", Bold("Pretty Print"), Dim(c.PrettyPrint())),
				fmt.Sprintf("%s:\n%s", Bold("Salt"), Dim(fmt.Sprintf("%v", c.Salt))),
				fmt.Sprintf("%s:\n%s", Bold("RsaPubKeyHash"), Dim(fmt.Sprintf("%v", c.RsaPubKeyHash))),
				fmt.Sprintf("%s:\n%s", Bold("RsaPubKeyLength"), Dim(strconv.Itoa(c.RsaPubKeyLength))),
				fmt.Sprintf("%s:\n%s", Bold("RSASubPayloads"), Dim(strconv.Itoa(c.RSASubPayloads))),
				fmt.Sprintf("%s:\n%s", Bold("Secret"), Dim(fmt.Sprintf("%v", c.Secret))),
			}

			v.WriteString(strings.Join(lines, "\n\n"))

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
		}

		if v, err := g.SetView(channelInfoExpandButton, maxX/2-40+12, yo1-3, maxX/2-40+28, yo1-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", channelInfoExpandButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
			chs.v.channelInfoBox.channelInfoExpandButton = v

			err = g.SetKeybinding(
				channelInfoExpandButton, gocui.MouseLeft, gocui.ModNone, chs.resizeInfoBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				channelInfoExpandButton, gocui.KeyEnter, gocui.ModNone, chs.resizeInfoBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if expand {
			chs.v.channelInfoBox.channelInfoExpandButton.Clear()
			chs.v.channelInfoBox.channelInfoExpandButton.WriteString(CenterView("Contract", chs.v.channelInfoBox.channelInfoExpandButton))
		} else {
			chs.v.channelInfoBox.channelInfoExpandButton.Clear()
			chs.v.channelInfoBox.channelInfoExpandButton.WriteString(CenterView("Expand", chs.v.channelInfoBox.channelInfoExpandButton))
		}

		if chs.useClipboard {
			if v, err := g.SetView(channelInfoCopyButton, maxX/2-8, yo1-3, maxX/2+8, yo1-1, 0); err != nil {
				if err != gocui.ErrUnknownView {
					return errors.Errorf(
						"Failed to set view %q: %+v", channelInfoCopyButton, err)
				}

				v.SelBgColor = gocui.ColorGreen
				v.SelFgColor = gocui.ColorBlack
				chs.v.channelInfoBox.channelInfoCopyButton = v

				v.WriteString(CenterView("Copy", v))

				err = g.SetKeybinding(
					channelInfoCopyButton, gocui.MouseLeft, gocui.ModNone, chs.copyPrettyPrint())
				if err != nil {
					return errors.Errorf(
						"failed to set key binding for left mouse button: %+v", err)
				}

				err = g.SetKeybinding(
					channelInfoCopyButton, gocui.KeyEnter, gocui.ModNone, chs.copyPrettyPrint())
				if err != nil {
					return errors.Errorf(
						"failed to set key binding for enter key: %+v", err)
				}
			}
		}

		if v, err := g.SetView(channelInfoCloseButton, maxX/2+40-28, yo1-3, maxX/2+40-12, yo1-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf(
					"Failed to set view %q: %+v", channelInfoCloseButton, err)
			}

			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
			chs.v.channelInfoBox.channelInfoCloseButton = v

			v.WriteString(CenterView("Close", v))

			err = g.SetKeybinding(
				channelInfoCloseButton, gocui.MouseLeft, gocui.ModNone, chs.closeInfoBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for left mouse button: %+v", err)
			}

			err = g.SetKeybinding(
				channelInfoCloseButton, gocui.KeyEnter, gocui.ModNone, chs.closeInfoBox())
			if err != nil {
				return errors.Errorf(
					"failed to set key binding for enter key: %+v", err)
			}
		}

		if _, err := g.SetCurrentView(channelInfoBoxInside); err != nil {
			return errors.Errorf(
				"Failed to set the current view to %q: %+v", channelInfoBoxInside, err)
		}

		return nil
	}
}

func (chs *Channels) closeInfoBox() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {

		chs.v.switchSubView(chs.v.main.subView)
		g.Cursor = false

		for _, name := range chs.v.channelInfoBox.list {
			err := g.DeleteView(name)
			if err != nil {
				return errors.Errorf(
					"Failed to delete view %q: %+v", name, err)
			}
		}
		return nil
	}
}

func (chs *Channels) resizeInfoBox() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				v.Highlight = false
			}()
		}()

		return chs.channelInfoBox(!chs.v.channelInfoBox.expanded)(g, v)
	}
}

func (chs *Channels) copyPrettyPrint() func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				v.Highlight = false
			}()
		}()

		c := chs.channels[chs.currentIndex].c
		clipboard.Write(clipboard.FmtText, []byte(c.PrettyPrint()))

		return nil
	}
}

func fixDimensions(x0, y0, x1, y1, maxX, maxY int) (int, int, int, int) {
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > maxX-1 {
		x1 = maxX - 1
	}
	if y1 > maxY-1 {
		y1 = maxY - 1
	}
	return x0, y0, x1, y1
}
