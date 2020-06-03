package quickfix

import (
	"fmt"
	"io"
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
		select {
		case msgIn <- fixIn{msg, parser.lastRead}:
		default:
			fmt.Println("!!!!! [TIMING]  MSGIN IS BLOCKED")
		}
	}
}
