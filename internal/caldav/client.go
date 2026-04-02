package caldav

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/emersion/go-webdav"
	gocaldav "github.com/emersion/go-webdav/caldav"
)

type Client struct {
	inner    *gocaldav.Client
	baseURL  string
	username string
	password string
}

type Calendar struct {
	Name        string
	Path        string
	Description string
}

func NewClient(ctx context.Context, baseURL, username, password string) (*Client, error) {
	httpClient := webdav.HTTPClientWithBasicAuth(http.DefaultClient, username, password)

	discovered, err := discoverContextURL(ctx, baseURL, username, password)
	if err == nil {
		baseURL = discovered
	}

	client, err := gocaldav.NewClient(httpClient, baseURL)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	return &Client{inner: client, baseURL: baseURL, username: username, password: password}, nil
}

func (c *Client) ListCalendars(ctx context.Context) ([]Calendar, error) {
	principal, err := c.findPrincipal(ctx)
	if err != nil {
		return nil, fmt.Errorf("finding principal: %w", err)
	}

	// Use path-only — the server rejects full URLs for PROPFIND
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

// findPrincipal sends a manual PROPFIND to the CalDAV root to discover
// the current user's principal URL.
func (c *Client) findPrincipal(ctx context.Context) (string, error) {
	body := `<?xml version="1.0" encoding="utf-8"?>
<d:propfind xmlns:d="DAV:">
  <d:prop>
    <d:current-user-principal/>
  </d:prop>
</d:propfind>`

	req, err := http.NewRequestWithContext(ctx, "PROPFIND", c.baseURL, bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Depth", "0")
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 207 {
		return "", fmt.Errorf("PROPFIND returned %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return parsePrincipalHref(respBody)
}

type multistatus struct {
	XMLName   xml.Name   `xml:"multistatus"`
	Responses []response `xml:"response"`
}

type response struct {
	PropStats []propstat `xml:"propstat"`
}

type propstat struct {
	Prop prop `xml:"prop"`
}

type prop struct {
	CurrentUserPrincipal currentUserPrincipal `xml:"current-user-principal"`
}

type currentUserPrincipal struct {
	Href string `xml:"href"`
}

func parsePrincipalHref(data []byte) (string, error) {
	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return "", fmt.Errorf("parsing XML: %w", err)
	}

	for _, r := range ms.Responses {
		for _, ps := range r.PropStats {
			if href := ps.Prop.CurrentUserPrincipal.Href; href != "" {
				return href, nil
			}
		}
	}

	return "", fmt.Errorf("no current-user-principal found in response")
}

func discoverContextURL(ctx context.Context, baseURL, username, password string) (string, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	wellKnown := baseURL + "/.well-known/caldav"

	req, err := http.NewRequestWithContext(ctx, "PROPFIND", wellKnown, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Depth", "0")

	noRedirect := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := resp.Header.Get("Location")
		if loc != "" {
			base, _ := url.Parse(wellKnown)
			r, _ := url.Parse(loc)
			return base.ResolveReference(r).String(), nil
		}
	}

	if resp.StatusCode == 207 {
		return wellKnown, nil
	}

	return "", fmt.Errorf("well-known returned %d", resp.StatusCode)
}
