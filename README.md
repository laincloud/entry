# Entry

[![Build Status](https://travis-ci.org/laincloud/entry.svg?branch=master)](https://travis-ci.org/laincloud/entry)
[![MIT license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)

## Documentation

相关文档见 [Entry 应用文档](https://laincloud.gitbooks.io/white-paper/content/outofbox/entry.html)

## Licensing
Entry is released under [MIT](https://github.com/laincloud/entry/blob/master/LICENSE) license.

## 打包上传到 PyPI

### 依赖

```
pip install twine  # 上传工具
pip install wheel  # 打包工具
```

### 打包上传

```
rm -rf dist/  # 清空以前的构建
python setup.py sdist  # 打包源代码
python setup.py bdist_wheel  # 构建 wheel
twine upload dist/*  # 上传
```
