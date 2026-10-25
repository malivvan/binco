// Command stream receives messages of different types over one connection, skipping the ones it does not need and stopping at an unknown type.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/malivvan/binco"
	"github.com/malivvan/binco/wirefmt"
	wtypeid "github.com/malivvan/binco/wirefmt/typeid"
)

func main() {
	fmt.Println("1. receiving a stream of typed messages")
	stats, err := transfer(sampleMessages())
	if err != nil {
		panic(err)
	}
	fmt.Printf("   received %d messages, skipped %d, consumed %d bytes\n",
		stats.received, stats.skipped, stats.bytes)

	fmt.Println("\n2. an unknown message type stops the reader")
	stats, err = transfer([]typedMessage{unknownMessage{}})
	if unexpected, ok := errors.AsType[binco.UnexpectedDTMError](err); ok {
		fmt.Printf("   stopped with %v (DTM %d)\n", err, unexpected.DTM())
	} else {
		fmt.Printf("   stopped with %v\n", err)
	}
	fmt.Printf("   %d messages were received before the unknown one\n", stats.received)
}

func transfer(messages []typedMessage) (stats, error) {
	client, server := net.Pipe()

	sendErr := make(chan error, 1)
	go func() {
		defer func() { _ = client.Close() }()
		w := bufio.NewWriter(client)
		for _, m := range messages {
			if _, err := m.writeTo(w); err != nil {
				sendErr <- err
				return
			}
			if err := w.Flush(); err != nil {
				sendErr <- err
				return
			}
		}
		sendErr <- nil
	}()

	stats, err := receive(bufio.NewReader(server))

	_ = server.Close()
	if sendErr := <-sendErr; sendErr != nil && err == nil {
		err = sendErr
	}
	return stats, err
}

type stats struct {
	received int
	skipped  int
	bytes    int
}

func receive(r wirefmt.Reader) (stats, error) {
	var st stats
	for {

		dtm, n, err := wtypeid.DTMSer.Unmarshal(r)
		if errors.Is(err, io.EOF) {
			return st, nil
		}
		if err != nil {
			return st, err
		}
		st.bytes += n

		switch dtm {
		case LoginDTM:
			login, n, err := LoginTyped.UnmarshalData(r)
			if err != nil {
				return st, err
			}
			st.received++
			st.bytes += n
			fmt.Printf("   login:  %q joined\n", login.User)

		case ChatDTM:
			chat, n, err := ChatTyped.UnmarshalData(r)
			if err != nil {
				return st, err
			}
			st.received++
			st.bytes += n
			fmt.Printf("   chat:   %s says %q at %s\n", chat.From, chat.Text,
				chat.At.Format(time.RFC3339))

		case UploadDTM:

			n, err := UploadTyped.SkipData(r)
			if err != nil {
				return st, err
			}
			st.skipped++
			st.bytes += n
			fmt.Printf("   upload: skipped %d bytes\n", n)

		default:

			return st, fmt.Errorf("unknown message type: %w", binco.NewUnexpectedDTMError(dtm))
		}
	}
}

func sampleMessages() []typedMessage {
	at := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	return []typedMessage{
		Login{User: "alice"},
		Chat{From: "alice", Text: "hello everyone", At: at},
		Upload{Name: "report.pdf", Data: []byte{1, 2, 3, 4}},
		Chat{From: "bob", Text: "thanks!", At: at.Add(time.Minute)},
	}
}

type unknownMessage struct{}

func (unknownMessage) writeTo(w wirefmt.Writer) (n int, err error) {

	if n, err = wtypeid.DTMSer.Marshal(binco.DTM(99), w); err != nil {
		return
	}
	var n1 int
	n1, err = w.Write([]byte{0xDE, 0xAD})
	return n + n1, err
}
