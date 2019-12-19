package quickfix

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/quickfixgo/quickfix/config"
)

type fileLog struct {
	eventLogger   *log.Logger
	messageLogger *log.Logger
}

func (l fileLog) OnIncoming(msg []byte) {
	msgType := getMsgType(msg)
	if msgType == "W" || msgType == "X" {
		return // don't save price data
	}
	l.messageLogger.Print(string(msg))
}

func (l fileLog) OnOutgoing(msg []byte) {
	msgType := getMsgType(msg)
	if msgType == "W" || msgType == "X" {
		return // don't save price data
	} else if msgType == "D" { // NewOrderSingle: API KEY (467), SECRET (2001), PASS (2002)
		redactTags("467=", msg)
		redactTags("2001=", msg)
		redactTags("2002=", msg)
	} else if msgType == "A" { // Logon: Password (554)
		redactTags("554=", msg)
	}
	l.messageLogger.Print(string(msg))
}

func getMsgType(msg []byte) string {
	var msgType string
	for i, c := range msg {
		if c == '3' && i+3 < len(msg) {
			if msg[i+1] == '5' && msg[i+2] == '=' {
				v := msg[i+3]
				for v != 1 {
					msgType += string(v)
					i++
					break
				}
			}
		}
	}
	return msgType
}

func redactTags(tag string, msg []byte) {
	tagLen := len(tag)
	for i, c := range msg {
		if tag[0] == c && i+tagLen < len(msg) {
			match := true
			for j, t := range tag[1:] {
				if byte(t) != msg[i+j+1] {
					match = false
					break
				}
			}
			if match {
				vIdx := i + tagLen
				v := msg[vIdx]
				for v != 1 { // 1 is the delimiter ascii value
					msg[vIdx] = '*'
					vIdx++
					v = msg[vIdx]
				}
			}
		}
	}
}

func (l fileLog) OnEvent(msg string) {
	l.eventLogger.Print(msg)
}

func (l fileLog) OnEventf(format string, v ...interface{}) {
	l.eventLogger.Printf(format, v...)
}

type fileLogFactory struct {
	globalLogPath   string
	sessionLogPaths map[SessionID]string
}

//NewFileLogFactory creates an instance of LogFactory that writes messages and events to file.
//The location of global and session log files is configured via FileLogPath.
func NewFileLogFactory(settings *Settings) (LogFactory, error) {
	logFactory := fileLogFactory{}

	var err error
	if logFactory.globalLogPath, err = settings.GlobalSettings().Setting(config.FileLogPath); err != nil {
		return logFactory, err
	}

	logFactory.sessionLogPaths = make(map[SessionID]string)

	for sid, sessionSettings := range settings.SessionSettings() {
		logPath, err := sessionSettings.Setting(config.FileLogPath)
		if err != nil {
			return logFactory, err
		}
		logFactory.sessionLogPaths[sid] = logPath
	}

	return logFactory, nil
}

func newFileLog(prefix string, logPath string) (fileLog, error) {
	l := fileLog{}

	eventLogName := path.Join(logPath, prefix+".event.current.log")
	messageLogName := path.Join(logPath, prefix+".messages.current.log")

	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		return l, err
	}

	fileFlags := os.O_RDWR | os.O_CREATE | os.O_APPEND
	eventFile, err := os.OpenFile(eventLogName, fileFlags, os.ModePerm)
	if err != nil {
		return l, err
	}

	messageFile, err := os.OpenFile(messageLogName, fileFlags, os.ModePerm)
	if err != nil {
		return l, err
	}

	logFlag := log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC
	l.eventLogger = log.New(eventFile, "", logFlag)
	l.messageLogger = log.New(messageFile, "", logFlag)

	return l, nil
}

func (f fileLogFactory) Create() (Log, error) {
	return newFileLog("GLOBAL", f.globalLogPath)
}

func (f fileLogFactory) CreateSessionLog(sessionID SessionID) (Log, error) {
	logPath, ok := f.sessionLogPaths[sessionID]

	if !ok {
		return nil, fmt.Errorf("logger not defined for %v", sessionID)
	}

	prefix := sessionIDFilenamePrefix(sessionID)
	return newFileLog(prefix, logPath)
}
