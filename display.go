package main

// Display files are complicated...sorry its so silly

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"os"
	"os/exec"
	"strconv"
)

const (
	topOffset = 1
	fgDefault = termbox.Attribute(0xe0)
)

func drawDir(active int, count int, selected map[string]bool, dir []pseudofile, offset, width int) {
	_, tbheight := termbox.Size()
	viewbox := tbheight - 2
	oob := 0
	// are we off the edge of the display
	if active+tbheight/2 > viewbox {
		oob = (active + tbheight/2) - viewbox
		if len(dir[oob:]) < viewbox {
			oob -= tbheight - 2 - len(dir[oob:])
		}
		if oob < 0 {
			oob = 0
		}
		dir = dir[oob:]
	}
	for i, f := range dir {
		if i+topOffset == tbheight-1 {
			break
		}
		str := f.name
		if f.isDir {
			str += "/"
		}
		if f.isLink {
			if f.link.broken {
				str += " ~> " + f.link.location
			} else {
				str += " -> " + f.link.location
			}
		}
		if selected[f.name] {
			str = " " + str
		}

		if len(str) > width-4 {
			str = str[:width-3] + ".."
		}
		for len(str) < width-1 {
			str += " "
		}

		a := (active == i+oob)
		// Append count to end if dir
		if f.isDir && a {
			c := strconv.Itoa(count)
			str = str[:len(str)-(len(c)+1)] + c + " "
		}
		if f.isLink && a && f.link.location != "" {
			if cf, err := os.Stat(f.link.location); err == nil && cf.IsDir() {
				c := strconv.Itoa(count)
				str = str[:len(str)-(len(c)+1)] + c + " "
			}
		}
		fg, bg := getColors(f, a, selected[f.name])

		printString(offset, i+topOffset, width, str, true, fg, bg)
	}
}

func getColors(f pseudofile, active, selected bool) (termbox.Attribute, termbox.Attribute) {
	fg := fgDefault
	bg := termbox.ColorDefault
	if active {
		bg = termbox.ColorBlue
	}

	if f.isDir {
		fg = termbox.ColorCyan
		if active {
			fg = fgDefault
		}
		fg |= termbox.AttrBold
	} else {

		if !f.isReal {
			fg = fgDefault
		} else if (f.f.Mode()&0111) != 0 && !f.isLink {
			fg = termbox.ColorYellow | termbox.AttrBold
		} else if f.isLink && f.link.location != "" {
			fg = termbox.ColorMagenta | termbox.AttrBold
			if cf, err := os.Stat(f.link.location); err == nil && cf.IsDir() {
				fg = termbox.ColorBlue | termbox.AttrBold
			}
			if f.link.broken {
				fg = termbox.ColorRed | termbox.AttrBold
			}
		}
	}
	if selected {
		fg = termbox.ColorYellow | termbox.AttrBold
	}
	return fg, bg
}

func drawParentDir(files []pseudofile, s *fmState, count int) {
	tbwidth, _ := termbox.Size()
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth
	if cw < 0 {
		cw = tbwidth
	}
	parentPath := getParentPath(s.cd)
	parentFiles, _, _ := readDir(parentPath)
	if _, ok := s.dt[parentPath]; !ok {
		s.dt[parentPath] = s.dt.newDirForParent(s.cd)
	}
	// Draw parent dir in first column
	width := int(float64(cr[0]) / 10.0 * float64(cw))
	drawDir(s.dt[parentPath].active, count, s.selectedFiles, parentFiles, 0, width)

}

