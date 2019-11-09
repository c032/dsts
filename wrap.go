package dsts

import (
	"time"
)

const (
	// wrapDelay specifies for how long the text will "stay" at the beginning
	// and at the end, before starting to scroll to the opposite extreme.
	wrapDelay = 1 * time.Second

	// wrapSpeed specifies how much to wait for before "scrolling" to the next
	// character.
	wrapSpeed = 300 * time.Millisecond
)

// Wrap limits the width of a provider's statuses.
//
// When the underlying provider returns a status that surpasses the given width
// in characters, the wrapper will output at most `width` characters, and will
// automatically "scroll" back and forth, so the full status can be eventually
// read.
func Wrap(p Provider, width int) Provider {
	return func(chOut chan<- I3Status) {
		chIn := make(chan I3Status)

		// Start the inner provider.
		go p(chIn)

		var (
			sts     I3Status
			prevSts I3Status
		)

		// Main loop for the outer provider.
		//
		// Each message from the inner provider triggers a new iteration of
		// this loop.
		for {
			since := time.Now()
			offset := 0
			direction := 1

		Scroll:
			for {
				select {
				// The underlying provider sent a new message.
				case sts = <-chIn:
					// If the current message is the same as the previous one,
					// ignore it because we don't want to "reset" the current
					// position of the scroll.
					if sts != prevSts {
						prevSts = sts

						rs := []rune(sts.FullText)

						// Messages shorter or equal than the maximum width
						// don't need scrolling, so we just pass them as-is.
						if len(rs) <= width {
							chOut <- sts

							continue
						}

						text := string(rs[:width])
						chOut <- I3Status{
							FullText: text,
							Color:    sts.Color,
						}

						break Scroll
					}

				// Time to "scroll" to the next character.
				case <-time.After(wrapSpeed):
					rs := []rune(sts.FullText)

					// Messages shorter or equal than the maximum width don't
					// need scrolling, so we just pass them as-is.
					if len(rs) <= width {
						continue
					}

					if time.Now().Sub(since) < wrapDelay {
						continue
					}

					if direction == 1 {
						// We're displaying the rightmost part of the status.
						//
						// Wait for a bit and then reverse the direction of the
						// scrolling.
						if offset+width >= len(rs) {
							direction = -1
							since = time.Now()
							time.Sleep(wrapDelay)
						}
					} else {
						// We're displaying the leftmost part of the status.
						//
						// Wait for a bit and then reverse the direction of the
						// scrolling.
						if offset == 0 {
							direction = 1
							since = time.Now()
							time.Sleep(wrapDelay)
						}
					}

					//
					// Scroll by one character and send an update.
					//

					offset += direction

					text := string(rs[offset : offset+width])
					chOut <- I3Status{
						FullText: text,
						Color:    sts.Color,
					}
				}
			}
		}
	}
}
