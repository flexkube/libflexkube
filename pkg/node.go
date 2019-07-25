package etcd

// Node represents cluster member, information how it should be set up, how
// to connect to it etc.
type Node struct {
	// Unique indentier for the node.
	Name string
	// Docker image name with tag of the node.
	Image string
}

// This function reads general state of the node.
func (node *Node) ReadState() error {
	if err := node.ReadImage(); err != nil {
		return err
	}
	return nil
}

// This function reads which image is currently running on existing deployment
// and sets it in the node object.
func (node *Node) ReadImage() error {
	node.Image = "notimplemented"
	return nil
}
