## PoetBot

这是一个简单的 Telegram 机器人，该机器人使用 go 编写并通过 tdlib 构建。

### 它能干什么？

机器人默认每隔 30s 更新一次用户昵称，昵称将从 `poet.txt` 中随机选取一行(30s 自动换昵称)。

### 如何使用？

#### 网络环境要求

首先由于 Telegram 的特殊性，部署该 Bot 的机器应该能够无障碍访问 Telegram 服务器(代理支持懒得写了)。

#### API ID 申请

在使用本机器人之前请确保已经申请了自己的 `Api ID` 和 `APi Hash`，如果没有请参考 [https://core.telegram.org/api/obtaining_api_id](https://core.telegram.org/api/obtaining_api_id) 自行创建。

#### Docker Compose(推荐)

由于 tdlib 需要交互式输入手机号等信息，所以在启动之前请先执行一次手动运行的初始化动作来生成数据文件

```yaml
version: '3.8'
services:
  caddy:
    image: mritd/poetbot
    container_name: poetbot
    restart: always
    volumes:
      - /etc/timezone:/etc/timezone
      - data:/data
    environment:
      - POETBOT_APPID=xxxx
      - POETBOT_APPHASH=xxxxxxxxxxxxxxx
    # 首次运行时请替换为该 entrypoint，并手动 exec 到容器内启动
    # 然后交互式输入手机号、验证码等信息
    #entrypoint: ["tail","-f","/dev/sdtout"]
volumes:
  data:
```

#### 二进制运行

```sh
./poetbot --apiid xxxxxx --apphash xxxxxxxxxxxxxxxxxxxxxxxxxxx --txtfile ./poet.txt
```

**默认情况下可以直接使用本项目下的 `poet.txt`，用户也可以自行创建这个昵称文本文件。**

#### 自行编译

编译前请确保按照 [TDLib build instructions generator](https://tdlib.github.io/td/build.html) 安装好 tdlib 依赖；编译环境需要 go 1.17+，
然后执行 `go build` 命令编译即可。

**Dockerfile 目前已经切换到 [Earthfile](https://github.com/mritd/autobuild/blob/main/earthfiles/poetbot/Earthfile)，关于如何使用 [Earthly](https://docs.earthly.dev/) 请自行参考官方文档。**
