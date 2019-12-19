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

func TestGetMsgType(t *testing.T) {
	logon := []byte("8=FIXT.1.19=11835=A34=249=demo-1-taker52=20191218-23:15:42.24156=OmniexFeed98=0108=30553=demo-1-taker554=x%hvFtF9xjpE1137=910=054")
	msgType := "A"
	output := getMsgType(logon)
	if msgType != output {
		t.Errorf("Failed to get correct msgType\nReceived: %s\nExpected: %s", output, msgType)
	}
	newOrder := []byte("8=FIXT.1.19=108435=D34=349=demo-9-taker52=20190529-01:11:27.39656=Omniex11=01DC0J8NX4SVGKSFW2CW27AEZZ15=ETH38=5040=254=255=ETH/BTC59=160=20190529-01:11:27.39678=679=gdax80=1000467=45a8edecb146bfe598df1c3f7e12f5e7661=992001=idvt3ig47lp2002=offuwiL37puuoYZkAlsuaomdxyYDH4ffwt73szP5DkursNBHSQwZtL1jOQznkwHA2mT7C8yFyUtgRYEPYfhN+Q==79=gemini80=1000467=LBuO0bgmXsmhw7ealfxO661=992001=2002=birJVYMANk7ssxsBaNToDdF8bmp79=bitfinex80=1000467=ZxBCsodMcB0KGJuqcKmUEobYqs7tdRF6KMcHXN2dN7B661=992001=2002=6gEfBkCylm10T7QnThFBqjra7cwp7CXRBZgT3tm7xzm79=binance80=1000467=elMTbsb4VIRA9Aa2wfBRPkIvHtXQu05vFYdwykSLiyZ3W2p2Zvi6zbwESnQfFjwv661=992001=2002=BPWbHOFvDlLyb9tGuNkNRiFURzxsBkdBQMfb6BKKIlndGqKVcpvt0bSD76AEBmux79=bittrex80=1000467=cbe82075119f4984b0c450400c4f1727661=992001=2002=e666f82c6fd144cab269a975544bff2879=kraken80=1000467=j7fefNM3lDYv5knvHhDz5yliVNLd9TrpUMEKx/9GIpBGPtjdFXyt/RMm661=992001=2002=qtOInzpcl6CFgA6A3tjiA9b2TOOp3jVCIX4L+BJ4MQvNSJoljC5i2BNaDXtlpNqbHn8uDjBLPBucZO1LlOVutQ==126=20190529-01:11:27.827453=1448=demo-9-taker447=D452=3847=100210=167")
	msgType = "D"
	output = getMsgType(newOrder)
	if msgType != output {
		t.Errorf("Failed to get correct msgType\nReceived: %s\nExpected: %s", output, msgType)
	}
}

