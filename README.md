# Json-Diff

[RFC 6092](https://tools.ietf.org/html/rfc6902) 的 Go 语言实现

## 功能：

* 不依赖于 `struct` 的 JSON 的序列化与反序列化
* 两个 JSON 串的差异比较
* 根据差异还原原 JSON 串

## 使用

```shell
go get -u github.com/520MianXiangDuiXiang520/json-diff
```

### 序列化与反序列化

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

`Unmarshal()` 将任意合法的 JSON 串反序列化成 `*JsonNode` 类型，基于该类型，可以更方便地操作 JSON 对象

`JsonNode` 包含以下方法：

* `Find(path string)`: 从当前 `JsonNode` 中找到满足 path 的子对象并返回， 如要查找 `CAA`, path 为 `/C/CA/CAA`
* `Equal(*JsonNode)`: 判断两个对象是否相等
* `ADD(key interface{}, value *JsonNode)`: 根据 key 向当前对象插入一个子对象
* `Replace(key interface{}, value *JsonNode)`: 根据 key 替换
* `Remove(key interface{})`: 根据 key shanc
* ...

`JsonNode` 对象可以使用 `Marshal()` 方法序列化成 JSON 字符数组

### diff

diff 遵循 RFC 6092 规范，两个 JSON 串的差异被分为 6 类：

1. `add`: 新增
2. `replace`: 替换
3. `remove`: 删除
4. `move`: 移动
5. `copy`: 复制
6. `test`: 测试

他们的格式如下：

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

> * `test` 用于还原时测试路径下的值是否与 value 相等

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
    res, _ := AsDiffs([]byte(json1), []byte(json2))
    fmt.Println(res)
}
```

使用 `AsDiffs()` 函数获取两个 JSON 串的差异，`AsDiffs()` 接收一组可选的参数 `JsonDiffOption` 他们的取值如下：

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

即默认情况下，差异串只有 `add, replace, remove` 三种， `remove` 也只会返回 path, 更改默认行为可以传入需要的 JsonDiffOption, 如：

```go
res, _ := AsDiffs([]byte(json1), []byte(json2), UseMoveOption, UseCopyOption, UseFullRemoveOption)
fmt.Println(res)
```

## 其他

什么样的两个 JSON 对象被认为是相等的：

* 对于一个对象 `{}`，顺序无关
* 对于一个列表 `[]`, 顺序相关

## 参考

[https://github.com/flipkart-incubator/zjsonpatch](https://github.com/flipkart-incubator/zjsonpatch)