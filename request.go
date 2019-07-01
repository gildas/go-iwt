package iwt

import (
	"math"
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gildas/go-logger"
)

type requestOptions struct {
	Method         string
	Path           string
	Headers        map[string]string
	Parameters     map[string]string
	Accept         string
	ContentType    string
	Payload        interface{}
	Content        []byte
	Authentication string
	DeliveryTag    string
	RequestID      string
}

// sendRequest sends an HTTP request to Box.com's API
func (client *Client) sendRequest(ctx context.Context, options *requestOptions, results interface{}) (reader io.ReadCloser, contentType string, err error) {
	if len(options.RequestID) == 0 {
		options.RequestID = uuid.Must(uuid.NewRandom()).String()
	}
	log := client.Logger.Scope("request").Record("reqid", options.RequestID).Child()

	// Building the request body
	reqContent, reqContentSize, err := client.buildReqContent(log, options)
	if err != nil {
		return
	}

	// Sanitizing the given options
	if len(options.Method) == 0 {
		if reqContentSize > 0 {
			options.Method = "POST"
		} else {
			options.Method = "GET"
		}
	}
	if len(options.Accept) == 0 {
		options.Accept = "application/json"
	}

	// Building a new HTTP request
	var endpoint *url.URL
	if strings.HasPrefix(options.Path, "http") {
		log.Debugf("Option Path is a full URL: %s", options.Path)
		endpoint, err = url.Parse(options.Path)
	} else {
		log.Debugf("Using: %s", client.CurrentAPIEndpoint().String())
		endpoint, err = url.Parse(client.CurrentAPIEndpoint().String() + options.Path)
	}
	if err != nil {
		return
	}
	req, err := http.NewRequest(options.Method, endpoint.String(), reqContent)
	if err != nil {
		return
	}

	// Setting request headers
	req.Header.Set("User-Agent", "GENESYS IWT Client "+VERSION)
	req.Header.Set("Accept",       options.Accept)
	req.Header.Set("X-Request-Id", options.RequestID)
	if len(options.ContentType) > 0 {
		req.Header.Set("Content-Type", options.ContentType)
	}
	if reqContentSize > 0 {
		req.Header.Set("Content-Length", strconv.Itoa(reqContentSize))
	}

	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	// Sending the request...
	log.Debugf("HTTP %s %s", req.Method, req.URL.String())
	log.Tracef("Request Headers: %#v", req.Header)
	httpclient := &http.Client{}
	if client.Proxy != nil {
		httpclient.Transport = &http.Transport{Proxy: http.ProxyURL(client.Proxy)}
	}
	start    := time.Now()
	res, err := httpclient.Do(req)
	duration := time.Since(start)
	if err != nil {
		log.Errorf("Failed to send request", err)
		return
	}
	defer res.Body.Close()

	// Reading the response body
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Failed to read response body", err)
		return
	}
	log.Debugf("Response %s in %s", res.Status, duration)
	log.Tracef("Response Headers: %#v", res.Header)
	log.Tracef("Response body (%d bytes): %s", len(resBody), string(resBody[:int(math.Min(1024,float64(len(resBody))))]))

	resRequestID := res.Header.Get("X-Request-Id")
	if resRequestID != options.RequestID {
		log = log.Record("duration", duration.Seconds()).Record("resreqid", resRequestID)
	}

	// Processing the response
	// TODO: Process redirections (3xx)
	if res.StatusCode == http.StatusFound {
		locationURL, err := res.Location()
		if err != nil {
			return nil, "", err
		}
		log.Infof("We should get stuff from %s", locationURL)
	}
	if res.StatusCode == 503 {
		return nil, "", StatusUnavailableService
	}
	if res.StatusCode >= 400 {
		//requestError := RequestError{}
		//if err = json.Unmarshal(resBody, &requestError); err == nil {
		//	return nil, requestError
		//}
		return nil, "", fmt.Errorf("%s", res.Status)
	}

	if results != nil {
		err = json.Unmarshal(resBody, results)
		if err != nil {
			log.Errorf("Failed to decode response", err)
			return
		}
	}
	return ioutil.NopCloser(bytes.NewReader(resBody)), res.Header.Get("Content-Type"), nil
}

// buildReqContent build the request body
// the ContentType can also be set as needed
func (client *Client) buildReqContent(log *logger.Logger, options *requestOptions) (body *bytes.Buffer, size int, err error) {
	body = bytes.NewBuffer(nil)

	if len(options.Parameters) > 0 {
		if options.ContentType == "application/x-www-form-urlencoded" {
			query := url.Values{}
			for param, value := range options.Parameters {
				query.Set(param, value)
			}
			encoded := query.Encode()
			body = bytes.NewBuffer([]byte(encoded))
			size = len(encoded)
		} else { // Create a multipart data form
			body = &bytes.Buffer{}

			writer := multipart.NewWriter(body)
			for param, value := range options.Parameters {
				if strings.HasPrefix(param, ">") {
					param = strings.TrimPrefix(param, ">")
					if len(value) == 0 {
						return nil, 0, fmt.Errorf("Empty value for field %s", param)
					}
					part, err := writer.CreateFormFile(param, value)
					if err != nil {
						return nil, 0, fmt.Errorf("Failed to create multipart for field %s, %s", param, err)
					}
					if len(options.Content) == 0 {
						return nil, 0, fmt.Errorf("Missing Content for Parameter %s", param)
					}
					written, err := io.Copy(part, bytes.NewReader(options.Content))
					if err != nil {
						return nil, 0, fmt.Errorf("Failed to write payload to multipart field %s, %s", param, err)
					}
					log.Tracef("Wrote %d bytes to field %s", written, param)
				} else {
					if err = writer.WriteField(param, value); err != nil {
						return nil, 0, fmt.Errorf("Failed to create field %s, %s", param, err)
					}
					log.Tracef("Added field %s = %s", param, value)
				}
			}
			if err := writer.Close(); err != nil {
				return nil, 0, fmt.Errorf("Failed to create multipart data, %s", err)
			}
			options.ContentType = writer.FormDataContentType()
		}
	} else if options.Payload != nil {
		// Create a JSON payload
		// TODO: Add other payload types like XML, etc
		payload, err := json.Marshal(options.Payload)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to encode payload into JSON, %s", err)
		}
		body = bytes.NewBuffer(payload)
		size = len(payload)
		if len(options.ContentType) == 0 {
			options.ContentType = "application/json"
		}
	} else if len(options.Content) > 0 {
		body = bytes.NewBuffer(options.Content)
		size = len(options.Content)
		if len(options.ContentType) == 0 {
			options.ContentType = "application/octet-stream"
		}
	}
	return
}