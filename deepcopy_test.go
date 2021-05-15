package json_diff

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"testing"
	"time"
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
	fileName := "./test_data/deepcopy_test/deepcopy_test.json"
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

func oldDeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func TestDeepCopy_speed(t *testing.T) {
	fileName := "./test_data/deepcopy_test/deepcopy_speed_test.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error("fail to open the ", fileName)
	}
	srcNode, _ := Unmarshal(input)
	deepCopySpeedTestHelper(srcNode, t, 1000, true, true)
	deepCopySpeedTestHelper(srcNode, t, 1000, true, false)
	deepCopySpeedTestHelper(srcNode, t, 1000, false, true)
}

func deepCopySpeedTestHelper(srcNode *JsonNode, t *testing.T, loop int, doOld, doNew bool) {
	path := "test_data/deepcopy_test/pprof_result"
	name := ""
	var err error
	if doOld && doNew {
		name = fmt.Sprintf("compared_cpu_profile_%d", time.Now().UnixNano())
	} else if doOld {
		name = fmt.Sprintf("old_func_cpu_profile_%d", time.Now().UnixNano())
	} else if doNew {
		name = fmt.Sprintf("new_func_cpu_profile_%d", time.Now().UnixNano())
	}
	f, err := os.Create(fmt.Sprintf("%s/%s", path, name))
	if err != nil {
		t.Errorf("can not Start CPU Profile: %v", err)
	}
	_ = pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	startTime := time.Now().UnixNano()
	for i := 0; i < loop; i++ {
		if doOld {
			oldCopyRes := new(JsonNode)
			err = oldDeepCopy(oldCopyRes, srcNode)
			if err != nil {
				t.Errorf("fail to copy by old function: %v", err)
			}
		}
		if doNew {
			_, err = DeepCopy(srcNode)
			if err != nil {
				t.Errorf("fail to copy by new function: %v", err)
			}
		}
	}
	spend := (time.Now().UnixNano() - startTime) / 1000000
	fmt.Printf("do %d loops spend %v ms \n", loop, spend)
}
