package gocardless

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

var _ = query.Values
var _ = bytes.NewBuffer
var _ = json.NewDecoder
var _ = errors.New

// WebhookService manages webhooks
type WebhookService struct {
	endpoint string
	token    string
	client   *http.Client
}

// Webhook model
type Webhook struct {
	CreatedAt                       string                 `url:"created_at,omitempty" json:"created_at,omitempty"`
	Id                              string                 `url:"id,omitempty" json:"id,omitempty"`
	IsTest                          bool                   `url:"is_test,omitempty" json:"is_test,omitempty"`
	RequestBody                     string                 `url:"request_body,omitempty" json:"request_body,omitempty"`
	RequestHeaders                  map[string]interface{} `url:"request_headers,omitempty" json:"request_headers,omitempty"`
	ResponseBody                    string                 `url:"response_body,omitempty" json:"response_body,omitempty"`
	ResponseBodyTruncated           bool                   `url:"response_body_truncated,omitempty" json:"response_body_truncated,omitempty"`
	ResponseCode                    int                    `url:"response_code,omitempty" json:"response_code,omitempty"`
	ResponseHeaders                 map[string]interface{} `url:"response_headers,omitempty" json:"response_headers,omitempty"`
	ResponseHeadersContentTruncated bool                   `url:"response_headers_content_truncated,omitempty" json:"response_headers_content_truncated,omitempty"`
	ResponseHeadersCountTruncated   bool                   `url:"response_headers_count_truncated,omitempty" json:"response_headers_count_truncated,omitempty"`
	Successful                      bool                   `url:"successful,omitempty" json:"successful,omitempty"`
	Url                             string                 `url:"url,omitempty" json:"url,omitempty"`
}

// WebhookListParams parameters
type WebhookListParams struct {
	After     string `url:"after,omitempty" json:"after,omitempty"`
	Before    string `url:"before,omitempty" json:"before,omitempty"`
	CreatedAt struct {
		Gt  string `url:"gt,omitempty" json:"gt,omitempty"`
		Gte string `url:"gte,omitempty" json:"gte,omitempty"`
		Lt  string `url:"lt,omitempty" json:"lt,omitempty"`
		Lte string `url:"lte,omitempty" json:"lte,omitempty"`
	} `url:"created_at,omitempty" json:"created_at,omitempty"`
	IsTest     bool `url:"is_test,omitempty" json:"is_test,omitempty"`
	Limit      int  `url:"limit,omitempty" json:"limit,omitempty"`
	Successful bool `url:"successful,omitempty" json:"successful,omitempty"`
}

// WebhookListResult response including pagination metadata
type WebhookListResult struct {
	Webhooks []Webhook `json:"webhooks"`
	Meta     struct {
		Cursors struct {
			After  string `url:"after,omitempty" json:"after,omitempty"`
			Before string `url:"before,omitempty" json:"before,omitempty"`
		} `url:"cursors,omitempty" json:"cursors,omitempty"`
		Limit int `url:"limit,omitempty" json:"limit,omitempty"`
	} `json:"meta"`
}

