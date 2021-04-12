package json_diff

import (
	"fmt"
	"testing"
)

func Test_keyReplace(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"common", args{key: "article~1a~01~001name"}, "article~01a~001~0001name"},
		{"common1", args{key: "01~1"}, "01~01"},
		{"common2", args{key: "0101~"}, "0101~"},
		{"common3", args{key: "0101/01"}, "0101~101"},
		{"common4", args{key: "0101/01~01"}, "0101~101~001"},
		{"common5", args{key: "article/name"}, "article~1name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyReplace(tt.args.key); got != tt.want {
				t.Errorf("keyReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keyRestore(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"common", args{key: "article~01a~001~0001name"}, "article~1a~01~001name"},
		{"common1", args{key: "article~1name"}, "article/name"},
		{"common2", args{key: "article_name"}, "article_name"},
		{"common3", args{key: "article~1a~001~0001name"}, "article/a~01~001name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyRestore(tt.args.key); got != tt.want {
				t.Errorf("keyRestore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleUnmarshal() {
	json := `{
        "A": 2,
        "B": [1, 2, 4],
        "C": {
          "CA": {"CAA": 1}
        }
      }`
	jsonNode := Unmarshal([]byte(json))
	fmt.Println(jsonNode)
}
