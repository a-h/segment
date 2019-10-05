package segment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stianeikeland/go-rpio"
)

// FourDigitSevenSegment allows control of a 4 digit, 7 segment display.
type FourDigitSevenSegment struct {
	digitPins   [4]rpio.Pin
	segmentPins [8]rpio.Pin
	// String to display.
	s string
}

// NewFourDigitSevenSegmentDisplay creates a new FourDigitSevenSegmentDisplay.
// The pins are the GPIO pins associated with the 12 pins of the 3461BS.
func NewFourDigitSevenSegmentDisplay(pD1, pa, pf, pD2, pD3, pb, pe, pd, pdp, pc, pg, pD4 rpio.Pin) *FourDigitSevenSegment {
	pD1.Output()
	pD2.Output()
	pD3.Output()
	pD4.Output()
	d := &FourDigitSevenSegment{
		digitPins:   [4]rpio.Pin{pD1, pD2, pD3, pD4},
		segmentPins: [8]rpio.Pin{pa, pb, pc, pd, pe, pf, pg, pdp},
	}
	return d
}

// ErrStringTooLong is returned when more than 4 characters are attempted to be displayed.
var ErrStringTooLong = errors.New("cannot display string because it is too long")

// Start the display.
func (d *FourDigitSevenSegment) Start(ctx context.Context, s string) (cancel func()) {
	d.Update(s)
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				d.render()
			}
		}
	}()
	return
}

// Update the display.
func (d *FourDigitSevenSegment) Update(s string) {
	d.s = strings.ToUpper(strings.TrimSpace(s))
}

// Render the string (at most 4 digits). Needs to be called at least once per 10ms to
// show up.
func (d *FourDigitSevenSegment) render() {
	if len(d.s) < 5 {
		d.renderSegments((d.s + "    ")[:4])
		return
	}
	for i := 0; i < len(d.s); i++ {
		// Render each set of segments 50 times to give the user
		// 50ms to view the character.
		seg := (d.s + "    ")[i : i+4]
		for j := 0; j < 50; j++ {
			d.renderSegments(seg)
		}
	}
}

func (d *FourDigitSevenSegment) renderSegments(s string) {
	for i, r := range s {
		d.light(i, characterToSegments[r])
	}
}

func (d *FourDigitSevenSegment) light(index int, values [8]bool) {
	// Switch off the light while we change things.
	d.digitPins[0].Low()
	d.digitPins[1].Low()
	d.digitPins[2].Low()
	d.digitPins[3].Low()

	// Set the correct segments.
	var allDown bool
	for i, v := range values {
		m := rpio.PullDown
		if !v {
			m = rpio.PullUp
			allDown = false
		}
		d.segmentPins[i].Pull(m)
	}
	if allDown {
		// Don't waste time if there's nothing to display.
		return
	}

	// Light up the digit.
	d.digitPins[index].High()

	// Give it time to shine.
	time.Sleep(time.Millisecond)
}

var characterToSegments = map[rune][8]bool{
	'1': [8]bool{false, true, true, false, false, false, false, false},
	'2': [8]bool{true, true, false, true, true, false, true, false},
	'3': [8]bool{true, true, true, true, false, false, true, false},
	'4': [8]bool{false, true, true, false, false, true, true, false},
	'5': [8]bool{true, false, true, true, false, true, true, false},
	'6': [8]bool{true, false, true, true, true, true, true, false},
	'7': [8]bool{true, true, true, false, false, false, false, false},
	'8': [8]bool{true, true, true, true, true, true, true, false},
	'9': [8]bool{true, true, true, true, false, true, true, false},
	'0': [8]bool{true, true, true, true, true, true, false, false},
	'A': [8]bool{true, true, true, false, true, true, true, false},
	'B': [8]bool{false, false, true, true, true, true, true, false},
	'C': [8]bool{true, false, false, true, true, true, false, false},
	'D': [8]bool{false, true, true, true, true, false, true, false},
	'E': [8]bool{true, false, false, true, true, true, true, false},
	'F': [8]bool{true, false, false, false, true, true, true, false},
	'G': [8]bool{true, true, true, true, false, true, true, false},
	'H': [8]bool{false, false, true, false, true, true, true, false},
	'I': [8]bool{false, true, true, false, false, false, false, false},
	'J': [8]bool{false, true, true, true, true, false, false, false},
	'K': [8]bool{true, false, true, false, true, true, true, false},
	'L': [8]bool{false, false, false, true, true, true, false, false},
	'M': [8]bool{true, true, true, false, true, true, false, true},
	'N': [8]bool{false, false, true, false, true, false, true, false},
	'O': [8]bool{false, false, true, true, true, false, true, false},
	'P': [8]bool{true, true, false, false, true, true, true, false},
	'Q': [8]bool{true, true, true, false, false, true, true, false},
	'R': [8]bool{false, false, false, false, true, false, true, false},
	'S': [8]bool{true, false, true, true, false, true, true, false},
	'T': [8]bool{true, true, true, false, false, false, false, false},
	'U': [8]bool{false, false, true, true, true, false, false, false},
	'V': [8]bool{false, true, true, true, true, true, false, true},
	'W': [8]bool{false, true, true, true, true, true, false, false},
	'X': [8]bool{false, true, true, false, true, true, true, true},
	'Y': [8]bool{false, true, true, true, false, true, true, false},
	'Z': [8]bool{true, true, false, true, true, false, true, true},
	'.': [8]bool{false, false, false, false, false, false, false, true},
	'-': [8]bool{false, false, false, false, false, false, true, false},
	'#': [8]bool{true, true, true, true, true, true, true, true},
}
