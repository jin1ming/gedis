package ps

import (
	"context"
	"github.com/jin1ming/Gedis/pkg/event"
	"github.com/jin1ming/Gedis/pkg/utils"
	"github.com/tidwall/redcon"
	"log"
	"os"
	"path"
	"sync"
	"time"
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
	aofFile  *os.File
	survive  bool
	ChBuffer chan redcon.Command
	mu       sync.Mutex
}

var _ PersistentStorageService = &AOFService{}

func NewAOFService() *AOFService {
	aof := AOFService{ChBuffer: make(chan redcon.Command, 1024)}
	aofFilePath := path.Join(utils.GetHomeDir(), "gedis.aof")
	var err error
	aof.aofFile, err = os.OpenFile(aofFilePath, os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		log.Fatalln("Aof file cannot be opened.")
	}
	
	aof.survive = true
	return &aof
}

func (aof *AOFService) LoadLocalData() {

}

func (aof *AOFService) Start(ctx context.Context) {
	log.Println("AOFService is running...")

	go aof.work()

	tw := event.GetGlobalTimingWheel()
	tw.AfterFunc(1*time.Second, func() {
		_ = aof.aofFile.Sync()
	})

	for {
		select {
		case <-ctx.Done():
			log.Println("AOFService is closing...")
			aof.stop()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (aof *AOFService) work() {

	defer func() {
		_ = aof.aofFile.Close()
	}()

	var err error

	for {
		info := <-aof.ChBuffer
		if len(info.Args) == 0 {
			aof.mu.Lock()
			if aof.survive == false {
				break
			}
			aof.mu.Unlock()
			continue
		}

		for _, a := range info.Args {
			log.Println(string(a))
			_, err = aof.aofFile.Write(a)
			if err != nil {
				log.Fatalln("AOFService can't write:", err)
			}
		}
	}
}

func (aof *AOFService) stop() {
	close(aof.ChBuffer)
	aof.mu.Lock()
	defer aof.mu.Lock()
	aof.survive = false
}
