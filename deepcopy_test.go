package json_diff

import (
	"io/ioutil"
	"testing"
)

func TestDeepCopy_emptyObject(t *testing.T) {
	srcStr := "{}"
	srcNode, _ := Unmarshal([]byte(srcStr))
	cp, err := DeepCopy(srcNode)
	if err != nil {
		t.Errorf("got an error: %v", err)
	}
	err = cp.ADD("child", NewValueNode(1, 2))
	if err != nil {
		t.Errorf("fail to add child: %v", err)
	}
	if _, ok := srcNode.ChildrenMap["child"]; ok {
		t.Errorf("change source object: %s", srcNode.String())
	}
}

func TestDeepCopy(t *testing.T) {
	fileName := "./test_data/deepcopy_test.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error("fail to open the ", fileName)
	}
	srcNode, _ := Unmarshal(input)
	cp, err := DeepCopy(srcNode)
	if err != nil {
		t.Errorf("got an error: %v", err)
	}
	err = cp.ADD("add_child", NewValueNode(1, 2))
	if err != nil {
		t.Errorf("fail to add child: %v", err)
	}
	err = cp.ADD("add_child", NewValueNode(1, 2))
	if err != nil {
		t.Errorf("fail to add child: %v", err)
	}
	// 更改深层 object
	path1 := "/obj/obj/c/ce/2/ceb"
	cpof, ok := cp.Find(path1)
	if !ok {
		t.Errorf("can not find: %s in cp node", path1)
	}
	cpof.Value = true
	of, ok := srcNode.Find(path1)
	if !ok {
		t.Errorf("can not find: %s in src node", path1)
	}
	if of.Value.(bool) {
		t.Errorf("change source object: %v", cpof.Value)
	}
	// 更改深层 slice
	path2 := "/obj/slice/4/bb/1"
	cpof, ok = cp.Find(path1)
	if !ok {
		t.Errorf("can not find: %s in cp node", path2)
	}
	cpof.Value = "ggg"
	of, ok = srcNode.Find(path2)
	if !ok {
		t.Errorf("can not find: %s in src node", path2)
	}
	if of.Value.(string) != "jjj" {
		t.Errorf("change source object: %v", cpof.Value)
	}
}
