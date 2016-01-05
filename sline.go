package main

import (
	"os"
	"strings"

	ui "github.com/gizak/termui"
)

// list where on item is selected
type SList struct {
	ui.List
	SelectableItems []string
	Selected        int
}

func NewSList() *SList {
	l := &SList{List: *ui.NewList()}
	return l
}

// Buffer implements ui.Bufferer interface
func (l *SList) Buffer() ui.Buffer {
	lselect := len(l.SelectableItems)

	if l.Selected >= lselect {
		l.Selected = lselect - 1
	}
	if l.Selected < 0 {
		l.Selected = 0
	}

	l.List.Items = make([]string, lselect)
	for i, item := range l.SelectableItems {
		if i == l.Selected {
			l.Items[i] = "[" + item + "](fg-white,bg-green)"
		} else {
			l.Items[i] = item
		}
	}
	return l.List.Buffer()
}

func (l *SList) SelectNext() {
	l.Selected += 1
	if l.Selected >= len(l.SelectableItems) {
		l.Selected -= 1
	}
}

func (l *SList) SelectPrevious() {
	if l.Selected == 0 {
		return
	}
	l.Selected -= 1
	if l.Selected < 0 {
		l.Selected = 0
	}
}

func (l *SList) CurrentSelection() string {
	if l.Selected >= len(l.SelectableItems) {
		return ""
	}
	return l.SelectableItems[l.Selected]
}

// PSList is where items are paged
type PSList struct {
	SList
	PageableItems []string
	PageSize      int
	CurrentPage   int
}

func NewPSList() *PSList {
	l := &PSList{SList: *NewSList()}
	return l
}

func (l *PSList) NextPage() {
	l.CurrentPage += 1
	if l.CurrentPage*l.PageSize >= len(l.PageableItems) {
		l.CurrentPage -= 1
		return
	}
	l.SList.Selected = 0
}

func (l *PSList) PreviousPage() {
	if l.CurrentPage == 0 {
		return
	}
	l.CurrentPage -= 1
	l.SList.Selected = 0
}

func (l *PSList) ResetPage() {
	l.CurrentPage = 0
}

func min(l, r int) int {
	if l < r {
		return l
	}
	return r
}

func max(l, r int) int {
	if l > r {
		return l
	}
	return r

}
func (l *PSList) Buffer() ui.Buffer {
	low := l.CurrentPage * l.PageSize
	l.SList.SelectableItems = l.PageableItems[l.CurrentPage*l.PageSize : min(len(l.PageableItems), low+l.PageSize)]
	return l.SList.Buffer()
}

// FSList where items can be filtered
type FSList struct {
	PSList
	FilterableItems []string
	Filter          string
}

func NewFSList() *FSList {
	l := &FSList{PSList: *NewPSList()}

	return l
}

func (l *FSList) AppendFilter(s string) {
	l.Filter = l.Filter + s
	l.CurrentPage = 0
	l.Selected = 0
}

// drop the last run, if any
func (l *FSList) TruncFilter() {
	fl := len(l.Filter)
	if fl <= 0 {
		return
	}
	l.Filter = l.Filter[:fl-1]
}

func (l *FSList) ResetFilter() {
	l.Filter = ""
	l.CurrentPage = 0
	l.Selected = 0
}
func (l *FSList) Buffer() ui.Buffer {
	l.PSList.PageableItems = filteredItems(l.Filter, l.FilterableItems)
	return l.PSList.Buffer()
}

func filterItem(filter string, item string) bool {
	i := 0
	for _, c := range filter {
		if i >= len(item) {
			return false
		}
		idx := strings.IndexRune(item[i:], c)
		if -1 == idx {
			return false
		} else {
			i = idx
		}
	}
	return true
}

func filteredItems(filter string, items []string) []string {
	r := make([]string, 0)
	for _, item := range items {
		if filterItem(filter, item) {
			r = append(r, item)
		}
	}
	return r
}

type DirPar struct {
	ui.Par
	Dir       string
	Searching bool
}

