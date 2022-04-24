package ps

import (
	"context"
)

type PersistentStorageService interface {
	LoadLocalData()
	work()
	Start(ctx context.Context)
	stop()
}
