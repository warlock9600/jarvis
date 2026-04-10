package screensaver

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

type Options struct {
	NoColor bool
}

type drop struct {
	y      int
	length int
	speed  int
	tick   int
}

type cell struct {
	ch    rune
	style string
}

var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%^&*()[]{}<>/\\|+-_~")

func Run(opts Options) error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("screensaver requires interactive terminal")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer term.Restore(fd, oldState)

	fmt.Fprint(os.Stdout, "\033[?25l\033[2J\033[H")
	defer fmt.Fprint(os.Stdout, "\033[0m\033[?25h\033[2J\033[H")

	rand.Seed(time.Now().UnixNano())
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	if w < 20 || h < 8 {
		return fmt.Errorf("terminal too small for screensaver")
	}

	drops := make([]drop, w)
	for i := range drops {
		drops[i] = newDrop(h)
	}

	quitCh := make(chan struct{}, 1)
	go readKeys(quitCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(70 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-quitCh:
			return nil
		case <-sigCh:
			return nil
		case <-ticker.C:
			nw, nh, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil && (nw != w || nh != h) {
				w, h = nw, nh
				drops = make([]drop, w)
				for i := range drops {
					drops[i] = newDrop(h)
				}
			}
			stepDrops(drops, h)
			renderFrame(drops, w, h, opts)
		}
	}
}

func newDrop(height int) drop {
	return drop{
		y:      -rand.Intn(height),
		length: 4 + rand.Intn(max(5, height/3)),
		speed:  1 + rand.Intn(3),
	}
}

func stepDrops(drops []drop, height int) {
	for i := range drops {
		drops[i].tick++
		if drops[i].tick%drops[i].speed != 0 {
			continue
		}
		drops[i].y++
		if drops[i].y-drops[i].length > height {
			drops[i] = newDrop(height)
		}
	}
}

func renderFrame(drops []drop, w, h int, opts Options) {
	grid := make([][]cell, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]cell, w)
		for x := 0; x < w; x++ {
			grid[y][x] = cell{ch: ' ', style: ""}
		}
	}

	for x := 0; x < w; x++ {
		d := drops[x]
		for t := 0; t < d.length; t++ {
			y := d.y - t
			if y < 0 || y >= h {
				continue
			}
			ch := charset[rand.Intn(len(charset))]
			style := "\033[2;32m"
			if t == 0 {
				style = "\033[97m"
			} else if t < 3 {
				style = "\033[32m"
			}
			if opts.NoColor {
				style = ""
			}
			grid[y][x] = cell{ch: ch, style: style}
		}
	}

	var b strings.Builder
	b.Grow(w*h*8 + 64)
	b.WriteString("\033[H")
	for y := 0; y < h; y++ {
		curStyle := ""
		for x := 0; x < w; x++ {
			c := grid[y][x]
			if c.style != curStyle {
				if c.style == "" {
					b.WriteString("\033[0m")
				} else {
					b.WriteString(c.style)
				}
				curStyle = c.style
			}
			b.WriteRune(c.ch)
		}
		b.WriteString("\033[0m")
		if y < h-1 {
			b.WriteByte('\n')
		}
	}
	_, _ = os.Stdout.WriteString(b.String())
}

func readKeys(quitCh chan<- struct{}) {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			return
		}
		if buf[0] == 'q' || buf[0] == 'Q' || buf[0] == 3 {
			select {
			case quitCh <- struct{}{}:
			default:
			}
			return
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
