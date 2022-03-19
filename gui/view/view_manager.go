package view

import (
	"context"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"dc-top/gui/view/window/container_logs_window"
	"dc-top/gui/view/window/containers_window"
	"dc-top/gui/view/window/docker_info_window"
	"dc-top/gui/view/window/general_info_window"
	"dc-top/gui/view/window/help_window"
	"log"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type _viewName uint8

const (
	main _viewName = iota
	logs
	main_help
	logs_help
	none
)

var _views map[_viewName]*View = make(map[_viewName]*View)

// var _curr_view _viewName

// var _prev_view _viewName
var _view_stack viewStack = newStack()
var _lock sync.Mutex

func InitDefaultView() {
	_lock.Lock()
	defer _lock.Unlock()

	default_view_ctx, default_view_cancel := context.WithCancel(context.TODO())

	log.Printf("Initiating default view")
	def_general_info_w := general_info_window.NewGeneralInfoWindow(default_view_ctx)
	def_containers_w := containers_window.NewContainersWindow()
	def_docker_info_w := docker_info_window.NewDockerInfoWindow()

	controls := help_window.MainControls()
	def_help_w := help_window.NewHelpWindow(
		default_view_ctx,
		controls,
		func() window.Dimensions {
			x1, y1, x2, y2 := window.MainHelpWindowSize()
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
	)

	bar_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.ContainersBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	def_bar_w := bar_window.NewBarWindow(ctx, cancel, bar_dimensions_generator)

	default_view := NewView(map[window.WindowType]window.Window{
		window.GeneralInfo:      &def_general_info_w,
		window.ContainersHolder: &def_containers_w,
		window.DockerInfo:       &def_docker_info_w,
		window.Bar:              &def_bar_w,
		window.Help:             &def_help_w,
	}, window.ContainersHolder)

	_views[main] = &default_view
	_view_stack.push(main)
	default_view.Open()
}

func ReturnToUpperView() {
	_lock.Lock()
	defer _lock.Unlock()
	if _view_stack.peek() != none {
		log.Println("Returning to previous view")
		_views[currentViewName()].Close()
		_view_stack.pop()
		window.GetScreen().Clear()
		_views[currentViewName()].ResumeWindows()
	} else {
		log.Println("Illegal upper view attempt")
	}
}

func DisplayMainHelp() {
	log.Printf("Changing to main help")
	changeToHelpView(main_help, main, help_window.MainControls())
}

func ChangeToLogView(container_id string) {
	log.Printf("Changing to logs")
	logs_window := container_logs_window.NewContainerLogsWindow(container_id)
	bar_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.LogsBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	logs_bar_w := bar_window.NewBarWindow(ctx, cancel, bar_dimensions_generator)
	logs_view := NewView(map[window.WindowType]window.Window{
		window.ContainerLogs: &logs_window,
		window.Bar:           &logs_bar_w,
	}, window.ContainerLogs)
	changeView(logs, main, &logs_view)
}

func DisplayLogHelp() {
	log.Printf("Changing to log help")
	changeToHelpView(logs_help, logs, help_window.LogControls())
}

func HandleKeyPress(key *tcell.EventKey) {
	_lock.Lock()
	defer _lock.Unlock()
	_views[currentViewName()].GetWindow(_views[currentViewName()].GetFocusedWindow()).KeyPress(*key)
}

func HandleMouseEvent(ev *tcell.EventMouse) {
	_lock.Lock()
	defer _lock.Unlock()
	switch currentViewName() {
	case main:
		DefaultView().GetWindow(window.ContainersHolder).MousePress(*ev)
	}
}

func CurrentView() *View {
	_lock.Lock()
	defer _lock.Unlock()
	return _views[currentViewName()]
}

func DefaultView() *View {
	return _views[main]
}

func LogView() *View {
	return _views[logs]
}

func CloseAll() {
	for _, view := range _views {
		view.Close()
	}
}

func changeToHelpView(new_view_key, prev_view_key _viewName, controls []help_window.Control) {
	help_window := help_window.NewHelpWindow(
		context.TODO(),
		controls,
		func() window.Dimensions {
			width, height := window.GetScreen().Size()
			x1, y1, x2, y2 := width/3, height/3, 2*width/3, (height/3 + len(controls) + 3)
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
	)
	help_view := NewView(map[window.WindowType]window.Window{
		window.Help: &help_window,
	}, window.Help)
	changeView(new_view_key, prev_view_key, &help_view)
}

func changeView(new_view_key, prev_view_key _viewName, view *View) {
	_lock.Lock()
	defer _lock.Unlock()
	_views[prev_view_key].PauseWindows()
	_views[new_view_key] = view
	_view_stack.push(new_view_key)
	view.Open()
}

func currentViewName() _viewName {
	return _view_stack.peek()
}
