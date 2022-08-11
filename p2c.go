package p2c

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	lag = 0.0
	success = 1000.0
	inflight = 1
	cpu = 500
	penalty = 250 * time.Second
	forceGap = int64(3 * time.Second)
	tau = 600 * int64(time.Millisecond)
)

var (
	ErrNodesEmpty = errors.New("nodes is empty")
)

type p2cNode struct {
	weight int64

	lag int64
	success int64
	inflight int64
	cpu int64

	stamp int64 //last collected timestamp
	pick int64 // last pick timestamp

	pickTimes int64
}

func (n *p2cNode) valid() bool {
	return atomic.LoadInt64(&n.cpu) < 900 && atomic.LoadInt64(&n.success) > 900
}

func (n *p2cNode) load() int64 {
	lag := int64(math.Sqrt(float64(atomic.LoadInt64(&n.lag) + 1)))
	load := lag * (atomic.LoadInt64(&n.inflight)) * atomic.LoadInt64(&n.cpu)
	if load == 0 {
		load = int64(penalty)
	}
	return load
}

type P2CPicker struct {
	Nodes []*p2cNode
	mu sync.Mutex
	rand *rand.Rand
}

func (pp *P2CPicker) PrePick() (na *p2cNode, nb *p2cNode) {
	for i := 0;i < 3;i++ {
		pp.mu.Lock()
		a, b := pp.rand.Intn(len(pp.Nodes)), pp.rand.Intn(len(pp.Nodes)-1)
		pp.mu.Unlock()
		if a == b {
			b += 1
		}
		na, nb = pp.Nodes[a], pp.Nodes[b]
		if na.valid() || nb.valid() {
			break
		}
	}
	return na, nb
}

func (pp *P2CPicker) Pick() (*PickResult, error) {
	var pc, upc *p2cNode
	now := time.Now().UnixNano()
	pp.mu.Lock()
	l := len(pp.Nodes)
	pp.mu.Unlock()
	if l == 0 {
		return nil, ErrNodesEmpty
	} else if l == 1 {
		pc = pp.Nodes[0]
	} else {
		na, nb := pp.PrePick()
		if atomic.LoadInt64(&na.success) * na.weight * nb.load() > atomic.LoadInt64(&nb.success) * nb.weight * nb.load() {
			pc, upc = na, nb
		} else {
			pc, upc = nb, na
		}
		if atomic.LoadInt64(&upc.pick) + forceGap < now {
			pc = upc
		}
	}
	atomic.StoreInt64(&pc.pick, now)
	atomic.AddInt64(&pc.inflight, 1)
	atomic.AddInt64(&pc.pickTimes, 1)
	return &PickResult{
		pc,
		BuildDoneInfo(pc),
	}, nil
}

type DoneInfo struct {
	err error
	cpu int64
}

func BuildDoneInfo(node *p2cNode) func(di *DoneInfo) {
	return func(di *DoneInfo) {
		atomic.AddInt64(&node.inflight, -1)
		now := time.Now().UnixNano()
		td := now - atomic.LoadInt64(&node.stamp)
		oldLag := atomic.LoadInt64(&node.lag)
		w := math.Exp(float64(-td)/float64(tau))
		lag := int64(w*float64(oldLag) + (1-w)*float64(now-atomic.LoadInt64(&node.pick)))
		atomic.StoreInt64(&node.lag, lag)
		atomic.StoreInt64(&node.cpu, di.cpu)
		atomic.StoreInt64(&node.stamp, now)
		success := int64(1000)
		if di.err != nil {
			success = 0
		}
		oldSuccess := atomic.LoadInt64(&node.success)
		atomic.StoreInt64(&node.success, int64(w*float64(oldSuccess) + (1-w)*float64(success)))
	}
}

type PickResult struct {
	Node *p2cNode
	Done func(*DoneInfo)
}

func NewP2cNode(weight int64) *p2cNode {
	return &p2cNode {
		weight: weight,
		lag: lag,
		success: success,
		inflight: inflight,
		cpu: cpu,
	}
}