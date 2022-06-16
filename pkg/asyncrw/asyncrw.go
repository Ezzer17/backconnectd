package asyncrw

import (
	_errors "errors"
)

type AsyncSenderReciever interface {
	Send(data []byte) error
	Recieve() ([]byte, error)
}

func AsyncSendRecieve(a AsyncSenderReciever, send chan<- []byte, recieve <-chan []byte) error {
	comm := make(chan []byte)
	errors := make(chan error)
	go func() {
		for {
			data, err := a.Recieve()
			if err != nil {
				errors <- err
				return
			}
			comm <- data
		}
	}()
	for {
		select {
		case data := <-comm:
			send <- data
		case data, more := <-recieve:
			if !more {
				return _errors.New("recieve channel closed")
			}
			err := a.Send(data)
			if err != nil {
				return err
			}
		case err := <-errors:
			return err
		}
	}
}