func drawChildDir(parent pseudofile, s *fmState, count *int) {
	tbwidth, tbheight := termbox.Size()
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth
	if cw < 0 {
		cw = tbwidth
	}
	offset := int(float64(cr[0])/10.0*float64(cw)) +
		int(float64(cr[1])/10.0*float64(cw))
	width := int(float64(cr[2]) / 10.0 * float64(cw))
	// Draw child directory or preview file < 100KB in last column
	if parent.isDir {
		childPath := s.cd + "/" + parent.name
		if s.cd == "/" {
			childPath = s.cd + parent.name
		}
		files, c, err := readDir(childPath)

		if !os.IsPermission(err) {
			if files[0].isReal {
				*count = c
			}
			if _, ok := s.dt[childPath]; !ok {
				s.dt[childPath] = &dir{active: 0}
			}
			drawDir(s.dt[childPath].active, 0, s.selectedFiles, files, offset, width)
		}
	} else if parent.isLink && parent.link.location != "" && !parent.link.broken {
		if f, err := os.Stat(parent.link.location); f.IsDir() && err == nil {
			childP := parent.link.location
			files, c, err := readDir(childP)
			if !os.IsPermission(err) && len(files) > 0 {
				if files[0].isReal {
					*count = c
				}
				if _, ok := s.dt[childP]; !ok {
					s.dt[childP] = &dir{active: 0}
				}
				drawDir(s.dt[childP].active, 0, s.selectedFiles, files, offset, width)
			}
		}
	} else if parent.isReal &&
		parent.f.Size() < 100*1024*1024 {

		n := parent.name
		cmd := exec.Command("cat", n)
		buf, _ := cmd.Output()
		if len(buf) > cw*tbheight-2 {
			buf = buf[:cw*tbheight-2]
		}
		printString(offset, topOffset, width,
			string(buf), conf.WrapText, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func drawHeader(userinput string, files []pseudofile, dt directoryTree, cd string) {
	tbwidth, _ := termbox.Size()
	// Print user/cd at top
	un := os.Getenv("USER")
	hn, err := os.Hostname()
	if err != nil {
		log.Errorln(err)
	}
	ustr := un + "@" + hn
	printString(0, 0, tbwidth, ustr, true, termbox.ColorGreen, termbox.ColorDefault)
	dn := cd
	oset := 0
	if cd != "/" {
		dn += "/"
		oset = 1
	}

	printString(len(ustr)+1, 0, tbwidth, dn, true, termbox.ColorBlue, termbox.ColorDefault)
	f := files[dt[cd].active]
	name := f.name
	if f.isDir {
		name += "/"
	}
	printString(len(ustr)+len(cd)+1+oset, 0, tbwidth, name,
		true, termbox.ColorDefault, termbox.ColorDefault)
}

func drawFooter(userinput string, files []pseudofile, dt directoryTree, cd string) {
	tbwidth, tbheight := termbox.Size()
	if len(userinput) > 0 {
		printString(0, tbheight-1, tbwidth,
			userinput+"█", true, termbox.ColorDefault, termbox.ColorDefault)
	} else {
		f := files[dt[cd].active]
		if f.isReal {
			s := fmt.Sprintf("%s %d %s %s",
				f.f.Mode(), f.f.Size(),
				f.f.ModTime().Format("Jan 2 15:04"), f.name)
			printString(0, tbheight-1, tbwidth,
				s, true, termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func draw(s *fmState) {

	tbw, tbh := termbox.Size()
	if tbw <= 0 || tbh <= 0 {
		return
	}
	files, amtFiles, err := readDir(".")
	if err != nil {
		log.Errorln(err)
	}

	// draw parent
	drawParentDir(files, s, amtFiles)
	childCount := 0
	drawChildDir(files[s.dt[s.cd].active], s, &childCount)

	{ // Draw current directory
		tbw, _ := termbox.Size()
		cr := conf.ColumnRatios
		cw := conf.ColumnWidth
		if cw < 0 {
			cw = tbw
		}
		offset := int(float64(cr[0]) / 10.0 * float64(cw))
		width := int(float64(cr[1]) / 10.0 * float64(cw))
		drawDir(s.dt[s.cd].active, childCount, s.selectedFiles, files, offset, width)
	}

	drawHeader(s.cmd, files, s.dt, s.cd)

	// draw footer for frame
	drawFooter(s.cmd, files, s.dt, s.cd)
	render()
}

func setupDisplay() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	// termbox.SetOutputMode(termbox.OutputNormal)
	termbox.SetOutputMode(termbox.Output256)
}

func render() {
	termbox.Flush()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func printString(x, y, maxWidth int, s string, wrap bool, fg, bg termbox.Attribute) {
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
				if !wrap {
					break
				}
				x = xstart
				y++
			}
		}
	}
}

func printPrompt(s string) {
	tbwidth, tbheight := termbox.Size()
	printString(tbwidth/4, tbheight/2, tbwidth,
		s, true, termbox.ColorDefault, termbox.ColorDefault)
	render()
}
