////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package username

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"strconv"
	"strings"
	"time"
)

const (
	containerBox = "containerBox"
	usernameList = "usernameList"
	selectButton = "selectButton"
	cancelButton = "cancelButton"
)

var activeList = []string{usernameList, selectButton, cancelButton}
var active = 0
var usernameArr []string
var selected = 0

func MakeUI(usernames []string, selected chan string, quit chan struct{}) {
	usernameArr = usernames
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		jww.FATAL.Panicf("Failed to make new GUI: %+v", err)
	}
	defer g.Close()

	g.Cursor = false
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen
	g.SelFrameColor = gocui.ColorGreen
	g.Highlight = true

	g.SetManagerFunc(makeLayout(usernames))

	err = initKeybindings(g, selected, quit)
	if err != nil {
		jww.FATAL.Panicf("Failed to generate key bindings: %+v", err)
	}

	if err = g.MainLoop(); err != nil && err != gocui.ErrQuit {
		jww.FATAL.Printf("Error in main loop: %+v", err)
		jww.FATAL.Panicf("Error in main loop: %+v", err)
	}
}

func makeLayout(usernames []string) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		x0, y0, x1, y1 := 0, 0, maxX-1, maxY-1
		if v, err := g.SetView(containerBox, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf("Failed to set view %q: %+v", containerBox, err)
			}
			v.Title = " Username Selection "
		}

		x0, y0, x1, y1 = x0+2, y0+1, x1-2, y1-5
		if v, err := g.SetView(usernameList, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf("Failed to set view %q: %+v", containerBox, err)
			}
			v.Frame = false
			v.Wrap = false

			if _, err = g.SetCurrentView(v.Name()); err != nil {
				return errors.Errorf(
					"Failed to set the current view to %q: %+v", v.Name(), err)
			}
			v.WriteString(printUsernames(v))
		}

		x0, y0, x1, y1 = maxX/2-28, maxY-4, maxX/2-12, maxY-2
		if v, err := g.SetView(selectButton, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf("Failed to set view %q: %+v", selectButton, err)
			}
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("Select Username", v))
		}

		x0, y0, x1, y1 = maxX/2+12, maxY-4, maxX/2+28, maxY-2
		if v, err := g.SetView(cancelButton, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return errors.Errorf("Failed to set view %q: %+v", cancelButton, err)
			}
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			v.WriteString(CenterView("Cancel", v))
		}

		return nil
	}
}

// initKeybindings initializes all key bindings for the entire UI.
func initKeybindings(g *gocui.Gui, selected chan string, quit chan struct{}) error {
	err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView)
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for tab: %+v", err)
	}

	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitWithMessage(quit))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for Ctrl + C: %+v", err)
	}

	err = g.SetKeybinding(
		cancelButton, gocui.MouseLeft, gocui.ModNone, quitWithMessage(quit))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	err = g.SetKeybinding(
		cancelButton, gocui.KeyEnter, gocui.ModNone, quitWithMessage(quit))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	err = g.SetKeybinding(
		selectButton, gocui.MouseLeft, gocui.ModNone, selectUsername(selected))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	err = g.SetKeybinding(
		selectButton, gocui.KeyEnter, gocui.ModNone, selectUsername(selected))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for enter key: %+v", err)
	}

	err = g.SetKeybinding(usernameList, gocui.KeyArrowUp, gocui.ModNone, moveView(-1))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for arrow up: %+v", err)
	}

	err = g.SetKeybinding(usernameList, gocui.MouseWheelUp, gocui.ModNone, scrollView(-1))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for wheel up: %+v", err)
	}

	err = g.SetKeybinding(usernameList, gocui.KeyArrowDown, gocui.ModNone, moveView(1))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for arrow down: %+v", err)
	}

	err = g.SetKeybinding(usernameList, gocui.MouseWheelDown, gocui.ModNone, scrollView(1))
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for wheel down: %+v", err)
	}

	err = g.SetKeybinding(usernameList, gocui.MouseLeft, gocui.ModNone, switchActive())
	if err != nil {
		return errors.Errorf(
			"failed to set key binding for left mouse button: %+v", err)
	}

	return nil
}

func switchActive() func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		jww.TRACE.Printf("Set current view to %s", v.Name())
		if _, err := g.SetCurrentView(v.Name()); err != nil {
			return errors.Errorf(
				"failed to set %s as current view: %+v", v.Name(), err)
		}

		for i, vName := range activeList {
			if vName == v.Name() {
				active = i
			}
		}

		if v.Name() == usernameList {
			ox, oy := v.Origin()
			_, cy := v.Cursor()

			selected = cy + oy
			v.Clear()
			v.WriteString(printUsernames(v))
			v.SetOrigin(ox, oy)
		}

		return nil
	}
}

func printUsernames(v *gocui.View) string {
	lines := make([]string, len(usernameArr))
	x0, _, x1, _ := v.Dimensions()
	width := strconv.Itoa(x1 - x0 - 3)
	for i, un := range usernameArr {
		line := fmt.Sprintf("%2d. %s", i, un)
		line = fmt.Sprintf("%-"+width+"s", line)
		if selected == i {
			line = "\x1b[30;42m" + line + "\x1b[0m"
		}
		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

func moveView(dy int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if v != nil && g.CurrentView() == v {
			v.Autoscroll = false
			ox, oy := v.Origin()
			_, height := v.Size()

			if (selected+dy < 0) || (selected+dy >= len(usernameArr)) {
				return nil
			}

			selected += dy
			v.Clear()
			v.WriteString(printUsernames(v))

			if selected-oy == height {
				if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
					if err := v.SetOrigin(ox, oy+dy); err != nil {
						return errors.Errorf("Failed to set origin of %q to %d, %d",
							v.Name(), ox, oy+dy)
					}
				}
			} else if oy == selected+1 {
				if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
					if err := v.SetOrigin(ox, oy+dy); err != nil {
						return errors.Errorf("Failed to set origin of %q to %d, %d",
							v.Name(), ox, oy+dy)
					}
				}
			} else if selected < oy {
				if err := v.SetOrigin(ox, selected); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, selected)
				}
			} else if selected > oy+height {
				if err := v.SetOrigin(ox, selected-height+1); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, selected)
				}
			} else {
				if err := v.SetOrigin(ox, oy); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, oy)
				}
			}

			// if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
			// 	if err := v.SetOrigin(ox, oy+dy); err != nil {
			// 		return errors.Errorf("Failed to set origin of %q to %d, %d",
			// 			v.Name(), ox, oy+dy)
			// 	}
			// }
		}
		return nil
	}
}

