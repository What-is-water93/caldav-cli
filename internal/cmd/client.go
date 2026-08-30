package cmd

import (
	"context"

	"github.com/What-is-water93/caldav-cli/internal/caldav"
)

type clientCtxKey struct{}

// clientFromContext returns the CalDAV client set up by the Before hook in root.go
func clientFromContext(ctx context.Context) *caldav.Client {
	return ctx.Value(clientCtxKey{}).(*caldav.Client)
}
