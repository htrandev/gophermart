package accrual

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/htrandev/gophermart/internal/domain"
	"go.uber.org/zap"
)

type ClientOptions struct {
	MaxRetry int
	Addr     string
	Client   *resty.Client
	Logger   *zap.Logger
}

func defaultOptions() *ClientOptions {
	return &ClientOptions{
		MaxRetry: 3,
		Addr:     "localhost:8080",
		Client:   resty.New(),
		Logger:   zap.NewNop(),
	}
}

func validateOptions(opts *ClientOptions) *ClientOptions {
	if opts == nil {
		return defaultOptions()
	}

	if opts.MaxRetry <= 0 {
		opts.MaxRetry = 3
	}

	if opts.Addr == "" {
		opts.Addr = "localhost:8080"
	}

	if opts.Client == nil {
		opts.Client = resty.New()
	}

	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	return opts
}

type Client struct {
	opts *ClientOptions
}

func NewClient(opts *ClientOptions) *Client {
	return &Client{opts: validateOptions(opts)}
}

func (c *Client) GetAccrual(ctx context.Context, number string) (domain.ClientResponse, error) {
	var accrual domain.Accrual

	u := c.buildURL(number)
	resp, err := c.opts.Client.R().
		SetContext(ctx).
		SetResult(&accrual).
		Get(u)

	if errors.Is(err, io.EOF) {
		return domain.ClientResponse{}, nil
	}
	if err != nil {
		return domain.ClientResponse{}, fmt.Errorf("get error: %w", err)
	}
	if resp.IsError() {
		return domain.ClientResponse{}, fmt.Errorf("response error: %v", resp.Error())
	}

	if resp.StatusCode() == http.StatusNoContent {
		return domain.ClientResponse{}, domain.ErrNotFound
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		duration, _ := parseRetryAfter(resp.Header().Get("Retry-After"))
		return domain.ClientResponse{RetryAfter: duration}, domain.ErrTooManyRequests
	}

	if resp.StatusCode() != http.StatusOK {
		return domain.ClientResponse{}, fmt.Errorf("get accural error with status code: %v", resp.StatusCode())
	}

	return domain.ClientResponse{
		Accrual: accrual,
	}, nil
}

func (c *Client) buildURL(number string) string {
	return c.opts.Addr + "/" + path.Join("api", "orders", number)
}

func parseRetryAfter(h string) (time.Duration, error) {
	if d, err := strconv.ParseInt(h, 10, 64); err == nil {
		return time.Duration(d) * time.Second, nil
	}
	t, err := time.Parse(time.RFC1123, h)
	if err != nil {
		return 0, err
	}
	return time.Until(t), nil
}
