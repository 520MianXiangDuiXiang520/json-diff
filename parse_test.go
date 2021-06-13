package json_diff

import (
	"github.com/520MianXiangDuiXiang520/json-diff/decode"
	"io/ioutil"
	"testing"
)

func BenchmarkMarshal(b *testing.B) {
	fileName := "./test_data/deepcopy_test/deepcopy_speed_test.json"
	// fileName := "./test_data/hash_test.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		b.Error("fail to open the ", fileName)
	}
	node, _ := decode.Unmarshal(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decode.Marshal(node)
	}
}

func BenchmarkMarshalOld(b *testing.B) {
	fileName := "./test_data/deepcopy_test/deepcopy_speed_test.json"
	// fileName := "./test_data/hash_test.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		b.Error("fail to open the ", fileName)
	}
	node, _ := Unmarshal(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Marshal(node)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	fileName := "./test_data/deepcopy_test/deepcopy_speed_test.json"
	// fileName := "./test_data/hash_test.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		b.Error("fail to open the ", fileName)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decode.Unmarshal(input)
	}
}

func BenchmarkUnmarshalOld(b *testing.B) {
	fileName := "./test_data/deepcopy_test/deepcopy_speed_test.json"
	// fileName := "./test_data/hash_test.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		b.Error("fail to open the ", fileName)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Unmarshal(input)
	}
}
