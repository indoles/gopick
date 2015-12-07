package main

import (
	"fmt"
	"os"

	ui "github.com/gizak/termui"
)

func isDir(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Mode().IsDir()
}

func main() {
	output := ""

	searching := false

	err := ui.Init()
	if err != nil {
		panic(err)

	}
	defer func() {
		fmt.Print(output)
	}()
	defer ui.Close()
	content := NewBox(ui.Body)
	ui.Body = &content.Grid
	ui.Render(ui.Body)

	ui.Handle("/sys/wnd/resize)", func(ui.Event) {
		content.Resize()
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		path, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = path + "/" + content.CurrentSelection()
		if isDir(path) {
			content.cd(path)
			content.ResetFilter()
			ui.Render(ui.Body)
		} else {
			output, err = os.Getwd()
			if err != nil {
				panic(err)
			}
			output += "/" + content.CurrentSelection()
			ui.StopLoop()
		}
	})
	ui.Handle("/sys/kbd/<escape>", func(ui.Event) {
		if searching {
			searching = false
			content.ToggleSearching()
			ui.Render(ui.Body)
		} else {
			output, err = os.Getwd()
			if err != nil {
				panic(err)
			}
			ui.StopLoop()
		}
	})
	ui.Handle("/sys/kbd/C-8", func(ui.Event) {
		content.TruncFilter()
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
		ui.Close()
		os.Exit(-1)
	})
	ui.Handle("/sys/kbd/C-a", func(ui.Event) {
		content.ResetFilter()
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/C-r", func(ui.Event) {
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/i", func(ui.Event) {
		if searching {
			content.AppendFilter("i")
		} else {
			searching = true
			content.ToggleSearching()
		}
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/n", func(ui.Event) {
		if searching {
			content.AppendFilter("n")
		} else {
			content.NextPage()
		}
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/p", func(ui.Event) {
		if searching {
			content.AppendFilter("p")
		} else {
			content.PreviousPage()
		}
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/j", func(ui.Event) {
		if searching {
			content.AppendFilter("j")
		} else {
			content.SelectNext()
		}
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/k", func(ui.Event) {
		if searching {
			content.AppendFilter("k")
		} else {
			content.SelectPrevious()
		}
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd/,", func(ui.Event) {
		content.CdUp()
		content.ResetFilter()
		ui.Render(ui.Body)
	})
	ui.Handle("/sys/kbd", func(e ui.Event) {
		output += " " + e.Path
		if e.Path == "/sys/kbd//" {
			content.cd("/")
			content.ResetFilter()

		} else {
			if !searching {
				searching = true
				content.ToggleSearching()
			}
			content.AppendFilter(e.Data.(ui.EvtKbd).KeyStr)
		}
		ui.Render(ui.Body)
	})
	ui.Loop()
}
