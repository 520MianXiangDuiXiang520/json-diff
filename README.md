# Json-Diff

[RFC 6902](https://tools.ietf.org/html/rfc6902) 的 Go 语言实现

[![GoDoc](https://camo.githubusercontent.com/ba58c24fb3ac922ec74e491d3ff57ebac895cf2deada3bf1c9eebda4b25d93da/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f67616d6d617a65726f2f776f726b6572706f6f6c3f7374617475732e737667)](https://pkg.go.dev/github.com/520MianXiangDuiXiang520/json-diff)

<a title="Apache License 2.0" target="_blank" href="https://github.com/520MianXiangDuiXiang520/json-diff/LICENSE"><img src="https://img.shields.io/badge/license-MIT-red.svg?style=flat-square"></a>

<a href="https://goreportcard.com/badge/github.com/520MianXiangDuiXiang520/json-diff"> <img src="https://goreportcard.com/badge/github.com/520MianXiangDuiXiang520/json-diff" /></a>
<a href="https://codeclimate.com/github/520MianXiangDuiXiang520/json-diff/maintainability"><img src="https://api.codeclimate.com/v1/badges/ed575aea812a025dfcc9/maintainability" /></a>

```shell
go get -u github.com/520MianXiangDuiXiang520/json-diff
```

## 功能：

### 序列化与反序列化

与官方 json 包的序列化和反序列化不同，官方包序列化需要指定一个 `interface{}`, 像：

```go
package main

import "json"

func main() {
  jsonStr := "{}"
  var jsonObj interface{}
  node := json.Unmarshal(&jsonObj, []byte(jsonStr))
  // ...
}
```

这样不方便编辑反序列化后的 json 对象， json-diff 可以将任意的 json 串转换成统一的 `JsonNode` 类型，并且提供一系列的增删查改方法，方便操作对象：

```go
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
```

### 差异比较

通过对比两个 Json 串，输出他们的差异或者通过差异串得到修改后的 json 串

```go
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
```

```go
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
```

#### 输出格式

输出一个 json 格式的字节数组，类似于：

```json
   [
     { "op": "test", "path": "/a/b/c", "value": "foo" },
     { "op": "remove", "path": "/a/b/c" },
     { "op": "add", "path": "/a/b/c", "value": [ "foo", "bar" ] },
     { "op": "replace", "path": "/a/b/c", "value": 42 },
     { "op": "move", "from": "/a/b/c", "path": "/a/b/d" },
     { "op": "copy", "from": "/a/b/d", "path": "/a/b/e" }
   ]
```

其中数组中的每一项代表一个差异点，格式由 RFC 6902 定义，op 表示差异类型，有六种：

1. `add`: 新增
2. `replace`: 替换
3. `remove`: 删除
4. `move`: 移动
5. `copy`: 复制
6. `test`: 测试

其中 move 和 copy 可以减少差异串的体积，但会增加差异比较的时间, 可以通过修改 `AsDiff()` 的 options 指定是否开启，options 的选项和用法如下：

```go
  // 返回差异时使用 Copy, 当发现新增的子串出现在原串中时，使用该选项可以将 Add 行为替换为 Copy 行为
  // 以减少差异串的大小，但这需要额外的计算，默认不开启
  UseCopyOption JsonDiffOption = 1 << iota

  // 仅在 UseCopyOption 选项开启时有效，替换前会添加 Test 行为，以确保 Copy 的路径存在
  UseCheckCopyOption

  // 返回差异时使用 Copy, 当发现差异串中两个 Add 和 Remove 的值相等时，会将他们合并为一个 Move 行为
  // 以此减少差异串的大小，默认不开启
  UseMoveOption

  // Remove 时除了返回 path, 还返回删除了的值，默认不开启
  UseFullRemoveOption
```

#### 相等的依据

对于一个对象，其内部元素的顺序不作为相等判断的依据，如

```json
{
  "a": 1,
  "b": 2,
}
```

和

```json
{
  "b": 2,
  "a": 1,
}
```

被认为是相等的。

对于一个列表，元素顺序则作为判断相等的依据，如：

```json
{
  "a": [1, 2]
}
```

和

```json
{
  "a": [2, 1]
}
```

被认为不相等。

只有一个元素的所有子元素全部相等，他们才相等

#### 原子性

根据 RFC 6092，差异合并应该具有原子性，即列表中有一个差异合并失败，之前的合并全部作废，而 test 类型就用来在合并差异之前检查路径和值是否正确，你可以通过选项开启它，但即便不使用 test，合并也是原子性的。

json-diff 在合并差异前会深拷贝源数据，并使用拷贝的数据做差异合并，一旦发生错误，将会返回 nil, 任何情况下都不会修改原来的数据。

## 参考

[https://github.com/flipkart-incubator/zjsonpatch](https://github.com/flipkart-incubator/zjsonpatch)
