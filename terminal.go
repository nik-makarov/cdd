package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

var errCancelled = errors.New("selection cancelled")

type keyEvent struct {
	name string
	text string
}

func runTerminal(m *browser, input, output *os.File) (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()
	resize := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	defer signal.Stop(resize)
	fd := int(input.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("enable raw terminal mode: %w", err)
	}
	defer func() {
		_, screenErr := io.WriteString(output, "\x1b[?25h\x1b[?1049l")
		restoreErr := term.Restore(fd, state)
		err = errors.Join(err, screenErr, restoreErr)
	}()
	if _, err := io.WriteString(output, "\x1b[?1049h\x1b[?25l"); err != nil {
		return fmt.Errorf("open terminal screen: %w", err)
	}
	redraw := func() error {
		if width, height, err := term.GetSize(int(output.Fd())); err == nil && width > 0 && height > 0 {
			m.width, m.height = width, height
		}
		_, err := io.WriteString(output, "\x1b[H\x1b[2J"+strings.ReplaceAll(m.view(), "\n", "\r\n"))
		return err
	}
	if err := redraw(); err != nil {
		return fmt.Errorf("draw terminal: %w", err)
	}
	var pending []byte
	var lastInput time.Time
	var buffer [4096]byte
	poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	for {
		select {
		case <-ctx.Done():
			return errCancelled
		case <-resize:
			if err := redraw(); err != nil {
				return fmt.Errorf("redraw terminal: %w", err)
			}
		default:
		}
		if _, err := unix.Poll(poll, 50); err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return fmt.Errorf("wait for terminal input: %w", err)
		}
		if poll[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return errCancelled
		}
		if poll[0].Revents&unix.POLLIN != 0 {
			n, err := unix.Read(fd, buffer[:])
			if err != nil {
				if errors.Is(err, unix.EINTR) || errors.Is(err, unix.EAGAIN) {
					continue
				}
				return fmt.Errorf("read terminal: %w", err)
			}
			if n == 0 {
				return errCancelled
			}
			pending = append(pending, buffer[:n]...)
			lastInput = time.Now()
		}
		changed := false
		for len(pending) > 0 {
			key, consumed := decodeKey(pending, time.Since(lastInput) >= 50*time.Millisecond)
			if consumed == 0 {
				break
			}
			pending = pending[consumed:]
			if key.name == "" && key.text == "" {
				continue
			}
			if m.handle(key) {
				return nil
			}
			changed = true
		}
		if changed {
			if err := redraw(); err != nil {
				return fmt.Errorf("redraw terminal: %w", err)
			}
		}
	}
}

// decodeKey accepts fragmented UTF-8 and the usual CSI/SS3 terminal keys.
// Unknown escape sequences are consumed, never inserted into the filter.
func decodeKey(data []byte, expired bool) (keyEvent, int) {
	if len(data) == 0 {
		return keyEvent{}, 0
	}
	if data[0] == '\x1b' {
		if len(data) == 1 {
			if expired {
				return keyEvent{name: "esc"}, 1
			}
			return keyEvent{}, 0
		}
		if data[1] != '[' && data[1] != 'O' {
			// Alt+character. Consume the whole rune, including fragmented UTF-8.
			if !utf8.FullRune(data[1:]) && !expired {
				return keyEvent{}, 0
			}
			_, n := utf8.DecodeRune(data[1:])
			return keyEvent{}, 1 + n
		}
		for i := 2; i < len(data); i++ {
			if data[i] >= 0x40 && data[i] <= 0x7e {
				sequence := string(data[2 : i+1])
				name := ""
				switch sequence {
				case "A":
					name = "up"
				case "B":
					name = "down"
				case "C":
					name = "right"
				case "D":
					name = "left"
				case "H", "1~", "7~":
					name = "home"
				case "F", "4~", "8~":
					name = "end"
				case "5~":
					name = "pgup"
				case "6~":
					name = "pgdown"
				}
				return keyEvent{name: name}, i + 1
			}
		}
		if expired || len(data) > 64 {
			return keyEvent{}, len(data)
		}
		return keyEvent{}, 0
	}
	switch data[0] {
	case 3:
		return keyEvent{name: "ctrl+c"}, 1
	case 4:
		return keyEvent{name: "ctrl+d"}, 1
	case 18:
		return keyEvent{name: "ctrl+r"}, 1
	case '\r', '\n':
		return keyEvent{name: "enter"}, 1
	case '\b', 127:
		return keyEvent{name: "backspace"}, 1
	}
	if !utf8.FullRune(data) && !expired {
		return keyEvent{}, 0
	}
	r, n := utf8.DecodeRune(data)
	if !unicode.IsPrint(r) || r == utf8.RuneError {
		return keyEvent{}, n
	}
	text := string(r)
	return keyEvent{name: text, text: text}, n
}
