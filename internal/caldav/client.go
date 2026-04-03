package caldav

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emersion/go-webdav"
	gocaldav "github.com/emersion/go-webdav/caldav"
)

type Client struct {
	inner *gocaldav.Client
}

type Calendar struct {
	Name        string
	Path        string
	Description string
}

func NewClient(ctx context.Context, baseURL, username, password string) (*Client, error) {
	httpClient := webdav.HTTPClientWithBasicAuth(http.DefaultClient, username, password)

	client, err := gocaldav.NewClient(httpClient, baseURL)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	return &Client{inner: client}, nil
}

func (c *Client) ListCalendars(ctx context.Context) ([]Calendar, error) {
	principal, err := c.inner.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, fmt.Errorf("finding principal: %w", err)
	}

	fmt.Printf("Principal: %s\n", principal)

	homeSet, err := c.inner.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("finding calendar home set: %w", err)
	}

	cals, err := c.inner.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("finding calendars: %w", err)
	}

	result := make([]Calendar, 0, len(cals))
	for _, cal := range cals {
		result = append(result, Calendar{
			Name:        cal.Name,
			Path:        cal.Path,
			Description: cal.Description,
		})
	}
	return result, nil
}
