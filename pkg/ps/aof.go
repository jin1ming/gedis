package ps

import (
	"context"
	"github.com/jin1ming/Gedis/pkg/utils"
	"log"
	"os"
	"path"
	"sync"
)

/*
例如：set count 1
存储格式为：
	*3  # 共计三个参数，分别是set count 1
	$3  # 参数长度为3个B（备注：set长度）
	set
	$5 # 下一个参数长度是5
	count
	$1 #下一个参数长度为1
	1
*/

type AOFService struct {
	aofFile *os.File
	survive bool
	buffer  chan string
	mu      *sync.Mutex
}

var _ PersistentStorageService = AOFService{}

func (aof AOFService) LoadData() {

}

func (aof AOFService) Start(ctx context.Context) {
	aofFilePath := path.Join(utils.GetHomeDir(), "gedis.aof")
	aofFile, err := os.OpenFile(aofFilePath, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("Aof file cannot be opened.")
	}
	defer func() {
		_ = aofFile.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			if len(aof.buffer) == 0 {
				return
			}
		default:

		}
	}
}

func (aof AOFService) WriteLine(line string) {

}

func (aof AOFService) Stop() {

}
