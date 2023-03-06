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

- 需要golang环境，开发所使用的go版本为`1.20`
- 需要安装版本大于等于1.8.11的TDLib
  - [编译安装](https://github.com/tdlib/td#building)，需要注意不要使用release中的源码，clone master或者1.8.11的commit
  - 如果使用ArchLinux，可以从AUR安装 [telegram-tdlib-git](https://aur.archlinux.org/packages/telegram-tdlib-git)


```bash
git clone https://github.com/Sakura0721/MIO.git
cd MIO
cp config.example.yaml config.yaml
```

根据 `config.yaml` 中的注释填入字段，设置`enabled`字段。

```bash
go build main.go
./main
```

第一次运行的时候可能会需要验证码，手动输入即可。

第一次运行之后可设置crontab，每小时执行一次：
```
crontab -e
# 0 * * * * cd /path/to/MIO && ./main>> log.txt
```

## Developing

如果想抓取mio的API的话有两种方法：

1. 电脑tg打开mio，使用fiddler抓包；
2. 把本项目跑起来，在第一次输出登录链接的时候迅速点击打开，使用chrome的开发者工具抓API；
