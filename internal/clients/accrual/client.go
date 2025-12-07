package accrual

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"

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

func (c *Client) GetAccrual(ctx context.Context, number string) (domain.Accrual, error) {
	var accrual domain.Accrual

	u := c.buildURL(number)
	resp, err := c.opts.Client.R().
		SetContext(ctx).
		SetResult(&accrual).
		Get(u)

	if errors.Is(err, io.EOF) {
		return domain.Accrual{}, nil
	}
	if err != nil {
		return domain.Accrual{}, fmt.Errorf("get error: %w", err)
	}
	if resp.IsError() {
		return domain.Accrual{}, fmt.Errorf("response error: %v", resp.Error())
	}

	return accrual, nil
}

func (c *Client) buildURL(number string) string {
	u := url.URL{
		Scheme: "http",
		Host:   c.opts.Addr,
		Path:   path.Join("api", "orders", number),
	}
	return u.String()
}
