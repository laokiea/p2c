package p2c

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestP2cPick(t *testing.T) {
	tests := []struct {
		name string
		nodeNum int
	}{
		//{"empty", 0},
		//{"single", 1},
		//{"double", 2},
		{"multiple", 5},
	}

	total := 1000

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(total)
			t.Parallel()
			picker := P2CPicker{rand: rand.New(rand.NewSource(time.Now().Unix()))}
			for i := 0;i < test.nodeNum;i++ {
				picker.Nodes = append(picker.Nodes, NewP2cNode(picker.rand.Int63n(100)))
			}
			for j := 1;j <= total;j++ {
				pr, err := picker.Pick()
				if err != nil {
					return
				}
				//atomic.AddInt64(&pr.Node.cpu, 1)
				di := &DoneInfo{err: err, cpu: picker.rand.Int63n(1000)}
				go func() {
					defer wg.Done()
					runtime.Gosched()
					pr.Done(di)
				}()
				if j%100 == 0 {
					fmt.Println(test.name, ": ")
					for _, node := range picker.Nodes {
						fmt.Printf("node stat: %#v\n", node)
					}
				}
			}
			wg.Wait()
		})
	}
}
