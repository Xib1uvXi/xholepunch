package predictor

import (
	"math/rand"
	"time"
)

type PseudorandomPortPredictor struct {
	targetPort int
	size       int
	ports      []int
	portBitmap *PortBitmap
}

func (p *PseudorandomPortPredictor) NextPort() int {
	// 输出 ports中的第一个，然后重置ports
	if len(p.ports) == 0 {
		return 0
	}

	port := p.ports[0]
	p.ports = p.ports[1:]

	return port
}

func NewPseudorandomPortPredictor(targetPort, size int) *PseudorandomPortPredictor {
	p := &PseudorandomPortPredictor{targetPort: targetPort, size: size, ports: make([]int, 0, size), portBitmap: NewPortBitmap()}

	// 主要应对一些简单线性端口预测算法
	p.initLinear()

	// 生成随机端口
	p.initLRandom(size)

	return p
}

func (p *PseudorandomPortPredictor) initLinear() {
	// 将target port前后20个端口放入 ports
	for i := p.targetPort - 30; i <= p.targetPort+30; i++ {
		if i < 0 || i > 65535 {
			continue
		}

		//if i == p.targetPort {
		//	continue
		//}

		used, err := p.portBitmap.IsPortSet(i)
		if err != nil {
			continue
		}

		if !used {
			p.ports = append(p.ports, i)
		}
	}
}

func (p *PseudorandomPortPredictor) initLRandom(size int) {
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))

	for times := 0; times < size; {
		port := gen.Intn(65535)
		if used, _ := p.portBitmap.IsPortSet(port); used {
			continue
		}

		if port < 1024 {
			continue
		}

		_ = p.portBitmap.SetPort(port)
		p.ports = append(p.ports, port)
		times++
	}
}
