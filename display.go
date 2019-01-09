package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
	"os/exec"
	"strconv"
)

func debug(x, y int, format string, v ...interface{}) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	s := fmt.Sprintf(format, v...)
	printString(x, y, 10000, s, fg, bg)
}

func drawDir(active int, count int, dir []os.FileInfo, offset, width int) {
	for i, f := range dir {
		str := f.Name()
		if f.IsDir() {
			str += "/"
		}

		fg := termbox.ColorDefault
		bg := termbox.ColorDefault

		if len(str) > width-4 {
			str = str[:width-3] + "..."
		}
		for len(str) < width {
			str += " "
		}

		if active == i {
			bg = termbox.ColorBlue
			if count > 0 {
				c := strconv.Itoa(count)
				str = str[:len(str)-(len(c)+1)] + c + " "
			}
		}

		if f.IsDir() {
			fg = termbox.ColorCyan | termbox.AttrBold
			if active == i {
				fg = termbox.ColorDefault | termbox.AttrBold
			}
		}
		printString(offset, i, width, str, fg, bg)
	}
}

func printString(x, y, maxWidth int, s string, fg, bg termbox.Attribute) {
	xstart := x
	for _, c := range s {
		if c == '\n' {
			x = xstart
			y++
		} else if c == '\r' {
			x = xstart
		} else {
			termbox.SetCell(x, y, c, fg, bg)
			x++
			if x > xstart+maxWidth {
				x = xstart
				y++
			}
		}
	}
}

func printPrompt(s string) {
	tbwidth, tbheight := termbox.Size()
	printString(tbwidth/4, tbheight/2, tbwidth,
		s, termbox.ColorDefault, termbox.ColorDefault)
	render()
}

func setupDisplay() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.SetOutputMode(termbox.OutputNormal)
}

func render() {
	termbox.Flush()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func draw(dt directoryTree, cd, userinput string) {
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth

	if cw < 0 {
		cw, _ = termbox.Size()
	}
	tbwidth, tbheight := termbox.Size()
	printString(0, tbheight-1, tbwidth,
		userinput, termbox.ColorDefault, termbox.ColorDefault)

	parentPath := getParentPath(cd)
	parentFiles := readDir(parentPath)
	if _, ok := dt[parentPath]; !ok {
		dt[parentPath] = dt.newDirForParent(cd)
	}
	offset := 0
	width := int(float64(cr[0]) / 10.0 * float64(cw))
	drawDir(dt[parentPath].active, 0, parentFiles, offset, width)

	files := readDir(".")

	offset = int(float64(cr[0])/10.0*float64(cw)) +
		int(float64(cr[1])/10.0*float64(cw))
	width = int(float64(cr[2]) / 10.0 * float64(cw))
	count := 0
	if len(files) > 0 {
		if files[dt[cd].active].IsDir() {
			childPath := cd + "/" + files[dt[cd].active].Name()
			files := readDir(childPath)
			count = len(files)
			if _, ok := dt[childPath]; !ok {
				dt[childPath] = &dir{active: 0}
			}
			drawDir(dt[childPath].active, 0, files, offset, width)
		} else if files[dt[cd].active].Size() < 100*1024*1024 {
			n := files[dt[cd].active].Name()
			cmd := exec.Command("cat", n)
			buf, _ := cmd.Output()
			buf = buf[:200]
			printString(offset, 0, width,
				string(buf), termbox.ColorDefault, termbox.ColorDefault)
		}
	}
	offset = int(float64(cr[0]) / 10.0 * float64(cw))
	width = int(float64(cr[1]) / 10.0 * float64(cw))
	drawDir(dt[cd].active, count, files, offset, width)

	render()
}
