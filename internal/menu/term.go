package menu

import (
	"log"
	"reflect"

	"github.com/pkg/term"
)

func channelTerm(reader, writer chan []byte) {
	t, err := term.Open("/dev/tty")
	if err != nil {
		panic(err)
	}
	term.RawMode(t)
	quit := make(chan struct{})
	go func() {
		defer close(reader)
		for {
			select {
			case <-quit:
				return
			default:
			}
			bytes := make([]byte, 4)
			n, err := t.Read(bytes)
			if reflect.DeepEqual(bytes[:n], []byte{0x1b}) {
				return
			}
			if err != nil {
				log.Println(err)
				return
			}
			reader <- bytes
		}
	}()
	for {
		data, more := <-writer
		if !more {
			break
		}
		t.Restore()
		t.Write(data)
		term.RawMode(t)
	}
	t.Restore()
	t.Close()
}
