# Entry

[![Build Status](https://travis-ci.org/laincloud/entry.svg?branch=master)](https://travis-ci.org/laincloud/entry)
[![MIT license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)

## Documentation

相关文档见 [Entry 应用文档](https://laincloud.gitbooks.io/white-paper/content/outofbox/entry.html)

## Licensing
Entry is released under [MIT](https://github.com/laincloud/entry/blob/master/LICENSE) license.

## 部署

### 配置

请参考 [example.json](example.json) 编写配置文件，并上传到 lvault：

```
lain secret add ${LAIN-Domain} web /lain/app/prod.json -f example.json
```

> - `smtp.address` 需要包含端口，如：${mail-address}:25
> - `smtp.password` 可选，为空时不使用 auth

## 审计

## 开发

### 由 `swagger.yml` 生成代码

```
go get -u github.com/go-swagger/go-swagger/cmd/swagger  # 安装 swagger
swagger generate server -f ./swagger.yml -t server/gen  # 生成代码
```

- `server/gen` 下除 `server/gen/restapi/configure_entry.go` 外均由 `go-swagger` 生成，请不要手动修改
- `server/gen/restapi/configure_entry.go` 包含初始化逻辑以及后端 API 配置
- `server/handler` 包含后端 API 的实际逻辑
