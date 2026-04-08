# Skill: Connect to Tencent Cloud Server（与 Vino_test 同机）

## Connection Details
- **IP**: 106.54.50.88
- **Port**: 22
- **User**: ubuntu
- **Key**: `F:/ItsyourTurnMy/backend/deploy/test.pem`

## SSH Command
```bash
ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -i F:/ItsyourTurnMy/backend/deploy/test.pem ubuntu@106.54.50.88
```

## Execute Remote Command
```bash
ssh -o HostKeyAlgorithms=+ssh-rsa -o PubkeyAcceptedKeyTypes=+ssh-rsa -o StrictHostKeyChecking=no -i F:/ItsyourTurnMy/backend/deploy/test.pem ubuntu@106.54.50.88 "<command>"
```

## Project Location on Server
- **xyKitchen 路径**: `/home/ubuntu/xyKitchen`
- 与 `Vino_test`（`/home/ubuntu/Vino_test`）并存时：**不同目录、不同端口**，互不影响。

## Port Allocation（避免冲突）
| 项目 | API 端口 | 说明 |
|------|----------|------|
| Vino_test | 5202 | Docker vino-backend |
| **xyKitchen** | **5402** | **本仓库 Go 二进制直接监听（非 Docker）** |

MySQL：xyKitchen 默认连接配置见 `backend/internal/config.go`（`DB_PORT` 默认 **3311**、`DB_NAME` 默认 **xykitchen_db**）；请使用独立库名/端口，避免与 vino-mysql(3308) 等业务库混用。
