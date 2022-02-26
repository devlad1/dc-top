package window

import (
	"context"
	"dc-top/docker"
	"log"

	"github.com/gdamore/tcell/v2"
)

type ShellWindow struct {
	id      string
	context context.Context
}

func NewShellWindow(id string, context context.Context) ShellWindow {
	return ShellWindow{
		id:      id,
		context: context,
	}
}

func (w *ShellWindow) Open(s tcell.Screen) {
	go w.main(s)
}

func (w *ShellWindow) Resize() {}

func (w *ShellWindow) KeyPress(_ tcell.EventKey) {}

func (w *ShellWindow) MousePress(_ tcell.EventMouse) {}

func (w *ShellWindow) HandleEvent(interface{}, WindowType) (interface{}, error) {
	log.Println("Shell window got event")
	panic(1)
}

func (w *ShellWindow) Close() {}

func (w *ShellWindow) main(s tcell.Screen) {
	var err error
	s.Suspend()
	defer func() {
		s.Resume()
		s.PostEvent(NewChangeToDefaultViewEvent())
		if err != nil {
			log.Println("resuming from exec")
			s.PostEvent(NewMessageEvent(Bar, ContainerShell, errorMessage{msg: []rune("Exec failed, container probably isn't running")}))
		}
	}()
	possible_shells := []string{"/bin/bash", "/bin/sh"}
	for _, sh := range possible_shells {
		err = docker.OpenShell(w.id, w.context, sh)
		if err == nil {
			return
		}
	}

}
