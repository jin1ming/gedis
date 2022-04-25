package ps

import (
	"bufio"
	"context"
	"github.com/jin1ming/Gedis/pkg/db"
	"github.com/jin1ming/Gedis/pkg/event"
	"github.com/jin1ming/Gedis/pkg/utils"
	"github.com/tidwall/redcon"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
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
	filePath string
}

var _ PersistentStorageService = &AOFService{}

func NewAOFService() *AOFService {
	aof := AOFService{
		ChBuffer: make(chan redcon.Command, 1024),
		survive:  true,
		filePath: path.Join(utils.GetHomeDir(), "gedis.aof"),
	}

	return &aof
}

func (aof *AOFService) LoadLocalData() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	aofFile, err := os.Open(aof.filePath)
	if err != nil {
		log.Println("Aof file failed to open. ->LoadLocalData")
		return
	}

	defer func() {
		_ = aofFile.Close()
	}()

	reader := bufio.NewReader(aofFile)
	for {
		data, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("Aof file failed to load.")
		}
		if len(data) == 0 {
			continue
		}
		if data[0] != '*' || len(data) == 1 {
			log.Fatalln("the header of data is not \"*\" + number.")
		}
		argsNum, err := strconv.Atoi(strings.TrimSpace(string(data[1:])))
		if err != nil {
			log.Fatalln(err)
		}
		var args [][]byte
		for i := 0; i < argsNum; i++ {
			nByte, _, _ := reader.ReadLine()
			n, _ := strconv.Atoi(strings.TrimSpace(string(nByte[1:])))
			a, _, _ := reader.ReadLine()
			if n != len(a) {
				log.Fatalln("Command parameter parsing failed")
			}
			args = append(args, a)
		}

		db.GetDB().ExecQueue <- db.CmdPackage{
			Args: args,
			Ch:   nil,
		}
	}
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
	var err error
	aof.aofFile, err = os.OpenFile(aof.filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalln("Aof file failed to open. -> work")
	}

	defer func() {
		_ = aof.aofFile.Close()
	}()

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

		aof.writeLine([]byte("*" + strconv.Itoa(len(info.Args))))
		for _, a := range info.Args {
			aof.writeLine([]byte("$" + strconv.Itoa(len(a))))
			aof.writeLine(a)
		}
	}
}

func (aof *AOFService) writeLine(line []byte) {
	_, err := aof.aofFile.Write(append(line, '\n'))
	if err != nil {
		log.Fatalln("AOFService can't write:", err)
	}
}

func (aof *AOFService) stop() {
	close(aof.ChBuffer)
	aof.mu.Lock()
	defer aof.mu.Lock()
	aof.survive = false
}
