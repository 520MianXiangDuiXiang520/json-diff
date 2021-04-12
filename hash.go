package json_diff

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

func setHash(node *JsonNode) string {
	var hashCode string
	switch node.Type {
	case JsonNodeTypeObject:
		hashCode = setObjectHash(node)
	case JsonNodeTypeSlice:
		hashCode = setSliceHash(node)
	case JsonNodeTypeValue:
		hashCode = hash(node.Value)
		node.Hash = hashCode
	}
	return hashCode
}

func setObjectHash(node *JsonNode) string {
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

func setSliceHash(node *JsonNode) string {
	h := bytes.NewBufferString("")
	for _, v := range node.Children {
		hc := setHash(v)
		h.WriteString(hc)
	}
	node.Hash = hash(h)
	return node.Hash
}
