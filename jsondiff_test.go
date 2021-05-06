package json_diff

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func ExampleAsDiffs() {
	json1 := `{
        "A": 1,
        "B": [1, 2, 3],
        "C": {
          "CA": 1
        }
      }`
	json2 := `{
        "A": 2,
        "B": [1, 2, 4],
        "C": {
          "CA": {"CAA": 1}
        }
      }`
	res, _ := AsDiffs([]byte(json1), []byte(json2), UseMoveOption, UseCopyOption, UseFullRemoveOption)
	fmt.Println(res)
}

func ExampleMergeDiff() {
	json2 := `{
        "A": 1,
        "B": [1, 2, 3, {"BA": 1}],
        "C": {
          "CA": 1,
          "CB": 2
        }
      }`
	diffs := `[
        {"op": "move", "from": "/A", "path": "/D"},
        {"op": "move", "from": "/B/0", "path": "/B/1"},
        {"op": "move", "from": "/B/2", "path": "/C/CB"}
      ]`
	res, _ := MergeDiff([]byte(json2), []byte(diffs))
	fmt.Println(res)
}

func Test_merge_smoke(t *testing.T) {
	type args struct {
		srcNode  *JsonNode
		diffNode *JsonNode
	}
	type testCase struct {
		name string
		args args
		want *JsonNode
	}
	fileName := "./test_data/mergeSmoke.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error("fail to open the ", fileName)
	}
	caseNode, _ := Unmarshal(input)
	cases := caseNode.ChildrenMap["cases"]
	testCases := make([]testCase, len(cases.Children))
	for i, tt := range cases.Children {
		name := tt.ChildrenMap["name"].Value.(string)
		srcNode := tt.ChildrenMap["src"]
		diffNode := tt.ChildrenMap["diff"]
		wantNode := tt.ChildrenMap["hope"]
		testCases[i] = testCase{
			name: name,
			args: args{
				srcNode:  srcNode,
				diffNode: diffNode,
			},
			want: wantNode,
		}
	}

	for _, cs := range testCases {
		t.Run(cs.name, func(t *testing.T) {
			src := new(JsonNode)
			err := DeepCopy(src, cs.args.srcNode)
			if err != nil {
				t.Errorf("fail to deepcopy src object")
			}
			res, err := MergeDiffNode(src, cs.args.diffNode)
			if err != nil {
				t.Errorf("fail to do merge(), get error: %v", err)
			}
			if !res.Equal(cs.want) {
				get, _ := Marshal(src)
				want, _ := Marshal(cs.want)
				t.Errorf("the value of after merge(%v) are not equal want(%v)", string(get), string(want))
			}

		})
	}
}

func m(n *JsonNode) []byte {
	r, _ := Marshal(n)
	return r
}

func getOptions(n *JsonNode) []JsonDiffOption {
	res := make([]JsonDiffOption, len(n.Children))
	for i, v := range n.Children {
		res[i] = JsonDiffOption(v.Value.(float64))
	}
	return res
}

func TestGetDiff(t *testing.T) {
	type args struct {
		source  []byte
		patch   []byte
		options []JsonDiffOption
	}
	fileName := "./test_data/getDiffTest.json"
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Error("fail to open the ", fileName)
	}
	caseNode, _ := Unmarshal(input)
	cases := caseNode.ChildrenMap["cases"].Children
	type ts struct {
		name    string
		args    args
		want    *JsonNode
		wantErr bool
	}
	tss := make([]ts, len(cases))
	for i, c := range cases {
		name := c.ChildrenMap["name"].Value.(string)
		source := m(c.ChildrenMap["src"])
		patch := m(c.ChildrenMap["patch"])
		options := getOptions(c.ChildrenMap["options"])
		want := c.ChildrenMap["want"]
		wantErr := c.ChildrenMap["want-error"].Value.(bool)
		tss[i] = ts{
			name: name,
			args: args{
				source:  source,
				patch:   patch,
				options: options,
			},
			want:    want,
			wantErr: wantErr,
		}
	}
	for _, tt := range tss {
		t.Run(tt.name, func(t *testing.T) {
		    src, _ := Unmarshal(tt.args.source)
		    pat, _ := Unmarshal(tt.args.patch)
			diffs := GetDiffNode(src, pat, tt.args.options...)
			if !eq(diffs, tt.want) {
				got, _ := Marshal(diffs)
				want, _ := Marshal(tt.want)
				t.Errorf("getDiff() got %s\n, but want %s", string(got), string(want))
			}
		})
	}
}

func eq(a, b *JsonNode) bool {
	aList := a.Children
	bList := b.Children
	for i := 0; i < len(aList); i++ {
		j := 0
		for ; j < len(bList); j++ {
			if aList[i].Equal(bList[j]) {
				break
			}
		}
		if j == len(bList) {
			return false
		}
	}
	return true
}
