package json_diff

import (
    `fmt`
    `log`
)

func ExampleJsonNode_ADD() {
    json := `{
        "A": 2,
        "B": [1, 2, 4],
        "C": {
          "CA": {"CAA": 1}
        }
      }`
    jsonNode, err := Unmarshal([]byte(json))
    if err != nil {
        log.Println(err)
        return
    }
    dValueStr := `{"DA": 1}`
    dObj, err := Unmarshal([]byte(dValueStr))
    if err != nil {
        log.Println(err)
        return
    }
    jsonNode.ADD("D", dObj)
    newJsonStr, err := Marshal(jsonNode)
    if err != nil {
        log.Println(err)
        return
    }
    fmt.Println(string(newJsonStr))
}