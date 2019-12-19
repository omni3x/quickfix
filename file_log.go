package quickfix

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/quickfixgo/quickfix/config"
)

type fileLog struct {
	eventLogger   *log.Logger
	messageLogger *log.Logger
}

func (l fileLog) OnIncoming(msg []byte) {
	msgStr := string(msg)
	msgTypeIdx := strings.Index(msgStr, "35=")
	if msgTypeIdx == -1 {
		// This should never happen
		l.messageLogger.Print(msgStr)
		return
	}

	msgTypeValueIdx := msgTypeIdx + 3
	msgType := msgStr[msgTypeValueIdx : msgTypeValueIdx+2]
	if msgType == "W" || msgType == "X" {
		return // don't save price data
	}
	l.messageLogger.Print(msgStr)
}

func (l fileLog) OnOutgoing(msg []byte) {
	msgStr := string(msg)
	msgTypeIdx := strings.Index(msgStr, "35=")
	if msgTypeIdx == -1 {
		// This should never happen
		l.messageLogger.Print(msgStr)
		return
	}

	msgTypeValueIdx := msgTypeIdx + 3
	msgType := msgStr[msgTypeValueIdx : msgTypeValueIdx+2]
	if msgType == "W" || msgType == "X" {
		return // don't save price data
	} else if msgType == "D" { // NewOrderSingle: API KEY (467), SECRET (2001), PASS (2002)
		msgStr = redactTags([]string{"467", "2001", "2002"}, msgStr)
	} else if msgType == "A" { // Logon: Password (554)
		msgStr = redactTags([]string{"554"}, msgStr)
	}
	l.messageLogger.Print(msgStr)
}

func redactTags(tags []string, msg string) string {
	redacted := msg
	for _, tag := range tags {
		tagIdx := strings.Index(msg, tag)
		if tagIdx == -1 {
			continue
		}
		replacePos := tagIdx + len(tag) + 1 // +1 is for the = sign after the tag
		delimIdx := strings.Index(msg[replacePos:], "\001")
		if delimIdx == -1 {
			continue // Should not happen
		}
		redacted = strings.Replace(redacted, msg[replacePos:delimIdx+1], "******", 1)
	}
	return redacted
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
