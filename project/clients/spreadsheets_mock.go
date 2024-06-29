package clients

import (
	"context"
	"sync"
)

type SpreadsheetsServiceMock struct {
	mock sync.Mutex

	AppendedRows map[string][][]string
}

func (m *SpreadsheetsServiceMock) AppendRow(ctx context.Context, sheetName string, row []string) error {
	m.mock.Lock()
	defer m.mock.Unlock()

	if m.AppendedRows == nil {
		m.AppendedRows = make(map[string][][]string)
	}

	m.AppendedRows[sheetName] = append(m.AppendedRows[sheetName], row)
	return nil
}
