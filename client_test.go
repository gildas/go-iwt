package iwt_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-iwt"
	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/suite"
)

type IWTTestSuite struct {
	suite.Suite
	Name   string
	Start  time.Time
	Client *iwt.Client
	Logger *logger.Logger
}

func TestIWTSuite(t *testing.T) {
	suite.Run(t, new(IWTTestSuite))
}

func (suite *IWTTestSuite) TestCanInstantiate() {
	suite.Require().NotEmpty(suite.Client.APIEndpoints, "Missing API endpoints")
	suite.Require().NotNil(suite.Client.APIEndpoints[0], "Primary API endpoint should not be nil")
	suite.Require().NotEmpty(suite.Client.APIEndpoints[0].String(), "Primary API endpoint should not be empty")
	suite.Client.Logger.Infof("Primary Endpoint: %s", suite.Client.CurrentAPIEndpoint().String())
}

func (suite *IWTTestSuite) TestCanFetchServerConfiguration() {
	config, err := suite.Client.GetServerConfiguration()
	suite.Require().Nil(err, "Failed to fetch server configuration, Error: %s", err)
	suite.Require().NotNil(config, "Failed to fetch server configuration")
	suite.Assert().NotEmpty(config.Capabilities, "No capabilities")
	chatCaps, ok := config.Capabilities["chat"]
	suite.Require().True(ok, "Server has no chat capability")
	suite.Assert().Contains(chatCaps, "start")
	suite.Assert().Contains(chatCaps, "reconnect")
	suite.Assert().Contains(chatCaps, "poll")
	suite.Assert().Contains(chatCaps, "sendMessage")
	suite.Assert().Contains(chatCaps, "exit")
	suite.Client.Logger.Infof("Configuration: %#v", config)
}

func (suite *IWTTestSuite) TestCanQueryQueue() {
	queue, err := suite.Client.QueryQueue("Line", iwt.WorkgroupQueue)
	suite.Require().Nil(err, "Failed to query queue, Error: %s", err)
	suite.Client.Logger.Infof("Queue: %#v", queue)
}

func (suite *IWTTestSuite) TestFailsQueryUnknownQueue() {
	queue, err := suite.Client.QueryQueue("UnknownQueue", iwt.WorkgroupQueue)
	suite.Require().NotNil(err)
	suite.Assert().Equal("error.websvc.unknownEntity.invalidQueue", err.Error())
	suite.Client.Logger.Infof("Queue: %#v", queue)
}

func (suite *IWTTestSuite) TestCanStartAndStopChat() {
	chat, err := suite.Client.StartChat(iwt.StartChatOptions{
		Queue: &iwt.Queue{Type: iwt.WorkgroupQueue, Name: "Line"},
		Guest: iwt.Participant{Name: "UnitTest"},
	})
	suite.Require().Nil(err, "Failed to start a chat, Error: %s", err)
	suite.Require().NotNil(chat, "Chat is nil")
	suite.Assert().NotEmpty(chat.ID, "Chat Identifier is empty")
	suite.Client.Logger.Infof("Chat: %#v", chat)

	time.Sleep(5 * time.Second)
	err = chat.Stop()
	suite.Require().Nil(err, "Failed to stop a chat, Error: %s", err)
}

func (suite *IWTTestSuite) SetupSuite() {
	suite.Name = "IWT"
	suite.Logger = CreateLogger(fmt.Sprintf("test-%s.log", strings.ToLower(suite.Name)))

	primaryAPI, err := url.Parse(core.GetEnvAsString("PRIMARY", ""))
	suite.Require().Nil(err, "Failed to parse Primary PureConnect URL")

	suite.Client = iwt.NewClient(context.Background(), iwt.ClientOptions{
		PrimaryAPI: primaryAPI,
		Logger:     suite.Logger,
	})
	suite.Require().NotNil(suite.Client, "Failed to instantiate a new IWT Client")

	suite.Logger.Infof("Start %s %s", suite.Name, strings.Repeat("=", 80-7-len(suite.Name)))
}

func (suite *IWTTestSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaing")
	} else {
		suite.Logger.Infof("Cleaning all data from %s", suite.Name)
		//err := suite.Provider.PurgeAll(nil, suite.Logger)
		//suite.Nil(err, "Failed to clean data. %s", err)
	}
	suite.Logger.Infof("End %s %s", suite.Name, strings.Repeat("=", 80-5-len(suite.Name)))
}

func (suite *IWTTestSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Start %s %s", testName, strings.Repeat("-", 80-7-len(testName)))
	suite.Start = time.Now()
}

func (suite *IWTTestSuite) AfterTest(suiteName, testName string) {
	duration := time.Now().Sub(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("End %s %s", testName, strings.Repeat("-", 80-5-len(testName)))
}
