package node

import "testing"

// ReadImage()
func TestReadImageFail(t *testing.T) {
	node := &Node{}
	err := node.ReadImage()
	if err != nil && node.Image != "" {
		t.Errorf("If reading iamge failed, node image should be nil, got '%s'", node.Image)
	}
}

func TestReadImageOk(t *testing.T) {
	node := &Node{}
	err := node.ReadImage()
	if err == nil && node.Image == "" {
		t.Errorf("If reading image succeeded, node image should not be empty")
	}
}

// ReadState()
func TestReadState(t *testing.T) {
	node := &Node{}
	if err := node.ReadState(); err != nil {
		t.Errorf("Reading node state should not fail")
	}
}

func TestReadStateSetImage(t *testing.T) {
	node := &Node{}
	err := node.ReadState()
	if err == nil && node.Image == "" {
		t.Errorf("Reading node state should read node image")
	}
}

// Validate()
func TestNodeNoName(t *testing.T) {
	node := &Node{}
	if err := node.Validate(); err == nil {
		t.Errorf("Node without name should not be a valid node")
	}
}

func TestNodeValidate(t *testing.T) {
	node := &Node{
		Name: "foo",
	}
	if err := node.Validate(); err != nil {
		t.Errorf("Node should be valid")
	}
}

// node.SelectContainerRuntime()
func TestNodeSelectContainerRuntimeDocker(t *testing.T) {
	node := &Node{
		Name: "foo",
	}
	if _, err := node.SelectContainerRuntime(); err != nil {
		t.Errorf("Selecting container runtime for default node should work")
	}
}
