package node

import (
	"fmt"

	"github.com/Github-Aiko/Aiko-Server/api/panel"
	"github.com/Github-Aiko/Aiko-Server/src/conf"
	vCore "github.com/Github-Aiko/Aiko-Server/src/core"
)

type Node struct {
	controllers []*Controller
}

func New() *Node {
	return &Node{}
}

func (n *Node) Start(nodes []*conf.NodeConfig, core vCore.Core) error {
	n.controllers = make([]*Controller, len(nodes))
	for i, c := range nodes {
		p, err := panel.New(c.ApiConfig)
		if err != nil {
			return err
		}
		// Register controller service
		n.controllers[i] = NewController(core, p, c.Options)
		err = n.controllers[i].Start()
		NodeType := c.ApiConfig.NodeType

		if err != nil {
			return fmt.Errorf("start node controller [%s-%s-%d] error: %s",
				c.ApiConfig.APIHost,
				NodeType,
				c.ApiConfig.NodeID,
				err)
		}
	}
	return nil
}

func (n *Node) Close() {
	for _, c := range n.controllers {
		err := c.Close()
		if err != nil {
			panic(err)
		}
	}
	n.controllers = nil
}
