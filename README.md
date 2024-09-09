# lark-doc
quick search for lark doc

<img width="718" alt="image" src="https://user-images.githubusercontent.com/63107263/209918773-731dfbb4-d5bc-4abe-91d1-4fa479db7c4b.png">

## Usage
![output](assets/output.gif)

## Config
1. `count` search count; default 9
2. `cache_file` cache file path; default `lark.json`
3. `session` get after login, you need to fill it!!


## Sample

```shell
# search realtime
session=xxx count=10 go run main.go query "hello"

# update cache
session=xxx count=10 cache_file=lark.json go run main.go trigger
```
