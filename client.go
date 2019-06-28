package iwt

import (
	"context"
	"net/url"
	"strings"

	"github.com/gildas/go-logger"
)

// Client is the IWT client to talk to PureConnect
type Client struct {
	APIEndpoints  []*url.URL       `json:"apiEndpoints"`
	EndPointIndex int             `json:"endpointIndex"`
	Proxy         *url.URL        `json:"proxy"`
	Language      string          `json:"language"`
	CACert        []byte          `json:"cacert"`
	Context       context.Context `json:"-"`
	Logger        *logger.Logger  `json:"-"`
}

// ClientOptions defines the options for instantiating a new IWT Client
// If you use https with the Primary/Failover API endpoint and they use a self-signed certificate, you must give the option CACert
type ClientOptions struct {
	PrimaryAPI  *url.URL       `json:"primary"`
	FailoverAPI *url.URL       `json:"failover"`
	CACert      []byte         `json:"cacert"`
	Proxy       *url.URL       `json:"proxy"`
	Language    string         `json:"language"`
	Logger      *logger.Logger `json:"-"`
}

// NewClient instantiates a new IWT Client
func NewClient(ctx context.Context, options ClientOptions) *Client {
	var err error

	// make sure we have a logger
	log := options.Logger
	if log == nil {
		log, err = logger.FromContext(ctx)
		if err != nil {
			log = logger.Create("IWT")
		}
	}

	client := &Client{
		APIEndpoints:  []*url.URL{},
		EndPointIndex: 0,
		Proxy:         options.Proxy,
		Language:      options.Language,
		CACert:        options.CACert,
		Context:       ctx,
		Logger:        log.Topic("iwt").Scope("iwt").Child(),
	}

	if options.PrimaryAPI == nil {
		options.PrimaryAPI, _ = url.Parse("https://localhost:3508")
	} else if !strings.HasSuffix(options.PrimaryAPI.Path, "/websvcs") {
		options.PrimaryAPI.Path = "/websvcs"
	}
	client.APIEndpoints = append(client.APIEndpoints, options.PrimaryAPI)

	if options.FailoverAPI != nil {
		if !strings.HasSuffix(options.FailoverAPI.Path, "/websvcs") {
			options.FailoverAPI.Path = "/websvcs"
		}
		client.APIEndpoints = append(client.APIEndpoints, options.FailoverAPI)
	}
	return client
}

// CurrentAPIEndpoint gives the current API Endpoint to use
func (client Client) CurrentAPIEndpoint() *url.URL {
	return client.APIEndpoints[client.EndPointIndex]
}

// NextAPIEndpoint switches to the next API endpoint (or back at the beginning)
func (client *Client) NextAPIEndpoint() *url.URL {
	if len(client.APIEndpoints) > 1 {
		client.EndPointIndex++
		if client.EndPointIndex >= len(client.APIEndpoints) {
			client.EndPointIndex = 0
		}
	}
	return client.APIEndpoints[client.EndPointIndex]
}