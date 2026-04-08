---
name: xykitchen-deploy
description: 提交 xyKitchen 代码到 Git、在 106 服务器 git pull，并用脚本仅重启监听 5402 的 Go 后端进程；不使用 Docker、不重启 Docker 守护进程、不影响其他端口进程。在用户要求部署、发布、更新 xyKitchen 后台时使用。
---

# xyKitchen 部署（无 Docker）

## 约束

- **禁止**执行 `docker compose`、`systemctl restart docker` 或任何会重启 Docker 守护进程/其他项目容器的操作。
- **仅**更新 xyKitchen：监听 **5402** 的后端进程；数据库与其他服务保持不动。
- SSH 与密钥路径见项目内 **`.cursor/skills/connectToTxCloud.md`**。

## 前置

- 服务器已 `git clone` 本仓库到 `/home/ubuntu/xyKitchen`（若仅本机有代码，需先 `git push` 到远程，服务器再能 `git pull`）。
- 宿主机可无全局 Go：`deploy/restart-backend.sh` 会优先用本机 `go`，否则用 **一次性** `docker run golang` 编译（拉模块时使用 `GOPROXY=https://goproxy.cn,direct`，不重启现有业务容器）。
- **MySQL**：在实例中创建库 `xykitchen_db`（可与现有 `vino-mysql` 共用实例、**不同库名**，宿主机端口一般为 **3308**）。将账号写入 `backend/.env`（参考 `backend/.env.example`）。`restart-backend.sh` 会在启动进程前 **`source backend/.env`**，使 `DB_PORT` 等生效。

## 本机：提交并推送

在仓库根目录 `F:/xyKitchen`：

```bash
git add -A
git status
git commit -m "<说明本次改动的完整句子>"
git push origin main
```

若无 `origin`，先添加远程再推送。

## 服务器：拉取并重启后端（仅 5402）

一键（将密钥路径换成本机实际 `.pem`）：

```bash
ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -i F:/ItsyourTurnMy/backend/deploy/test.pem ubuntu@106.54.50.88 "cd /home/ubuntu/xyKitchen && git pull && chmod +x deploy/restart-backend.sh && ./deploy/restart-backend.sh"
```

- `deploy/restart-backend.sh` 会在 `backend/` 下执行 `go build -o xykitchen-server ./cmd/server`，结束占用 **5402** 的旧进程后启动新二进制，并请求 `http://127.0.0.1:5402/api/health` 校验。

## 可选：使用 systemd（首次配置）

首次可安装 unit（仅需一次，之后可用 `sudo systemctl restart xykitchen` 代替脚本内 nohup；仍不碰 Docker）：

```bash
sudo cp /home/ubuntu/xyKitchen/deploy/xykitchen.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now xykitchen
```

日常更新仍建议 **`git pull` + `go build` + `systemctl restart xykitchen`**，与脚本二选一，勿混用两套启动方式 unless 已手动停掉另一种。

## 验证

```bash
ssh ... ubuntu@106.54.50.88 "curl -s http://127.0.0.1:5402/api/health"
```

对外访问（若安全组已放行）：`http://106.54.50.88:5402/`（管理页 `admin.html`）、`http://106.54.50.88:5402/api/health`。
