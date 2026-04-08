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
| 项目 | 端口 | 说明 |
|------|------|------|
| Vino_test 主站 | 5201 | Docker 前端 |
| Vino_test API | 5202 | Docker vino-backend |
| **xyKitchen 网页** | **5401** | **Node `frontend-web`（首页动画，与小程序一致）** |
| **xyKitchen API** | **5402** | **Go 二进制（非 Docker Compose）** |

MySQL：xyKitchen 使用独立库名（如 `xykitchen_db`）；可与 `vino-mysql` 宿主机 **3308** 共用实例，账号写在 `backend/.env`。
