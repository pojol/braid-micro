package meta

type Service struct {
	Info ServiceInfo

	Nodes []Node
	Tags  []string
}

type ServiceInfo struct {
	ID   string
	Name string
}

// Node 发现节点结构
type Node struct {
	ID   string
	Name string

	Address string
	Port    int

	Metadata map[string]interface{}
}

func (n *Node) GetWidget() int {
	return 0
}

func (n *Node) SetWidget(widget int) {

}
