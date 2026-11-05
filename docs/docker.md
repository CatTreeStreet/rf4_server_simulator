# Docker 部署

此部署方式在一个容器内运行 Linux 版 seed、gameserver 和 gateway，适合 Linux VPS 或支持 Docker Compose 的 NAS 进行本地/受控网络测试。它不包含 Windows 系统托盘 launcher，也不提供 launcher 内嵌的 Web 管理控制台（9500）；管理后台不是独立容器服务。

## 1. 准备配置

执行：cp config.example.yaml config.yaml

编辑 config.yaml，至少修改以下配置：

- gateway.advertise_host：填写客户端能够访问此服务器的 IP 或域名，不能保留默认的 127.0.0.1（除非客户端与服务端在同一网络命名空间）。
- gateway.tunnel.psk 与 game.tunnel.psk：设置为相同的高强度随机密钥。
- admin.jwt_secret、admin.password：替换默认占位值。不要将配置文件提交到 Git。
- gateway.login_addr 和 gateway.realtime_addr：保持容器内监听端口分别为 :443 和 :9341。
- gateway.obs_addr / game.obs_addr：保持默认的 :9501 / :9502，供容器健康检查使用。

证书会在首次启动时生成并保存在 Docker 命名卷 rf4-certs 中；数据库保存在 rf4-data 中。

## 2. 构建并启动

执行：docker compose up -d --build
查看状态：docker compose ps
查看日志：docker compose logs -f rf4-server

容器启动时会先幂等执行数据库初始化，再启动 gameserver，等待其就绪后启动 gateway。任一服务退出，容器会停止并由 Docker 的重启策略拉起。

默认端口映射：443/TCP 用于 HTTPS 登录，9341/TCP 用于实时游戏连接。

如果主机端口冲突，可通过环境变量修改主机侧端口，例如：
RF4_LOGIN_PORT=8443 RF4_REALTIME_PORT=19341 docker compose up -d

注意：改变主机端口后，客户端接入配置/登录响应必须与实际对外端口保持一致；仅改变端口映射不会自动修改游戏协议中公布的端口。

## 3. 更新

执行：docker compose down
执行：docker compose build --pull
执行：docker compose up -d

不要执行 docker compose down -v，除非确认要永久删除数据库和证书卷。

## 4. 重要安全边界

- 此项目是客户端协议与本地服务端模拟研究，不是完整官服替代品。请只在你有权测试的客户端和隔离网络中使用，并遵守游戏服务条款及当地法律。
- 当前项目文档明确指出登录流程不认证真实游戏账户，且实时连接的 token 校验并非完整账户认证。不要将端口直接暴露到公网。
- 内部隧道和观测端口不映射到宿主机；不要额外将 9401、9501、9502 暴露到公网。
- 不要分享 config.yaml、certs/、数据库或私钥。
- Docker 版本没有 Windows launcher 的托盘、控制中心和游戏启动功能。

## 5. 故障排查

查看状态：docker compose ps
查看日志：docker compose logs --tail=200 rf4-server
查看健康状态：docker inspect --format '{{json .State.Health}}' rf4-server

若健康检查失败，先检查日志、配置中的隧道 PSK 是否一致、静态数据库文件是否存在，以及 443/9341 是否已被宿主机其他服务占用。
