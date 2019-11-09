// Package mpd implements a dsts.Provider for displaying the currently playing
// song from MPD.
//
// See <https://www.musicpd.org/doc/html/protocol.html>.
package mpd

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/c032/dsts"
)

const (
	colorError  = "#e20024"
	colorNormal = "#ffffff"

	reconnectInterval = 1 * time.Second
	refreshInterval   = 300 * time.Millisecond
)

// makeError creates a status from an error.
func makeError(err error) dsts.I3Status {
	return dsts.I3Status{
		FullText: err.Error(),
		Color:    colorError,
	}
}

// Dial returns a `dsts.Provider` for an MPD listening on `addr`.
func Dial(addr string) dsts.Provider {
	return func(ch chan<- dsts.I3Status) {
		// Infinite loop so we always try to reconnect when some error occurs.
		for ; ; time.Sleep(reconnectInterval) {
			// Wrap iterations in a function just so that we can `defer`
			// without problem.
			//
			// Returning from this function triggers a reconnect.
			func() {
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					ch <- makeError(err)

					return
				}
				defer conn.Close()

				sc := bufio.NewScanner(conn)
				if !sc.Scan() {
					ch <- makeError(errors.New("unexpected eof"))

					return
				}

				greeting := sc.Text()
				if !strings.HasPrefix(greeting, "OK MPD ") {
					ch <- makeError(errors.New("unexpected mpd response"))

					return
				}

				//
				// We're ready to interact with MPD.
				//

				for ; ; time.Sleep(refreshInterval) {
					isPlaying := false

					// Check whether MPD is currently playing something.
					fmt.Fprintf(conn, "status\n")
					for sc.Scan() {
						line := sc.Text()
						if line == "OK" {
							break
						}

						components := strings.Split(line, ":")
						if len(components) == 2 {
							key := strings.ToLower(strings.TrimSpace(components[0]))
							value := strings.ToLower(strings.TrimSpace(components[1]))

							if key == "state" && value == "play" {
								isPlaying = true
							}
						}
					}

					err = sc.Err()
					if err != nil {
						ch <- makeError(err)

						return
					}

					// Hide the component while nothing is being played.
					if !isPlaying {
						ch <- dsts.I3Status{
							FullText: "",
							Color:    colorNormal,
						}

						continue
					}

					//
					// Retrieve song information.
					//
					// We need to read lines until we get a line containing
					// only "OK".
					//

					var (
						artist string
						album  string
						title  string
					)

					fmt.Fprintf(conn, "currentsong\n")
					for sc.Scan() {
						line := sc.Text()
						if line == "OK" {
							break
						}

						components := strings.SplitN(line, ":", 2)
						if len(components) == 2 {
							key := strings.ToLower(strings.TrimSpace(components[0]))
							value := strings.TrimSpace(components[1])

							switch key {
							case "artist":
								artist = value
							case "album":
								album = value
							case "title":
								title = value
							}
						}
					}

					if title == "" {
						continue
					}

					//
					// Format the output message and send it.
					//

					var currentlyPlaying string
					if album != "" {
						currentlyPlaying = "(" + album + ")"
					}
					if artist != "" {
						currentlyPlaying = strings.TrimSpace(artist+" "+currentlyPlaying) + " -"
					}

					currentlyPlaying = strings.TrimSpace(currentlyPlaying + " " + title)

					ch <- dsts.I3Status{
						FullText: currentlyPlaying,
						Color:    colorNormal,
					}
				}
			}()
		}
	}
}