func (d *DirPar) str() string {
	prefix := "[- "
	middle := d.Dir
	postfix := "](fg-green)"
	if d.Searching {
		prefix = "[+ "
		postfix = "](fg-red)"
	}
	l := len(middle)
	w := ui.TermWidth()
	if l > w-2 {
		middle = "..." + middle[l-w+5:]
	}
	return prefix + middle + postfix
}

func NewDirPar() *DirPar {
	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dirPar := &DirPar{Par: *ui.NewPar(""), Dir: d, Searching: false}
	dirPar.Height = 1
	dirPar.Border = false
	return dirPar
}

func (d *DirPar) ToggleSearching() {
	d.Searching = !d.Searching
}

func (d *DirPar) Buffer() ui.Buffer {
	d.Par.Text = d.str()
	return d.Par.Buffer()
}

func (d *DirPar) Cd(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	cdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	d.Dir = cdir
}

type FilterPar struct {
	ui.Par
	Filter string
}

func NewFilterPar() *FilterPar {
	filterPar := &FilterPar{Par: *ui.NewPar(""), Filter: ""}
	filterPar.Height = 1
	filterPar.Border = false
	return filterPar
}

func (f *FilterPar) Buffer() ui.Buffer {
	f.Par.Text = "[* " + string(f.Filter) + "](fg-green)"
	return f.Par.Buffer()
}

func (f *FilterPar) AppendFilter(s string) {
	f.Filter = f.Filter + s
}

func (f *FilterPar) ResetFilter() {
	f.Filter = ""
}

func (f *FilterPar) TruncFilter() {
	fl := len(f.Filter)
	if fl <= 0 {
		return
	}
	f.Filter = f.Filter[:fl-1]
}

type Box struct {
	ui.Grid
	CDir    string
	top     *DirPar
	content *FSList
	filter  *FilterPar
}

func NewBox(grid *ui.Grid) *Box {
	h := ui.TermHeight()
	cDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	topLine := NewDirPar()

	dirContents := NewFSList()
	dir, err := os.Open(".")
	if err != nil {
		panic(err)
	}
	fileNames, err := dir.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	dirContents.FilterableItems = fileNames
	dirContents.PageSize = min(len(fileNames), max(0, h-2))
	dirContents.Border = false
	dirContents.Height = dirContents.PageSize

	bottomLine := NewFilterPar()

	grid.AddRows(ui.NewRow(ui.NewCol(12, 0, topLine)),
		ui.NewRow(ui.NewCol(12, 0, dirContents)),
		ui.NewRow(ui.NewCol(12, 0, bottomLine)))
	grid.Align()

	box := &Box{
		Grid:    *grid,
		CDir:    cDir,
		top:     topLine,
		content: dirContents,
		filter:  bottomLine,
	}
	return box
}

func (b *Box) CurrentSelection() string {
	return b.content.CurrentSelection()
}

func (b *Box) ToggleSearching() {
	b.top.ToggleSearching()
}

func (b *Box) CdUp() {
	b.cd("..")
}

func (b *Box) Resize() {
	h := ui.TermHeight()
	b.content.PageSize = min(len(b.content.FilterableItems), max(0, h-2))
	b.content.Height = b.content.PageSize
	b.Grid.Align()
}

func (b *Box) cd(dir string) {
	h := ui.TermHeight()
	b.top.Cd(dir)
	dirF, err := os.Open(".")
	if err != nil {
		panic(err)
	}
	fileNames, err := dirF.Readdirnames(-1)
	if err != nil {
		panic(err)
	}
	b.content.FilterableItems = fileNames
	b.content.PageSize = min(len(fileNames), max(0, h-2))
	b.content.Height = b.content.PageSize
	b.Grid.Align()
}

func (b *Box) NextPage() {
	b.content.NextPage()
}

func (b *Box) PreviousPage() {
	b.content.PreviousPage()
}

func (b *Box) SelectNext() {
	b.content.SelectNext()
}

func (b *Box) SelectPrevious() {
	b.content.SelectPrevious()
}

func (b *Box) AppendFilter(s string) {
	b.content.AppendFilter(s)
	b.filter.AppendFilter(s)
}

func (b *Box) TruncFilter() {
	b.content.TruncFilter()
	b.filter.TruncFilter()
}

func (b *Box) ResetFilter() {
	b.content.ResetFilter()
	b.filter.ResetFilter()
}
