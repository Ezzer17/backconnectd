package menu

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/mattn/go-runewidth"

	"github.com/ezzer17/backconnectd/internal/client"
	"github.com/ezzer17/backconnectd/internal/sessioninfo"
	"github.com/ezzer17/backconnectd/pkg/tube"
	pb "github.com/ezzer17/backconnectd/proto"
)

// Menu is a slice of sessions.
// It also stores selected session
type Menu struct {
	items        []*sessioninfo.SessionInfo
	selected     int
	screen       tcell.Screen
	screenEvents chan tcell.Event
	serverEvents chan *pb.SessionEvent
	serverErrors chan error
	errors       *tube.Tube
	quit         chan struct{}
	client       *client.Client
}

// New returns an empty Menu
func New(screen tcell.Screen, client *client.Client) *Menu {
	items := make([]*sessioninfo.SessionInfo, 0)
	m := Menu{items,
		0, screen,
		make(chan tcell.Event), make(chan *pb.SessionEvent),
		make(chan error, 3), tube.New(3),
		make(chan struct{}), client,
	}
	return &m
}

// Append appends item to Menu
func (m *Menu) Append(item *sessioninfo.SessionInfo) {
	m.items = append(m.items, item)
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
	if len(m.items) == 0 {
		return nil
	}
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

// Render renders the menu
func (m *Menu) Render() {
	var selectedStyle tcell.Style
	accentStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack.TrueColor()).Background(tcell.ColorWhite)
	m.screen.Clear()
	if len(m.items) > 0 {
		for n, sess := range m.items {
			if n == m.selected {
				selectedStyle = accentStyle
			} else {
				selectedStyle = tcell.StyleDefault
			}
			m.emitStr(0, n, selectedStyle, sess.String())
		}
	} else {
		w, h := m.screen.Size()
		m.emitStr(w/2, h/2, selectedStyle, "No sessions available!")
	}

	_, h := m.screen.Size()
	for _, error := range m.drainErrors() {
		m.errors.Push(error)
	}
	if len(m.errors.Content()) > 0 {
		m.emitStr(0, h-len(m.errors.Content())-1, tcell.StyleDefault, "Errors:")
		for n, error := range m.errors.Content() {
			m.emitStr(0, h-n-1, tcell.StyleDefault, error.Error())
		}
	}
	m.screen.Sync()
}

func (m *Menu) Init() error {
	if err := m.screen.Init(); err != nil {
		return err
	}
	return nil
}

func (m *Menu) sendError(err error) {
	select {
	case m.serverErrors <- err:
	default:
	}
}

func (m *Menu) drainErrors() []error {
	var errs []error
	for {
		select {
		case err := <-m.serverErrors:
			errs = append(errs, err)
		default:
			return errs
		}
	}
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
						current := m.Current()
						if current != nil {
							if err := m.client.SessionKill(current.ID()); err != nil {
								m.sendError(err)
							}

						}
					}
				}
				if evt.Key() == tcell.KeyDown {
					m.Down()
				}
				if evt.Key() == tcell.KeyEnter {
					current := m.Current()
					if current != nil {
						m.screen.Suspend()
						toremote := make(chan []byte)
						tolocal := make(chan []byte)
						go func() {
							if err := m.client.SessionConnect(current.ID(), toremote, tolocal); err != nil {
								close(toremote)
								m.sendError(fmt.Errorf("connection closed: %s", err))
							}
						}()
						channelTerm(tolocal, toremote)
						m.screen.Resume()
						m.Render()
					}
				}
			}
		case <-time.Tick(time.Second):
			m.Render()
		case evt := <-m.serverEvents:
			switch evt.Type {
			case pb.EventType_DELETE:
				sessID := uuid.Must(uuid.Parse(evt.Session.Id))
				if ok := m.DeleteByID(sessID); !ok {
					m.sendError(fmt.Errorf("no such session %s", sessID))
				}
				m.Render()
			case pb.EventType_ADD:
				m.Append(&sessioninfo.SessionInfo{
					SID:         uuid.Must(uuid.Parse(evt.Session.Id)),
					SInitTime:   evt.Session.InitTime.AsTime(),
					SRemoteAddr: evt.Session.RemoteAddress,
				})
				m.Render()
			default:
				m.sendError(fmt.Errorf("unexpected evt %v", evt))
			}
		}
	}
}

func (m *Menu) RunTillExit() {
	go m.client.ChanelEvents(m.serverEvents, m.serverErrors)
	go m.screen.ChannelEvents(m.screenEvents, m.quit)
	m.evloop()
	m.screen.Fini()
}
