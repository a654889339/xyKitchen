---
name: xykitchen-deploy
description: 提交 xyKitchen 到 Git、在 106 服务器 git pull，并仅重启本项目的 Go 后端（5402）与 Node 网页前台（5401）；不使用 Docker Compose、不重启 Docker 守护进程、不影响其他端口进程。在用户要求部署、发布、更新 xyKitchen 时使用。
---

# xyKitchen 部署

## 硬约束

- **禁止** `docker compose`、`systemctl restart docker` 或任何影响其他容器/端口的操作。
- 只动 **API 5402**（Go）和 **网页 5401**（Node Express），数据库和其他服务不碰。

## 环境说明

| 项目 | 值 |
|------|----|
| 本机 Shell | **PowerShell**（不支持 `&&`，用 `;` 分隔命令） |
| 本机仓库路径 | `F:\xyKitchen` |
| SSH 密钥 | `F:/ItsyourTurnMy/backend/deploy/test.pem` |
| 服务器 | `ubuntu@106.54.50.88` |
| 服务器仓库 | `/home/ubuntu/xyKitchen` |
| SSH 选项 | `-o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no` |

## 步骤

### 1. 本机：提交并推送

所有 Shell 命令用 `Set-Location` 而非 `cd`，用 `;` 而非 `&&`。

```powershell
Set-Location F:\xyKitchen; git add -A; git status
```

确认无敏感文件（如 `.env` 含密钥）后：

```powershell
Set-Location F:\xyKitchen; git commit -m "<简述改动>"
```

```powershell
Set-Location F:\xyKitchen; git push origin main
```

> `git push` 需要 `required_permissions: ["full_network"]`。

### 2. 服务器：拉取 + 部署

SSH 到服务器执行一行命令；`required_permissions: ["full_network"]`，`block_until_ms: 120000`（编译耗时可能较长）。

```powershell
ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -i F:/ItsyourTurnMy/backend/deploy/test.pem ubuntu@106.54.50.88 "cd /home/ubuntu/xyKitchen && git fetch origin && git reset --hard origin/main && chmod +x deploy/restart-backend.sh deploy/restart-frontend.sh && ./deploy/restart-backend.sh && ./deploy/restart-frontend.sh"
```

- `restart-backend.sh`：编译 Go（或 Docker 一次性编译）→ 杀旧进程 → source `.env` → 启动 → 重试 health check（最多 15s）。
- `restart-frontend.sh`：`npm ci`（首次）→ 杀旧进程 → 启动 Node → 重试 health check（最多 10s）。

#### 若后端 health check 失败

1. 检查日志：

```powershell
ssh ... ubuntu@106.54.50.88 "tail -50 /home/ubuntu/xyKitchen/logs/backend.log"
```

2. 常见原因：数据库连接失败（检查 `backend/.env`）、端口占用、编译错误。
3. 后端失败会中断前端部署，需单独执行前端脚本：

```powershell
ssh ... ubuntu@106.54.50.88 "cd /home/ubuntu/xyKitchen && ./deploy/restart-frontend.sh"
```

### 3. 验证

```powershell
ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -i F:/ItsyourTurnMy/backend/deploy/test.pem ubuntu@106.54.50.88 "curl -sf http://127.0.0.1:5402/api/health; echo; curl -sf -o /dev/null -w '%{http_code}' http://127.0.0.1:5401/"
```

期望输出：`{"code":0,...}` + `200`。

对外地址（安全组放行后）：
- API / 管理后台：`http://106.54.50.88:5402/`
- 网页首页：`http://106.54.50.88:5401/`
