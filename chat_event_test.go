package iwt_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-iwt"
	"github.com/gildas/go-logger"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type ChatEventSuite struct {
	suite.Suite
	Name   string
	Start  time.Time
	Client *iwt.Client
	Logger *logger.Logger
}

func TestChatEventSuite(t *testing.T) {
	suite.Run(t, new(ChatEventSuite))
}

// *****************************************************************************
// Suite Tools

func (suite *ChatEventSuite) SetupSuite() {
	_ = godotenv.Load()
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:         fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:   true,
			SourceInfo:   true,
			FilterLevels: logger.NewLevelSet(logger.TRACE),
		},
	).Child("test", "test")

	primaryAPI, err := url.Parse(core.GetEnvAsString("PRIMARY", ""))
	suite.Require().Nil(err, "Failed to parse Primary PureConnect URL")

	suite.Client = iwt.NewClient(context.Background(), iwt.ClientOptions{
		PrimaryAPI: primaryAPI,
		Logger:     suite.Logger,
	})
	suite.Require().NotNil(suite.Client, "Failed to instantiate a new IWT Client")

	suite.Logger.Infof("Suite Start %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
}

func (suite *IWTTestSuite) ChatEventSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaing")
	} else {
		suite.Logger.Infof("Cleaning all data from %s", suite.Name)
		//err := suite.Provider.PurgeAll(nil, suite.Logger)
		//suite.Nil(err, "Failed to clean data. %s", err)
	}
	suite.Logger.Infof("Suite End %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *ChatEventSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *ChatEventSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}

func (suite *ChatEventSuite) LoadTestData(filename string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(".", "testdata", filename))
}

func (suite *ChatEventSuite) UnmarshalData(filename string, v interface{}) error {
	data, err := suite.LoadTestData(filename)
	suite.Logger.Infof("Loaded %s: %s", filename, string(data))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// *****************************************************************************

func (suite *ChatEventSuite) TestCanUnmarshalFileEvent() {
	payload, err := suite.LoadTestData("chat-event-file.json")
	suite.Require().Nilf(err, "Failed to load test data, Error: %s", err)

	event, err := iwt.UnmarshalChatEvent(payload)
	suite.Require().Nilf(err, "Failed to unmarshal json, Error: %s", err)
	suite.Require().NotNil(event, "Failed to unmarshal json, the result is nil")
	suite.Logger.Infof("Event: %+#v", event)

	actual, ok := event.(*iwt.FileEvent)
	suite.Require().True(ok, "Event is not of the proper type")
	suite.Require().NotNil(actual, "Actual Event should not be nil")
	suite.Assert().Equal("4cf50a4c-73bb-4e24-a6ff-d069dfcc6ceb", actual.Participant.ID)
	suite.Assert().Equal("Administrator", actual.Participant.Name)
	suite.Assert().Equal("Agent", actual.Participant.Type)
	suite.Assert().Equal("/websvcs/chat/getfile/4173f183-ed9d-4b0c-8a19-4899676ab206/46da9750794c465581c043128ad5e28c/-image.jpeg", actual.Path)
}

func (suite *ChatEventSuite) TestCanUnmarshalParticipantStateChangedEvent() {
	payload, err := suite.LoadTestData("chat-event-participantstatechanged.json")
	suite.Require().Nilf(err, "Failed to load test data, Error: %s", err)

	event, err := iwt.UnmarshalChatEvent(payload)
	suite.Require().Nilf(err, "Failed to unmarshal json, Error: %s", err)
	suite.Require().NotNil(event, "Failed to unmarshal json, the result is nil")
	suite.Logger.Infof("Event: %+#v", event)

	actual, ok := event.(*iwt.ParticipantStateChangedEvent)
	suite.Require().True(ok, "Event is not of the proper type")
	suite.Require().NotNil(actual, "Actual Event should not be nil")
	suite.Assert().Equal("4173f183-ed9d-4b0c-8a19-4899676ab206", actual.Participant.ID)
	suite.Assert().Equal("Bob Minion", actual.Participant.Name)
	suite.Assert().Equal("active", actual.Participant.State)
}

func (suite *ChatEventSuite) TestCanUnmarshalTextEvent() {
	payload, err := suite.LoadTestData("chat-event-text.json")
	suite.Require().Nilf(err, "Failed to load test data, Error: %s", err)

	event, err := iwt.UnmarshalChatEvent(payload)
	suite.Require().Nilf(err, "Failed to unmarshal json, Error: %s", err)
	suite.Require().NotNil(event, "Failed to unmarshal json, the result is nil")
	suite.Logger.Infof("Event: %+#v", event)

	actual, ok := event.(*iwt.TextEvent)
	suite.Require().True(ok, "Event is not of the proper type")
	suite.Require().NotNil(actual, "Actual Event should not be nil")
	suite.Assert().Equal("4cf50a4c-73bb-4e24-a6ff-d069dfcc6ceb", actual.Participant.ID)
	suite.Assert().Equal("Administrator", actual.Participant.Name)
	suite.Assert().Equal("Agent", actual.Participant.Type)
	suite.Assert().Equal("banana", actual.Text)
}

func (suite *ChatEventSuite) TestCanUnmarshalTypingIndicatorEvent() {
	payload, err := suite.LoadTestData("chat-event-typingindicator.json")
	suite.Require().Nilf(err, "Failed to load test data, Error: %s", err)

	event, err := iwt.UnmarshalChatEvent(payload)
	suite.Require().Nilf(err, "Failed to unmarshal json, Error: %s", err)
	suite.Require().NotNil(event, "Failed to unmarshal json, the result is nil")
	suite.Logger.Infof("Event: %+#v", event)

	actual, ok := event.(*iwt.TypingIndicatorEvent)
	suite.Require().True(ok, "Event is not of the proper type")
	suite.Require().NotNil(actual, "Actual Event should not be nil")
	suite.Assert().Equal("4cf50a4c-73bb-4e24-a6ff-d069dfcc6ceb", actual.Participant.ID)
}

func (suite *ChatEventSuite) TestCanUnmarshalURLEvent() {
	payload, err := suite.LoadTestData("chat-event-url.json")
	suite.Require().Nilf(err, "Failed to load test data, Error: %s", err)

	event, err := iwt.UnmarshalChatEvent(payload)
	suite.Require().Nilf(err, "Failed to unmarshal json, Error: %s", err)
	suite.Require().NotNil(event, "Failed to unmarshal json, the result is nil")
	suite.Logger.Infof("Event: %+#v", event)

	actual, ok := event.(*iwt.URLEvent)
	suite.Require().True(ok, "Event is not of the proper type")
	suite.Require().NotNil(actual, "Actual Event should not be nil")
	suite.Assert().Equal("4cf50a4c-73bb-4e24-a6ff-d069dfcc6ceb", actual.Participant.ID)
	suite.Assert().Equal("Administrator", actual.Participant.Name)
	suite.Assert().Equal("https://www.genesys.com", actual.URL.String())
}
