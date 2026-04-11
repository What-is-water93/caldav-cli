package caldav

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	ical "github.com/emersion/go-ical"
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

// ListEvents queries all VEVENT objects in the given calendar and returns
// them as go-ical events. Callers can read any iCalendar property directly
// off the underlying component (e.g. ev.Props.Get(ical.PropSummary)).
func (c *Client) ListEvents(ctx context.Context, calendarPath string) ([]ical.Event, error) {
	query := &gocaldav.CalendarQuery{
		CompRequest: gocaldav.CalendarCompRequest{
			Name: "VCALENDAR",
			Comps: []gocaldav.CalendarCompRequest{{
				Name: "VEVENT",
				Props: []string{
					"SUMMARY",
					"DTSTART",
					"DTEND",
					"UID",
					"DESCRIPTION",
					"LOCATION",
					"COLOR",
				},
			}},
		},
		CompFilter: gocaldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []gocaldav.CompFilter{{
				Name: "VEVENT",
			}},
		},
	}

	objects, err := c.inner.QueryCalendar(ctx, calendarPath, query)
	if err != nil {
		return nil, fmt.Errorf("querying calendar events: %w", err)
	}

	var events []ical.Event
	for _, obj := range objects {
		events = append(events, obj.Data.Events()...)
	}
	return events, nil
}

// UploadEvents splits a calendar into individual VEVENT resources and uploads
// each one separately, as required by the CalDAV spec.
//
// If color is non-empty, it is written to each VEVENT as the RFC 7986 COLOR
// property. The value must be a CSS3 color name (e.g. "blue").
func (c *Client) UploadEvents(ctx context.Context, calendarPath string, cal *ical.Calendar, color string) (int, error) {
	if !strings.HasSuffix(calendarPath, "/") {
		calendarPath += "/"
	}

	events := cal.Events()
	if len(events) == 0 {
		return 0, fmt.Errorf("ICS data does not contain any VEVENT components")
	}

	uploaded := 0
	for _, event := range events {
		uid := ""
		if prop := event.Props.Get(ical.PropUID); prop != nil {
			uid = prop.Value
		}
		if uid == "" {
			return uploaded, fmt.Errorf("VEVENT is missing a UID property")
		}

		if color != "" {
			event.Props.SetText("COLOR", color)
		}

		singleCal := ical.NewCalendar()
		singleCal.Props.SetText(ical.PropVersion, "2.0")
		singleCal.Props.SetText(ical.PropProductID, "-//caldav-cli//EN")
		singleCal.Children = append(singleCal.Children, event.Component)

		path := calendarPath + uid + ".ics"
		if _, err := c.inner.PutCalendarObject(ctx, path, singleCal); err != nil {
			return uploaded, fmt.Errorf("uploading event %q (UID %s): %w",
				event.Props.Get(ical.PropSummary).Value, uid, err)
		}
		uploaded++
	}

	return uploaded, nil
}

func (c *Client) ResolveCalendar(ctx context.Context, calendarPath string) (string, error) {
	if calendarPath != "" {
		return calendarPath, nil
	}

	calendars, err := c.ListCalendars(ctx)
	if err != nil {
		return "", err
	}

	switch len(calendars) {
	case 0:
		return "", fmt.Errorf("no calendars found on the server")
	case 1:
		fmt.Printf("Auto-selected calendar: %s (%s)\n", calendars[0].Name, calendars[0].Path)
		return calendars[0].Path, nil
	default:
		var msg strings.Builder
		msg.WriteString("multiple calendars found, please specify one with --calendar:\n")
		for _, cal := range calendars {
			fmt.Fprintf(&msg, "  - %s\t%s\n", cal.Path, cal.Name)
		}
		return "", fmt.Errorf("%s", msg.String())
	}
}
