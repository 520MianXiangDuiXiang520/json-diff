package json_diff

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/520MianXiangDuiXiang520/json-diff/decode"
	"sort"
	"strings"
)

func hashByMD5(s []byte) string {
	h := md5.New()
	h.Write(s)
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)

}

func hash(v interface{}) string {
	return hashByMD5([]byte(fmt.Sprintf("%v", v)))
}

func setHash(node *decode.JsonNode) string {
	var hashCode string
	switch node.Type {
	case decode.JsonNodeTypeObject:
		hashCode = setObjectHash(node)
	case decode.JsonNodeTypeSlice:
		hashCode = setSliceHash(node)
	case decode.JsonNodeTypeValue:
		hashCode = hash(node.Value)
		node.Hash = hashCode
	}
	return hashCode
}

func setObjectHash(node *decode.JsonNode) string {
	hashList := make([]string, len(node.ChildrenMap))
	for _, v := range node.ChildrenMap {
		hc := setHash(v)
		hashList = append(hashList, hc)
	}
	sort.Strings(hashList)
	hashCode := hash(strings.Join(hashList, ""))
	node.Hash = hashCode
	return hashCode
}

func setSliceHash(node *decode.JsonNode) string {
	h := bytes.NewBufferString("")
	for _, v := range node.Children {
		hc := setHash(v)
		h.WriteString(hc)
	}
	node.Hash = hash(h)
	return node.Hash
}
