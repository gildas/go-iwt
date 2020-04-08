package iwt

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gildas/go-logger"
	"github.com/gildas/go-request"
)

// Client is the IWT client to talk to PureConnect
type Client struct {
	APIEndpoints  []*url.URL      `json:"apiEndpoints"`
	EndPointIndex int             `json:"endpointIndex"`
	Proxy         *url.URL        `json:"proxy"`
	Language      string          `json:"language"`
	CACert        []byte          `json:"cacert"`
	Context       context.Context `json:"-"`
	Logger        *logger.Logger  `json:"-"`
}

// ClientOptions defines the options for instantiating a new IWT Client
// If you use https with the Primary/Backup API endpoint and they use a self-signed certificate, you must give the option CACert
type ClientOptions struct {
	PrimaryAPI *url.URL       `json:"primary"`
	BackupAPI  *url.URL       `json:"backup"`
	CACert     []byte         `json:"cacert"`
	Proxy      *url.URL       `json:"proxy"`
	Language   string         `json:"language"`
	Logger     *logger.Logger `json:"-"`
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
		Logger:        log.Child("iwt", "iwt"),
	}

	if options.PrimaryAPI == nil {
		options.PrimaryAPI, _ = url.Parse("https://localhost:3508")
	} else if !strings.HasSuffix(options.PrimaryAPI.Path, "/websvcs") {
		options.PrimaryAPI.Path = "/websvcs"
	}
	client.APIEndpoints = append(client.APIEndpoints, options.PrimaryAPI)

	if options.BackupAPI != nil {
		if !strings.HasSuffix(options.BackupAPI.Path, "/websvcs") {
			options.BackupAPI.Path = "/websvcs"
		}
		client.APIEndpoints = append(client.APIEndpoints, options.BackupAPI)
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

// URLWithPath gets a full URL from a given path
func (client *Client) URLWithPath(path string) *url.URL {
	if !strings.HasPrefix(path, "http") {
		path = client.CurrentAPIEndpoint().String() + path
	}
	endpoint, err := url.Parse(path)
	if err != nil {
		client.Logger.Errorf("Invalid URL request for %s", path, err)
		return nil
	}
	return endpoint
}

func (client *Client) post(path string, payload, results interface{}) (*request.ContentReader, error) {
	return request.Send(&request.Options{
		Context:   client.Context,
		Method:    http.MethodPost,
		URL:       client.URLWithPath(path),
		UserAgent: "GENESYS IWT Client " + VERSION,
		Payload:   payload,
		Logger:    client.Logger,
	}, results)
}

func (client *Client) get(path string, results interface{}) (*request.ContentReader, error) {
	return request.Send(&request.Options{
		Context:   client.Context,
		URL:       client.URLWithPath(path),
		UserAgent: "GENESYS IWT Client " + VERSION,
		Logger:    client.Logger,
	}, results)
}
