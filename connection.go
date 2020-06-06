package quickfix

import (
	"fmt"
	"io"
	"time"
)

func writeLoop(connection io.Writer, messageOut chan []byte, log Log) {
	for {
		msg, ok := <-messageOut
		if !ok {
			return
		}

		if _, err := connection.Write(msg); err != nil {
			log.OnEvent(err.Error())
		}
	}
}

func readLoop(parser *parser, msgIn chan fixIn) {
	defer close(msgIn)

	for {
		msg, err := parser.ReadMessage()
		if err != nil {
			return
		}
		fmt.Println("BEFORE CHANNEL WRITE DELTA: ", time.Since(parser.lastRead))
		select {
		case msgIn <- fixIn{msg, parser.lastRead}:
			// fmt.Println("AFTER CHANNEL WRITE DELTA: ", time.Since(parser.lastRead))
		default:
			fmt.Println("BLOCKED: ", time.Now())
		}
	}
}
