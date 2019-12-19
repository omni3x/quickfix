package quickfix

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
)

func TestFileLog_NewFileLogFactory(t *testing.T) {

	factory, err := NewFileLogFactory(NewSettings())

	if err == nil {
		t.Error("Should expect error when settings have no file log path")
	}

	cfg := `
# default settings for sessions
[DEFAULT]
ConnectionType=initiator
ReconnectInterval=60
SenderCompID=TW
FileLogPath=.

# session definition
[SESSION]
BeginString=FIX.4.1
TargetCompID=ARCA
FileLogPath=mydir

[SESSION]
BeginString=FIX.4.1
TargetCompID=ARCA
SessionQualifier=BS
`
	stringReader := strings.NewReader(cfg)
	settings, _ := ParseSettings(stringReader)

	factory, err = NewFileLogFactory(settings)

	if err != nil {
		t.Error("Did not expect error", err)
	}

	if factory == nil {
		t.Error("Should have returned factory")
	}
}

type fileLogHelper struct {
	LogPath string
	Prefix  string
	Log     Log
}

func newFileLogHelper(t *testing.T) *fileLogHelper {
	prefix := "myprefix"
	logPath := path.Join(os.TempDir(), fmt.Sprintf("TestLogStore-%d", os.Getpid()))

	log, err := newFileLog(prefix, logPath)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	return &fileLogHelper{
		LogPath: logPath,
		Prefix:  prefix,
		Log:     log,
	}
}

func TestNewFileLog(t *testing.T) {
	helper := newFileLogHelper(t)

	tests := []struct {
		expectedPath string
	}{
		{path.Join(helper.LogPath, fmt.Sprintf("%v.messages.current.log", helper.Prefix))},
		{path.Join(helper.LogPath, fmt.Sprintf("%v.event.current.log", helper.Prefix))},
	}

	for _, test := range tests {
		if _, err := os.Stat(test.expectedPath); os.IsNotExist(err) {
			t.Errorf("%v does not exist", test.expectedPath)
		}
	}
}

func TestFileLog_Append(t *testing.T) {
	helper := newFileLogHelper(t)

	messageLogFile, err := os.Open(path.Join(helper.LogPath, fmt.Sprintf("%v.messages.current.log", helper.Prefix)))
	if err != nil {
		t.Error("Unexpected error", err)
	}
	defer messageLogFile.Close()

	eventLogFile, err := os.Open(path.Join(helper.LogPath, fmt.Sprintf("%v.event.current.log", helper.Prefix)))
	if err != nil {
		t.Error("Unexpected error", err)
	}
	defer eventLogFile.Close()

	messageScanner := bufio.NewScanner(messageLogFile)
	eventScanner := bufio.NewScanner(eventLogFile)

	helper.Log.OnIncoming([]byte("incoming"))
	if !messageScanner.Scan() {
		t.Error("Unexpected EOF")
	}

	helper.Log.OnEvent("Event")
	if !eventScanner.Scan() {
		t.Error("Unexpected EOF")
	}

	newHelper := newFileLogHelper(t)
	newHelper.Log.OnIncoming([]byte("incoming"))
	if !messageScanner.Scan() {
		t.Error("Unexpected EOF")
	}

	newHelper.Log.OnEvent("Event")
	if !eventScanner.Scan() {
		t.Error("Unexpected EOF")
	}
}

func TestRedactTags(t *testing.T) {
	logon := "8=FIXT.1.19=11835=A34=249=demo-1-taker52=20191218-23:15:42.24156=OmniexFeed98=0108=30553=demo-1-taker554=x%hvFtF9xjpE1137=910=054"
	expectedLogon := "8=FIXT.1.19=11835=A34=249=demo-1-taker52=20191218-23:15:42.24156=OmniexFeed98=0108=30553=demo-1-taker554=******1137=910=054"
	redactedLogon := redactTags([]string{"554"}, logon)
	if expectedLogon != redactedLogon {
		t.Errorf("Incorrected Redaction. Received: %s. Expected: %s", redactedLogon, expectedLogon)
	}
}
