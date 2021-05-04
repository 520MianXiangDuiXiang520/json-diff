package json_diff

import (
	"io/ioutil"
	"testing"
)

func TestDeepCopy(t *testing.T) {
	fileName := "./test_data/mergeSmoke.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error("fail to open the ", fileName)
	}
	src, _ := Unmarshal(input)
	type args struct {
		dst interface{}
		src interface{}
	}
	dst := new(JsonNode)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "smoke", args: args{
			dst: dst,
			src: src,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeepCopy(tt.args.dst, tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("DeepCopy() error = %v, wantErr %v", err, tt.wantErr)
			}
			if dst == nil || !dst.Equal(src) {
				t.Errorf("Values are not equal after DeepCopy, dst is %v, but src is %v", dst, src)
			}
			dst.Value = "dst"
			if dst.Value == src.Value {
				t.Errorf("dst and src point to the same object")
			}
		})
	}
}
