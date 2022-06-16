package menu

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/mattn/go-runewidth"

	"github.com/ezzer17/backconnectd/internal/client"
	"github.com/ezzer17/backconnectd/internal/rpc"
	"github.com/ezzer17/backconnectd/internal/sessioninfo"
)

// Menu is a slice of sessions.
// It also stores selected session
type Menu struct {
	items        []*sessioninfo.SessionInfo
	selected     int
	screen       tcell.Screen
	screenEvents chan tcell.Event
	serverEvents chan *rpc.Event
	serverErrors chan error
	quit         chan struct{}
	client       *client.Client
}

// New returns an empty Menu
func New(screen tcell.Screen, client *client.Client) *Menu {
	items := make([]*sessioninfo.SessionInfo, 0)
	m := Menu{items,
		0, screen,
		make(chan tcell.Event), make(chan *rpc.Event),
		make(chan error),
		make(chan struct{}), client,
	}
	// m.Render()
	return &m
}

// Append appends item to Menu
func (m *Menu) Append(item *sessioninfo.SessionInfo) {
	m.items = append(m.items, item)
	m.Render()
}

// DeleteByID deletes item with correspondind ID from slice
func (m *Menu) DeleteByID(id uuid.UUID) bool {
	for index, value := range m.items {
		if value.ID() == id {
			m.items[len(m.items)-1], m.items[index] = m.items[index], m.items[len(m.items)-1]
			m.items = m.items[:len(m.items)-1]
			return true
		}
	}
	m.Render()
	return false
}

// Get returns item from Menu with given index
func (m *Menu) Get(idx int) (*sessioninfo.SessionInfo, bool) {
	if idx >= len(m.items) || idx < 0 {
		return nil, false
	}
	return m.items[idx], true
}

// Current returns currently selected item
func (m *Menu) Current() *sessioninfo.SessionInfo {
	if m.selected > len(m.items)-1 {
		m.selected = len(m.items) - 1
	}
	return m.items[m.selected]
}

// Down moves selected ptr up if possible
func (m *Menu) Down() {
	if m.selected < len(m.items)-1 {
		m.selected++
	} else {
		m.selected = len(m.items) - 1
	}
	m.Render()
}

// Up moves selected ptr down if possible
func (m *Menu) Up() {
	if m.selected > 0 {
		m.selected--
	}
	m.Render()
}

// Len returns length of Menu
func (m *Menu) Len() int {
	return len(m.items)
}

// Items returns a slice of items in Menu
func (m *Menu) Items() []*sessioninfo.SessionInfo {
	return m.items
}

func (m *Menu) emitStr(x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		m.screen.SetContent(x, y, c, comb, style)
		x += w
	}
}

func (m *Menu) Render() {
	var selectedStyle tcell.Style
	m.screen.Clear()
	if len(m.items) > 0 {
		for n, sess := range m.items {
			if n == m.selected {
				selectedStyle = tcell.StyleDefault.Foreground(tcell.ColorCadetBlue.TrueColor()).Background(tcell.ColorWhite)
			} else {
				selectedStyle = tcell.StyleDefault
			}
			m.emitStr(0, n, selectedStyle, sess.String())
		}
	} else {
		w, h := m.screen.Size()
		m.emitStr(w/2, h/2, selectedStyle, "No sessions available!")
	}
	m.screen.Sync()
}

func (m *Menu) Init() error {
	if err := m.screen.Init(); err != nil {
		return err
	}
	return nil
}

func (m *Menu) evloop() {
	for {
		select {
		case ev := <-m.screenEvents:
			switch evt := ev.(type) {
			case *tcell.EventResize:
				m.screen.Sync()
				m.Render()
			case *tcell.EventKey:
				if evt.Key() == tcell.KeyEscape {
					m.quit <- struct{}{}
					return
				}
				if evt.Key() == tcell.KeyUp {
					m.Up()
				}
				if evt.Key() == tcell.KeyRune {
					if evt.Rune() == 'k' {
						m.client.SessionKill(m.Current().ID())
					}
				}
				if evt.Key() == tcell.KeyDown {
					m.Down()
				}
				if evt.Key() == tcell.KeyEnter {
					m.client.SessionConnect(m.Current().ID())
					m.quit <- struct{}{}
					m.screen.Fini()
					readch := make(chan []byte)
					writech := make(chan []byte)
					// quitch := make(chan struct{})
					go m.client.ChanelRawData(readch, writech)
					ChannelTerm(writech, readch) //???
					go m.client.ChanelEvents(m.serverEvents, m.serverErrors)
					go m.screen.ChannelEvents(m.screenEvents, m.quit)
				}
			}
		case <-time.Tick(time.Second):
			m.Render()
		case evt := <-m.serverEvents:
			switch evt.EType {
			case rpc.Delete:
				if ok := m.DeleteByID(evt.Sess.ID()); !ok {
					panic("No such session")
				}
			case rpc.Add:
				m.Append(&evt.Sess)
			}
		case sErr := <-m.serverErrors:
			if _, ok := sErr.(client.ConnectionClosed); ok {
				m.quit <- struct{}{}
				return
			}
			fmt.Printf("Error %s\n", sErr)
		}
	}
}

func (m *Menu) RunTillExit() {
	go m.client.ChanelEvents(m.serverEvents, m.serverErrors)
	go m.screen.ChannelEvents(m.screenEvents, m.quit)
	m.evloop()
	m.screen.Fini()
}