func scrollView(dy int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if v != nil && g.CurrentView() == v {
			v.Autoscroll = false
			ox, oy := v.Origin()
			_, height := v.Size()

			if (dy+oy >= 0) && (dy+oy+height <= len(v.ViewBufferLines())) {
				if err := v.SetOrigin(ox, oy+dy); err != nil {
					return errors.Errorf("Failed to set origin of %q to %d, %d",
						v.Name(), ox, oy+dy)
				}
			}
		}
		return nil
	}
}

func nextView(g *gocui.Gui, _ *gocui.View) error {
	// TODO: get current view first
	nextIndex := (active + 1) % len(activeList)
	name := activeList[nextIndex]

	_, err := g.View(name)
	if err != nil {
		return errors.Errorf("Failed to get view %q: %+v", name, err)
	}

	if _, err = setCurrentViewOnTop(g, name); err != nil {
		return errors.Errorf("Failed to set %q to top: %+v", name, err)
	}

	active = nextIndex
	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, errors.Errorf(
			"Failed to set the current view to %q: %+v", name, err)
	}
	return g.SetViewOnTop(name)
}

func selectUsername(selectedChan chan string) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				v.Highlight = false
			}()
		}()

		err := confirmBox(g, selectedChan)
		if err != nil {
			return err
		}

		return nil
	}
}

// TODO: close channel window and send message
func quitWithMessage(quit chan struct{}) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Highlight = true
		defer func() {
			go func() {
				time.Sleep(500 * time.Millisecond)
				v.Highlight = false
			}()
		}()
		go func() {
			time.Sleep(500 * time.Millisecond)
			quit <- struct{}{}
		}()
		return gocui.ErrQuit
	}
}

// Center returns the text centered, using spaces, in the given width.
func Center(s string, w int) string {
	return fmt.Sprintf("%[1]*s", -w, fmt.Sprintf("%[1]*s", (w+len(s))/2, s))
}

// CenterView returns the text centered in the view.
func CenterView(s string, v *gocui.View) string {
	x0, _, x1, _ := v.Dimensions()
	w := x1 - x0 - 1
	return Center(s, w)
}

const (
	confirmDialogue = "confirmDialogue"
	yesButton       = "yesButton"
	NoButton        = "NoButton"
)

func confirmBox(g *gocui.Gui, selectedChan chan string) error {
	activeList = []string{yesButton, NoButton}

	maxX, maxY := g.Size()
	x0, y0, x1, y1 := fixDimensions(maxX/2-40, maxY/2-10, maxX/2+40, maxY/2+10, maxX, maxY)
	if v, err := g.SetView(confirmDialogue, x0, y0, x1, y1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf("Failed to set view %q: %+v", confirmDialogue, err)
		}

		v.WriteString("\n\n")
		v.WriteString(CenterView("Are you sure you want the following username?", v))
		v.WriteString("\n\n")
		v.WriteString(CenterView("\""+usernameArr[selected]+"\"", v))
	}

	if v, err := g.SetView(yesButton, maxX/2-28, y1-4, maxX/2-12, y1-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf("Failed to set view %q: %+v", yesButton, err)
		}
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		v.WriteString(CenterView("Yes", v))

		err = g.SetKeybinding(
			v.Name(), gocui.MouseLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
				selectedChan <- usernameArr[selected]
				return gocui.ErrQuit
			})
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for left mouse button: %+v", err)
		}

		err = g.SetKeybinding(
			v.Name(), gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
				selectedChan <- usernameArr[selected]
				return gocui.ErrQuit
			})
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for enter key: %+v", err)
		}

		g.SetCurrentView(v.Name())
	}

	if v, err := g.SetView(NoButton, maxX/2+12, y1-4, maxX/2+28, y1-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return errors.Errorf("Failed to set view %q: %+v", NoButton, err)
		}
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		v.WriteString(CenterView("No", v))

		err = g.SetKeybinding(
			v.Name(), gocui.MouseLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
				activeList = []string{usernameList, selectButton, cancelButton}
				active = 1
				for _, name := range []string{confirmDialogue, yesButton, NoButton} {
					err := g.DeleteView(name)
					if err != nil {
						return errors.Errorf(
							"Failed to delete view %q: %+v", name, err)
					}
				}
				return nil
			})
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for left mouse button: %+v", err)
		}

		err = g.SetKeybinding(
			v.Name(), gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
				activeList = []string{usernameList, selectButton, cancelButton}
				active = 1
				for _, name := range []string{confirmDialogue, yesButton, NoButton} {
					err := g.DeleteView(name)
					if err != nil {
						return errors.Errorf(
							"Failed to delete view %q: %+v", name, err)
					}
				}
				return nil
			})
		if err != nil {
			return errors.Errorf(
				"failed to set key binding for enter key: %+v", err)
		}
	}

	return nil

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
