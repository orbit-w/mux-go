package multiplexers

import (
	"github.com/orbit-w/mux-go"
)

type Cell struct {
	mux.IMux
}

func (c *Cell) Dial(host string, maxConns int) {

}
