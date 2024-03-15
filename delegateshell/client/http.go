package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	pb "google.golang.org/protobuf/proto"

	"github.com/cenkalti/backoff/v4"
	"github.com/drone/go-task/delegateshell/delegate"
	"github.com/sirupsen/logrus"

	"github.com/wings-software/dlite/logger"
)

const (
	registerEndpoint         = "/api/agent/delegates/register?accountId=%s"
	heartbeatEndpoint        = "/api/agent/delegates/heartbeat-with-polling?accountId=%s"
	taskStatusEndpoint       = "/api/agent/v2/tasks/%s/delegates/%s?accountId=%s"
	runnerEventsPollEndpoint = "/api/agent/delegates/%s/runner-events?accountId=%s"
	executionPayloadEndpoint = "/api/executions/%s/payload?delegateId=%s&accountId=%s&delegateInstanceId=%s"
)

var (
	registerTimeout      = 30 * time.Second
	taskEventsTimeout    = 60 * time.Second
	sendStatusRetryTimes = 5
)

// defaultClient is the default http.Client.
var defaultClient = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// New returns a new client.
func New(endpoint, id, secret string, skipverify bool, additionalCertsDir string) *HTTPClient {
	return getClient(endpoint, id, "", delegate.NewTokenCache(id, secret), skipverify, additionalCertsDir)
}

func NewFromToken(endpoint, id, token string, skipverify bool, additionalCertsDir string) *HTTPClient {
	return getClient(endpoint, id, token, nil, skipverify, additionalCertsDir)
}

func getClient(endpoint, id, token string, cache *delegate.TokenCache, skipverify bool, additionalCertsDir string) *HTTPClient {
	log := logrus.New()
	c := &HTTPClient{
		Logger:            log,
		Endpoint:          endpoint,
		SkipVerify:        skipverify,
		AccountID:         id,
		Client:            defaultClient,
		AccountTokenCache: cache,
		Token:             token,
	}
	if skipverify {
		c.Client = &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipverify, //nolint:gosec
				},
			},
		}
	} else if additionalCertsDir != "" {
		// If additional certs are specified, we append them to the existing cert chain

		// Use the system certs if possible
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		log.Infof("additional certs dir to allow: %s\n", additionalCertsDir)

		files, err := os.ReadDir(additionalCertsDir)
		if err != nil {
			log.Errorf("could not read directory %s, error: %s", additionalCertsDir, err)
			c.Client = clientWithRootCAs(skipverify, rootCAs)
			return c
		}

		// Go through all certs in this directory and add them to the global certs
		for _, f := range files {
			path := filepath.Join(additionalCertsDir, f.Name())
			log.Infof("trying to add certs at: %s to root certs\n", path)
			// Create TLS config using cert PEM
			rootPem, err := os.ReadFile(path)
			if err != nil {
				log.Errorf("could not read certificate file (%s), error: %s", path, err.Error())
				continue
			}
			// Append certs to the global certs
			ok := rootCAs.AppendCertsFromPEM(rootPem)
			if !ok {
				log.Errorf("error adding cert (%s) to pool, please check format of the certs provided.", path)
				continue
			}
			log.Infof("successfully added cert at: %s to root certs", path)
		}
		c.Client = clientWithRootCAs(skipverify, rootCAs)
	}
	return c
}

func clientWithRootCAs(skipverify bool, rootCAs *x509.CertPool) *http.Client {
	// Create the HTTP Client with certs
	config := &tls.Config{
		//nolint:gosec
		InsecureSkipVerify: skipverify,
	}
	if rootCAs != nil {
		config.RootCAs = rootCAs
	}
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: config,
		},
	}
}

// An HTTPClient manages communication with the runner API.
type HTTPClient struct {
	Client            *http.Client
	Logger            logger.Logger
	Endpoint          string
	AccountID         string
	AccountTokenCache *delegate.TokenCache
	SkipVerify        bool
	Token             string
}

// Register registers the runner with the manager
func (p *HTTPClient) Register(ctx context.Context, r *RegisterRequest) (*RegisterResponse, error) {
	req := r
	resp := &RegisterResponse{}
	path := fmt.Sprintf(registerEndpoint, p.AccountID)
	_, err := p.retry(ctx, path, "POST", req, resp, createBackoff(ctx, registerTimeout), true) //nolint: bodyclose
	return resp, err
}

// Heartbeat sends a periodic heartbeat to the server
func (p *HTTPClient) Heartbeat(ctx context.Context, r *RegisterRequest) error {
	req := r
	path := fmt.Sprintf(heartbeatEndpoint, p.AccountID)
	_, err := p.doJson(ctx, path, "POST", req, nil)
	return err
}

// GetRunnerEvents gets a list of events which can be executed on this runner
func (p *HTTPClient) GetRunnerEvents(ctx context.Context, id string) (*RunnerEventsResponse, error) {
	path := fmt.Sprintf(runnerEventsPollEndpoint, id, p.AccountID)
	events := &RunnerEventsResponse{}
	_, err := p.doJson(ctx, path, "GET", nil, events)
	return events, err
}

// Acquire tries to acquire a specific task
func (p *HTTPClient) GetExecutionPayload(ctx context.Context, delegateID, taskID string) (*RunnerAcquiredTasks, error) {
	path := fmt.Sprintf(executionPayloadEndpoint, taskID, delegateID, p.AccountID, delegateID)
	payload := &RunnerAcquiredTasks{}
	_, err := p.doJson(ctx, path, "GET", nil, payload)
	if err != nil {
		logrus.WithError(err).Error("Error making http call")
	}
	return payload, err
}

