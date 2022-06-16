package menu

import (
	"log"
	"reflect"

	"github.com/pkg/term"
)

func ChannelTerm(reader chan<- []byte, writer <-chan []byte) {
	t, err := term.Open("/dev/tty")
	if err != nil {
		panic(err)
	}
	term.RawMode(t)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
			}
			bytes := make([]byte, 4)
			n, err := t.Read(bytes)
			if reflect.DeepEqual(bytes[:n], []byte("A")) {
				close(reader)
				return
			}
			if err != nil {
				log.Println(err)
				return
			}
			// t.Write(bytes)
			reader <- bytes
		}
	}()
	for {
		data, more := <-writer
		if !more {
			log.Printf("Writer closed")
			quit <- struct{}{}
			break
		}
		t.Write(data)
	}
	t.Restore()
	t.Close()

}
