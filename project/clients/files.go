package clients

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

type filesClient struct {
	clients *clients.Clients
}

func NewFilesClient(clients *clients.Clients) filesClient {
	return filesClient{
		clients: clients,
	}
}

func (c filesClient) UploadFile(ctx context.Context, fileID string, body string) error {
	resp, err := c.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(
		ctx,
		fileID,
		body,
	)
	if err != nil {
		return fmt.Errorf("failed to put files: %w", err)
	}

	if resp.StatusCode() == http.StatusConflict {
		log.FromContext(ctx).Infof("file %s already exists", fileID)
		return nil
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode())
	}

	return nil
}
