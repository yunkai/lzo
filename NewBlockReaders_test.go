package lzo

import (
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestNewBlockReaders(t *testing.T) {
	now := time.Now()

	runtime.GOMAXPROCS(runtime.NumCPU())

	f, err := os.Open("./testdata/input.TestL1CacheLzo.lzo")
	if err != nil {
		t.Fatal(err)
		return
	}

	chanCnt := 5
	blockChans := make([]chan *Blocker, chanCnt)
	for i := 0; i < chanCnt; i++ {
		blockChans[i] = make(chan *Blocker, 10)
	}

	z, err := NewBlockReaders(f, blockChans...)
	if err != nil {
		t.Fatal(err)
		return
	}

	z.StartBlockReaders()

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(chanCnt)

	for i := 0; i < chanCnt; i++ {
		go func(idx int) {
			defer waitGroup.Done()

			adler32 := adler32.New()
			crc32 := crc32.NewIEEE()
			for {
				select {
				case blocker, open := <-z.BlockerChans[idx]:
					if !open {
						fmt.Fprintf(os.Stderr, "BlockerChans[%d] has been closed\n", idx)
						if z.Error() != io.EOF {
							t.Fatalf("BlockerChans[%d] error:%v\n", idx, z.Error())
						}
						return
					}
					data, err := z.BlockDecompress(blocker, adler32, crc32)
					if err != nil {
						t.Fatalf("BlockerChans[%d] error:%v", idx, err)
						return
					}

					fmt.Printf(string(data))
				}
			}
		}(i)
	}

	waitGroup.Wait()
	fmt.Fprintf(os.Stderr, "elapsed:%v, err:%v\n", time.Since(now), z.err)

}