func TestRedactTags(t *testing.T) {
	logon := []byte("8=FIXT.1.19=11835=A34=249=demo-1-taker52=20191218-23:15:42.24156=OmniexFeed98=0108=30553=demo-1-taker554=x%hvFtF9xjpE1137=910=054")
	expectedLogon := []byte("8=FIXT.1.19=11835=A34=249=demo-1-taker52=20191218-23:15:42.24156=OmniexFeed98=0108=30553=demo-1-taker554=************1137=910=054")
	redactTags("554=", logon)
	if string(expectedLogon) != string(logon) {
		t.Errorf("Incorrect Logon 554= Redaction.\nReceived: %s\nExpected: %s", string(logon), string(expectedLogon))
	}

	newOrder := []byte("8=FIXT.1.19=108435=D34=349=demo-9-taker52=20190529-01:11:27.39656=Omniex11=01DC0J8NX4SVGKSFW2CW27AEZZ15=ETH38=5040=254=255=ETH/BTC59=160=20190529-01:11:27.39678=679=gdax80=1000467=45a8edecb146bfe598df1c3f7e12f5e7661=992001=idvt3ig47lp2002=offuwiL37puuoYZkAlsuaomdxyYDH4ffwt73szP5DkursNBHSQwZtL1jOQznkwHA2mT7C8yFyUtgRYEPYfhN+Q==79=gemini80=1000467=LBuO0bgmXsmhw7ealfxO661=992001=2002=birJVYMANk7ssxsBaNToDdF8bmp79=bitfinex80=1000467=ZxBCsodMcB0KGJuqcKmUEobYqs7tdRF6KMcHXN2dN7B661=992001=2002=6gEfBkCylm10T7QnThFBqjra7cwp7CXRBZgT3tm7xzm79=binance80=1000467=elMTbsb4VIRA9Aa2wfBRPkIvHtXQu05vFYdwykSLiyZ3W2p2Zvi6zbwESnQfFjwv661=992001=2002=BPWbHOFvDlLyb9tGuNkNRiFURzxsBkdBQMfb6BKKIlndGqKVcpvt0bSD76AEBmux79=bittrex80=1000467=cbe82075119f4984b0c450400c4f1727661=992001=2002=e666f82c6fd144cab269a975544bff2879=kraken80=1000467=j7fefNM3lDYv5knvHhDz5yliVNLd9TrpUMEKx/9GIpBGPtjdFXyt/RMm661=992001=2002=qtOInzpcl6CFgA6A3tjiA9b2TOOp3jVCIX4L+BJ4MQvNSJoljC5i2BNaDXtlpNqbHn8uDjBLPBucZO1LlOVutQ==126=20190529-01:11:27.827453=1448=demo-9-taker447=D452=3847=100210=167")
	expectedNewOrder := []byte("8=FIXT.1.19=108435=D34=349=demo-9-taker52=20190529-01:11:27.39656=Omniex11=01DC0J8NX4SVGKSFW2CW27AEZZ15=ETH38=5040=254=255=ETH/BTC59=160=20190529-01:11:27.39678=679=gdax80=1000467=********************************661=992001=idvt3ig47lp2002=offuwiL37puuoYZkAlsuaomdxyYDH4ffwt73szP5DkursNBHSQwZtL1jOQznkwHA2mT7C8yFyUtgRYEPYfhN+Q==79=gemini80=1000467=********************661=992001=2002=birJVYMANk7ssxsBaNToDdF8bmp79=bitfinex80=1000467=*******************************************661=992001=2002=6gEfBkCylm10T7QnThFBqjra7cwp7CXRBZgT3tm7xzm79=binance80=1000467=****************************************************************661=992001=2002=BPWbHOFvDlLyb9tGuNkNRiFURzxsBkdBQMfb6BKKIlndGqKVcpvt0bSD76AEBmux79=bittrex80=1000467=********************************661=992001=2002=e666f82c6fd144cab269a975544bff2879=kraken80=1000467=********************************************************661=992001=2002=qtOInzpcl6CFgA6A3tjiA9b2TOOp3jVCIX4L+BJ4MQvNSJoljC5i2BNaDXtlpNqbHn8uDjBLPBucZO1LlOVutQ==126=20190529-01:11:27.827453=1448=demo-9-taker447=D452=3847=100210=167")
	redactTags("467=", newOrder)
	if string(expectedNewOrder) != string(newOrder) {
		t.Errorf("Incorrect NewOrderSingle 467= Redaction\nReceived: %s\nExpected: %s", string(newOrder), string(expectedNewOrder))
	}

	redactTags("2001=", newOrder)
	expectedNewOrder = []byte("8=FIXT.1.19=108435=D34=349=demo-9-taker52=20190529-01:11:27.39656=Omniex11=01DC0J8NX4SVGKSFW2CW27AEZZ15=ETH38=5040=254=255=ETH/BTC59=160=20190529-01:11:27.39678=679=gdax80=1000467=********************************661=992001=***********2002=offuwiL37puuoYZkAlsuaomdxyYDH4ffwt73szP5DkursNBHSQwZtL1jOQznkwHA2mT7C8yFyUtgRYEPYfhN+Q==79=gemini80=1000467=********************661=992001=2002=birJVYMANk7ssxsBaNToDdF8bmp79=bitfinex80=1000467=*******************************************661=992001=2002=6gEfBkCylm10T7QnThFBqjra7cwp7CXRBZgT3tm7xzm79=binance80=1000467=****************************************************************661=992001=2002=BPWbHOFvDlLyb9tGuNkNRiFURzxsBkdBQMfb6BKKIlndGqKVcpvt0bSD76AEBmux79=bittrex80=1000467=********************************661=992001=2002=e666f82c6fd144cab269a975544bff2879=kraken80=1000467=********************************************************661=992001=2002=qtOInzpcl6CFgA6A3tjiA9b2TOOp3jVCIX4L+BJ4MQvNSJoljC5i2BNaDXtlpNqbHn8uDjBLPBucZO1LlOVutQ==126=20190529-01:11:27.827453=1448=demo-9-taker447=D452=3847=100210=167")
	if string(expectedNewOrder) != string(newOrder) {
		t.Errorf("Incorrect NewOrderSingle 2001= Redaction\nReceived: %s\nExpected: %s", string(newOrder), string(expectedNewOrder))
	}
}

func TestRedactWithMissingTag(t *testing.T) {
	logon := []byte("8=FIXT.1.19=11835=A34=249=demo-1-taker52=20191218-23:15:42.24156=OmniexFeed98=0108=30553=demo-1-taker554=x%hvFtF9xjpE1137=910=054")
	expectedLogon := []byte("8=FIXT.1.19=11835=A34=249=demo-1-taker52=20191218-23:15:42.24156=OmniexFeed98=0108=30553=demo-1-taker554=x%hvFtF9xjpE1137=910=054")
	redactTags("2001=", logon)
	if string(expectedLogon) != string(logon) {
		t.Errorf("Incorrect Logon 554= Redaction.\nReceived: %s\nExpected: %s", string(logon), string(expectedLogon))
	}
}
