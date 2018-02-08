# Entry

[![Build Status](https://travis-ci.org/laincloud/entry.svg?branch=master)](https://travis-ci.org/laincloud/entry)
[![MIT license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)

## Documentation

相关文档见 [Entry 应用文档](https://laincloud.gitbooks.io/white-paper/content/outofbox/entry.html)

## Licensing
Entry is released under [MIT](https://github.com/laincloud/entry/blob/master/LICENSE) license.

## 审计

## 开发

### 由 `swagger.yml` 生成代码

```
go get -u github.com/go-swagger/go-swagger/cmd/swagger  # 安装 swagger
swagger generate server -f ./swagger.yml -t server/gen  # 生成代码
```
