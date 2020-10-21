# JSON lint

Simple JSON linter. There are no advantages over any other linters. Made just for fun. 

## Usage

Good JSON:
```go
src := `{"name":"John","age":31,"city":"New York"}`
_, err := jsonlint.ValidateStr(src)
assertNil(err)
```

Bad JSON:
```go
src := `{"id":1,"name":"Foo","price":123,"tags":["Bar",,"Eek"],"stock":{"warehouse":300,"retail":20}}`
offset, err := jsonlint.ValidateStr(src)
assertEqual(err, jsonlint.ErrUnexpId)
assertEqual(offset, 47)
```
