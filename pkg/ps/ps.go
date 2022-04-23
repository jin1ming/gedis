package ps

import (
	"context"
)

type PersistentStorageService interface {
	LoadData()
	WriteLine(line string)
	Start(ctx context.Context)
	Stop()
}