// SendStatus updates the status of a task
func (p *HTTPClient) SendStatus(ctx context.Context, delegateID, taskID string, r *TaskResponse) error {
	path := fmt.Sprintf(taskStatusEndpoint, taskID, delegateID, p.AccountID)
	logrus.Info("response: ", path)
	req := r
	retryNumber := 0
	var err error
	for retryNumber < sendStatusRetryTimes {
		_, err = p.retry(ctx, path, "POST", req, nil, createBackoff(ctx, taskEventsTimeout), true) //nolint: bodyclose
		if err == nil {
			return nil
		}
		retryNumber++
	}
	return err
}

func (p *HTTPClient) retry(ctx context.Context, path, method string, in, out interface{}, b backoff.BackOffContext, ignoreStatusCode bool) (*http.Response, error) { //nolint: unparam
	for {
		res, err := p.doJson(ctx, path, method, in, out)
		// do not retry on Canceled or DeadlineExceeded
		if ctxErr := ctx.Err(); ctxErr != nil {
			p.logger().Errorf("http: context canceled")
			return res, ctxErr
		}

		duration := b.NextBackOff()

		if res != nil {
			// Check the response code. We retry on 500-range
			// responses to allow the server time to recover, as
			// 500's are typically not permanent errors and may
			// relate to outages on the server side.
			if (ignoreStatusCode && err != nil) || res.StatusCode > 501 {
				p.logger().Errorf("http: server error: re-connect and re-try: %s", err)
				if duration == backoff.Stop {
					p.logger().Errorf("max retry limit reached, task status won't be updated")
					return nil, err
				}
				time.Sleep(duration)
				continue
			}
		} else if err != nil {
			p.logger().Errorf("http: request error: %s", err)
			if duration == backoff.Stop {
				p.logger().Errorf("max retry limit reached, task status won't be updated")
				return nil, err
			}
			time.Sleep(duration)
			continue
		}
		return res, err
	}
}

func (p *HTTPClient) doJson(ctx context.Context, path, method string, in, out interface{}) (*http.Response, error) {
	var buf = &bytes.Buffer{}
	// marshal the input payload into json format and copy
	// to an io.ReadCloser.
	if in != nil {
		if err := json.NewEncoder(buf).Encode(in); err != nil {
			p.logger().Errorf("could not encode input payload: %s", err)
		}
	}
	res, body, err := p.do(ctx, path, method, buf)
	if err != nil {
		return res, err
	}
	if nil == out {
		return res, nil
	}
	if jsonErr := json.Unmarshal(body, out); jsonErr != nil {
		return res, jsonErr
	}

	return res, nil
}

func (p *HTTPClient) doProto(ctx context.Context, path, method string, in pb.Message, out pb.Message) (*http.Response, error) {
	// marshal the input payload into proto format and copy
	// to an io.ReadCloser.
	input, err := pb.Marshal(in)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(input)
	res, body, err := p.do(ctx, path, method, buf)
	if err != nil {
		return res, err
	}
	if nil == out {
		return res, nil
	}
	if protoErr := pb.Unmarshal(body, out); protoErr != nil {
		return res, protoErr
	}

	return res, nil
}

// do is a helper function that posts a signed http request with
// the input encoded and response decoded from json.
func (p *HTTPClient) do(ctx context.Context, path, method string, in *bytes.Buffer) (*http.Response, []byte, error) {
	endpoint := p.Endpoint + path
	req, err := http.NewRequest(method, endpoint, in)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)

	// the request should include the secret shared between
	// the agent and server for authorization.
	token := ""
	if p.Token != "" {
		token = p.Token
	} else {
		token, err = p.AccountTokenCache.Get()
		if err != nil {
			p.logger().Errorf("could not generate account token: %s", err)
			return nil, nil, err
		}
	}
	req.Header.Add("Authorization", "Delegate "+token)
	req.Header.Add("Content-Type", "application/json")
	res, err := p.Client.Do(req)
	if res != nil {
		defer func() {
			// drain the response body so we can reuse
			// this connection.
			if _, err = io.Copy(io.Discard, io.LimitReader(res.Body, 4096)); err != nil {
				p.logger().Errorf("could not drain response body: %s", err)
			}
			res.Body.Close()
		}()
	}
	if err != nil {
		return res, nil, err
	}

	// if the response body return no content we exit
	// immediately. We do not read or unmarshal the response
	// and we do not return an error.
	if res.StatusCode == 204 {
		return res, nil, nil
	}

	// else read the response body into a byte slice.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil, err
	}

	if res.StatusCode > 299 {
		// if the response body includes an error message
		// we should return the error string.
		if len(body) != 0 {
			return res, body, errors.New(
				string(body),
			)
		}
		// if the response body is empty we should return
		// the default status code text.
		return res, body, errors.New(
			http.StatusText(res.StatusCode),
		)
	}
	return res, body, nil
}

// logger is a helper function that returns the default logger
// if a custom logger is not defined.
func (p *HTTPClient) logger() logger.Logger {
	if p.Logger == nil {
		return logger.Discard()
	}
	return p.Logger
}

func createBackoff(ctx context.Context, maxElapsedTime time.Duration) backoff.BackOffContext {
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = maxElapsedTime
	return backoff.WithContext(exp, ctx)
}
