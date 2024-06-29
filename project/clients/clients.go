package clients

import (
	"context"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func NewClients(gatewayAddress string) (*clients.Clients, error) {
	return clients.NewClients(
		gatewayAddress,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
}
