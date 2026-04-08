---
name: xykitchen-deploy
description: 提交 xyKitchen 到 Git、在 106 服务器 git pull，并仅重启本项目的 Go 后端（5402）与 Node 网页前台（5401）；不使用 Docker Compose、不重启 Docker 守护进程、不影响其他端口进程。在用户要求部署、发布、更新 xyKitchen 时使用。
---

# xyKitchen 部署（无 Docker Compose）

## 约束

- **禁止**执行 `docker compose`、`systemctl restart docker` 或任何会重启 Docker 守护进程/其他项目容器的操作。
- **仅**更新 xyKitchen：**API 5402**（Go）、**网页 5401**（Node 静态服务）；数据库与其他服务保持不动。
- SSH 与密钥路径见 **`.cursor/skills/connectToTxCloud.md`**。

## 前置

- 服务器已 clone 本仓库到 `/home/ubuntu/xyKitchen`（可先 `git push` 再服务器 `git pull`）。
- **后端**：宿主机可无全局 Go；`deploy/restart-backend.sh` 可本机 `go build`，否则用 **一次性** `docker run golang` 编译（`GOPROXY=https://goproxy.cn,direct`），不重启现有业务容器。
- **前端**：需 **Node.js ≥18** 与 `npm`（`frontend-web` 使用 Express）；首次在 `frontend-web` 执行 `npm ci` 或 `npm install`。
- **MySQL**：库 `xykitchen_db` 等配置见 `backend/.env`；`restart-backend.sh` 会 **`source backend/.env`** 再启动进程。

## 本机：提交并推送

在仓库根目录（如 `F:/xyKitchen`）：

```bash
git add -A
git status
git commit -m "<说明本次改动的完整句子>"
git push origin main
```

## 服务器：拉取并更新（后端 + 前端）

一键（密钥路径按本机实际 `.pem` 修改）：

```bash
ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -i F:/ItsyourTurnMy/backend/deploy/test.pem ubuntu@106.54.50.88 "cd /home/ubuntu/xyKitchen && git fetch origin && git reset --hard origin/main && chmod +x deploy/restart-backend.sh deploy/restart-frontend.sh && ./deploy/restart-backend.sh && ./deploy/restart-frontend.sh"
```

- **`deploy/restart-backend.sh`**：编译并重启监听 **5402** 的 `xykitchen-server`，`curl` 校验 `/api/health`。
- **`deploy/restart-frontend.sh`**：`cd frontend-web`，`npm ci`（无 lock 时 `npm install`），重启监听 **5401** 的 `node server.js`，校验根路径可访问。

若仅更新后端或仅更新前端，可单独执行对应脚本。

## 可选：systemd（首次）

```bash
sudo cp /home/ubuntu/xyKitchen/deploy/xykitchen.service /etc/systemd/system/
sudo cp /home/ubuntu/xyKitchen/deploy/xykitchen-frontend.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now xykitchen xykitchen-frontend
```

日常更新仍可用 **脚本** 或 **`systemctl restart xykitchen`** / **`systemctl restart xykitchen-frontend`**，勿混用两套启动方式 unless 已停掉另一种。

## 验证

```bash
ssh ... ubuntu@106.54.50.88 "curl -s http://127.0.0.1:5402/api/health"
ssh ... ubuntu@106.54.50.88 "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:5401/"
```

对外（安全组放行后）：API `http://106.54.50.88:5402/`、管理页、网页首页 `http://106.54.50.88:5401/`。
