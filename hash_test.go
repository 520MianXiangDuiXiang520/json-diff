package json_diff

import (
	"io/ioutil"
	"testing"
)

func Test_setHash(t *testing.T) {
	fileName := "./test_data/hash_test.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error("fail to open the ", fileName)
	}
	inputNode, _ := Unmarshal(input)
	hashCode := setHash(inputNode)
	for i := 0; i < 100; i++ {
		inputNode, _ := Unmarshal(input)
		hc := setHash(inputNode)
		if hc != hashCode {
			t.Errorf("Get a different hashcode(%s)", hc)
		}
	}
}
