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
)

const boldS = "\x1b[38;1;255m"
const boldU = "\x1b[0m"
const dimS = "\x1b[38;5;250m"
const dimU = "\x1b[0m"

// Bold returns the text as bold.
func Bold(s string) string {
	return boldS + s + boldU
}

// Dim returns the text as dim.
func Dim(s string) string {
	return dimS + s + dimU
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
