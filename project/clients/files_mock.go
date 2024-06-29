package clients

import (
	"context"
	"sync"
)

type FilesServiceMock struct {
	mock sync.Mutex

	SavedFiles map[string]string
}

func (m *FilesServiceMock) UploadFile(ctx context.Context, fileID string, body string) error {
	m.mock.Lock()
	defer m.mock.Unlock()

	if m.SavedFiles == nil {
		m.SavedFiles = make(map[string]string)
	}

	m.SavedFiles[fileID] = body
	return nil
}