// List
// Returns a [cursor-paginated](#api-usage-cursor-pagination) list of your
// webhooks.
func (s *WebhookService) List(ctx context.Context, p WebhookListParams, opts ...RequestOption) (*WebhookListResult, error) {
	uri, err := url.Parse(fmt.Sprintf(s.endpoint + "/webhooks"))
	if err != nil {
		return nil, err
	}

	o := &requestOptions{
		retries: 3,
	}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	var body io.Reader

	v, err := query.Values(p)
	if err != nil {
		return nil, err
	}
	uri.RawQuery = v.Encode()

	req, err := http.NewRequest("GET", uri.String(), body)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+s.token)

	req.Header.Set("GoCardless-Version", "2015-07-06")

	req.Header.Set("GoCardless-Client-Library", "<no value>")
	req.Header.Set("GoCardless-Client-Version", "1.0.0")
	req.Header.Set("User-Agent", userAgent)

	for key, value := range o.headers {
		req.Header.Set(key, value)
	}

	client := s.client
	if client == nil {
		client = http.DefaultClient
	}

	var result struct {
		Err *APIError `json:"error"`
		*WebhookListResult
	}

	err = try(o.retries, func() error {
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		err = responseErr(res)
		if err != nil {
			return err
		}

		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return err
		}

		if result.Err != nil {
			return result.Err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if result.WebhookListResult == nil {
		return nil, errors.New("missing result")
	}

	return result.WebhookListResult, nil
}

type WebhookListPagingIterator struct {
	cursor   string
	response *WebhookListResult
	params   WebhookListParams
	service  *WebhookService
}

func (c *WebhookListPagingIterator) Next() bool {
	if c.cursor == "" && c.response != nil {
		return false
	}

	return true
}

func (c *WebhookListPagingIterator) Value(ctx context.Context) (*WebhookListResult, error) {
	if !c.Next() {
		return c.response, nil
	}

	s := c.service
	p := c.params
	p.After = c.cursor

	uri, err := url.Parse(fmt.Sprintf(s.endpoint + "/webhooks"))

	if err != nil {
		return nil, err
	}

	var body io.Reader

	v, err := query.Values(p)
	if err != nil {
		return nil, err
	}
	uri.RawQuery = v.Encode()

	req, err := http.NewRequest("GET", uri.String(), body)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("GoCardless-Version", "2015-07-06")

	client := s.client
	if client == nil {
		client = http.DefaultClient
	}

	var result struct {
		Err *APIError `json:"error"`
		*WebhookListResult
	}

	err = try(3, func() error {
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		err = responseErr(res)

		if err != nil {
			return err
		}

		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return err
		}

		if result.Err != nil {
			return result.Err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if result.WebhookListResult == nil {
		return nil, errors.New("missing result")
	}

	c.response = result.WebhookListResult
	c.cursor = c.response.Meta.Cursors.After
	return c.response, nil
}

func (s *WebhookService) All(ctx context.Context, p WebhookListParams) *WebhookListPagingIterator {
	return &WebhookListPagingIterator{
		params:  p,
		service: s,
	}
}

// Get
// Retrieves the details of an existing webhook.
func (s *WebhookService) Get(ctx context.Context, identity string, opts ...RequestOption) (*Webhook, error) {
	uri, err := url.Parse(fmt.Sprintf(s.endpoint+"/webhooks/%v",
		identity))
	if err != nil {
		return nil, err
	}

	o := &requestOptions{
		retries: 3,
	}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	var body io.Reader

	req, err := http.NewRequest("GET", uri.String(), body)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+s.token)

	req.Header.Set("GoCardless-Version", "2015-07-06")

	req.Header.Set("GoCardless-Client-Library", "<no value>")
	req.Header.Set("GoCardless-Client-Version", "1.0.0")
	req.Header.Set("User-Agent", userAgent)

	for key, value := range o.headers {
		req.Header.Set(key, value)
	}

	client := s.client
	if client == nil {
		client = http.DefaultClient
	}

	var result struct {
		Err     *APIError `json:"error"`
		Webhook *Webhook  `json:"webhooks"`
	}

	err = try(o.retries, func() error {
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		err = responseErr(res)
		if err != nil {
			return err
		}

		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return err
		}

		if result.Err != nil {
			return result.Err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if result.Webhook == nil {
		return nil, errors.New("missing result")
	}

	return result.Webhook, nil
}

// Retry
// Requests for a previous webhook to be sent again
func (s *WebhookService) Retry(ctx context.Context, identity string, opts ...RequestOption) (*Webhook, error) {
	uri, err := url.Parse(fmt.Sprintf(s.endpoint+"/webhooks/%v/actions/retry",
		identity))
	if err != nil {
		return nil, err
	}

	o := &requestOptions{
		retries: 3,
	}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}
	if o.idempotencyKey == "" {
		o.idempotencyKey = NewIdempotencyKey()
	}

	var body io.Reader

	req, err := http.NewRequest("POST", uri.String(), body)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+s.token)

	req.Header.Set("GoCardless-Version", "2015-07-06")

	req.Header.Set("GoCardless-Client-Library", "<no value>")
	req.Header.Set("GoCardless-Client-Version", "1.0.0")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", o.idempotencyKey)

	for key, value := range o.headers {
		req.Header.Set(key, value)
	}

	client := s.client
	if client == nil {
		client = http.DefaultClient
	}

	var result struct {
		Err     *APIError `json:"error"`
		Webhook *Webhook  `json:"webhooks"`
	}

	err = try(o.retries, func() error {
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		err = responseErr(res)
		if err != nil {
			return err
		}

		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return err
		}

		if result.Err != nil {
			return result.Err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if result.Webhook == nil {
		return nil, errors.New("missing result")
	}

	return result.Webhook, nil
}
