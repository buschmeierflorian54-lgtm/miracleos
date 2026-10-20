package main

import (
	"github.com/MiracleOS-Team/desktop/internal/apptabs"
	"strconv"
	"github.com/gotk3/gotk3/gtk"
	"path/filepath"
)

type explorerTab struct {
	frameID int
	path, query string
	history     []string
	historyAt   int
}

func (fm *glassFileManager) saveTab() {
	if fm.tabsBar == nil || len(fm.tabs) == 0 {
		return
	}
	query, _ := fm.search.GetText()
	fm.tabs[fm.activeTab] = explorerTab{frameID:fm.tabs[fm.activeTab].frameID,path:fm.current,query:query,history:append([]string(nil), fm.history...),historyAt:fm.historyAt}
}
func (fm *glassFileManager) openTab(path string) {
	fm.saveTab()
	fm.nextFrameTab++
	fm.tabs = append(fm.tabs, explorerTab{frameID:fm.nextFrameTab,path: path, historyAt: -1})
	fm.selectTab(len(fm.tabs) - 1)
}
func (fm *glassFileManager) selectTab(index int) {
	if index < 0 || index >= len(fm.tabs) {
		return
	}
	fm.activeTab = index
	tab := fm.tabs[index]
	fm.history = append([]string(nil), tab.history...)
	fm.historyAt = tab.historyAt
	fm.search.SetText(tab.query)
	fm.navigate(tab.path, len(tab.history) == 0)
}
func (fm *glassFileManager) closeTab(index int) {
	if len(fm.tabs) == 1 {
		fm.window.Destroy()
		return
	}
	fm.saveTab()
	fm.tabs = append(fm.tabs[:index], fm.tabs[index+1:]...)
	active := fm.activeTab
	if index < active || active >= len(fm.tabs) {
		active--
	}
	fm.selectTab(active)
}
func (fm *glassFileManager) renderTabs() {
	if fm.frameTabs!=nil {
		s:=apptabs.State{Name:"Dateien",CanNew:true}
		for i,t:=range fm.tabs {name:=filepath.Base(t.path);if t.path==fm.home{name="Home"};id:=strconv.Itoa(t.frameID);if i==fm.activeTab{s.Active=id};s.Tabs=append(s.Tabs,apptabs.Tab{ID:id,Title:name})}
		_ = fm.frameTabs.Publish(s)
	}
	if fm.tabsBar == nil {
		return
	}
	fm.tabsBar.GetChildren().Foreach(func(item interface{}) {
		if w, ok := item.(gtk.IWidget); ok {
			w.ToWidget().Destroy()
		}
	})
	for index, tab := range fm.tabs {
		index := index
		row, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)
		menuClass(row, "fm-tab")
		if index == fm.activeTab {
			menuClass(row, "fm-tab-active")
		}
		name := filepath.Base(tab.path)
		if tab.path == fm.home {
			name = "Home"
		}
		button, _ := gtk.ButtonNewWithLabel(name)
		button.SetTooltipText(tab.path)
		button.Connect("clicked", func() { fm.saveTab(); fm.selectTab(index) })
		close, _ := gtk.ButtonNewWithLabel("×")
		close.SetTooltipText("Tab schließen")
		close.Connect("clicked", func() { fm.closeTab(index) })
		row.PackStart(button, false, false, 0)
		row.PackStart(close, false, false, 0)
		fm.tabsBar.PackStart(row, false, false, 0)
	}
	fm.tabsBar.PackStart(fm.control("+", "Neuer Tab", func() { fm.openTab(fm.home) }), false, false, 0)
	fm.tabsBar.ShowAll()
}
