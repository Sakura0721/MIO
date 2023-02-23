# go-mio-scripts

mioGameBot的自动化脚本。
代码写的不好，自己用的。

## Features

- 送心给指定的用户
- 自动给当天送过心的用户回赠
- 给指定用户增加时间
- 自动给八小时内为你增加过时间的用户增加时间
- 为公开列表中前五十个用户增加时间

## How to use

需要golang环境，开发所使用的go版本为`1.20`

```bash
git clone https://github.com/Sakura0721/MIO.git
cd MIO
cp config.example.yaml config.yaml
```

根据 `config.yaml` 中的注释填入字段，设置`enabled`字段。

```bash
go run main.go
```

第一次运行的时候可能会需要验证码，手动输入即可。

第一次运行之后可设置crontab，每小时执行一次：
```
crontab -e
# 0 * * * * go run main.go >> log.txt
```

