# Zeek Runner 服务集成指南

## 📋 概述

本项目已集成 `zeek_runner` 服务的 gRPC 和 HTTP 接口，提供 PCAP 流量包的 Zeek 分析能力。

## 🏗️ 架构设计

```
Nebula Backend
    ├── gRPC Client (优先)
    │   └── localhost:50051
    └── HTTP Client (降级)
        └── localhost:8080
            └── zeek_runner
                ├── /api/v1/analyze (POST)
                ├── /api/v1/healthz (GET)
                └── /api/v1/version/zeek (GET)
```

## 📦 目录结构

```
internal/services/zeek/
├── pb/                          # 自动生成的 protobuf 代码
│   ├── zeek_runner.pb.go       # 消息定义
│   └── zeek_runner_grpc.pb.go  # gRPC 服务定义
├── client.go                    # gRPC & HTTP 客户端封装
├── http_client.go               # HTTP 客户端封装
├── service.go                   # 业务服务层
└── README.md                    # 本文档
```

## 🚀 快速开始

### 1. 安装依赖

```bash
cd /Users/randolph/goodjob/nebula
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

### 2. 配置服务

编辑 `config.yaml` 文件，添加 Zeek Runner 配置：

```yaml
zeek_runner:
  grpc_address: "localhost:50051"
  http_address: "http://localhost:8080"
  timeout: 30
  enabled: true
  
  # 重试配置（可选）
  retry:
    max_attempts: 3
    backoff_base_delay: 1
    backoff_max_delay: 10
  
  # 连接池配置（可选）
  pool:
    max_idle_conns: 10
    max_open_conns: 20
    conn_max_lifetime: 1800
```

**注意**：服务会在应用启动时自动初始化，无需手动调用。

## 💡 使用示例

### Go 代码调用

```go
import (
    "context"
    "nebula/internal/services/zeek"
)

// 分析 PCAP
func analyzeFile(service *zeek.Service) error {
    req := zeek.AnalyzePCAPRequest{
        TaskID:               "task-001",
        UUID:                 "user-uuid",
        OnlyNotice:           false,
        PcapID:               "pcap-123",
        PcapPath:             "/data/sample.pcap",
        ScriptID:             "",
        ScriptPath:           "",
        ExtractedFilePath:    "/tmp/extracted",
        ExtractedFileMinSize: 1024, // KB
    }

    result, err := service.AnalyzePCAP(context.Background(), req)
    if err != nil {
        return err
    }

    fmt.Printf("分析任务已启动：%s\n", result.TaskID)
    return nil
}

// 检查服务健康状态
func checkHealth(service *zeek.Service) error {
    health, err := service.CheckHealth(context.Background())
    if err != nil {
        return err
    }

    if !health.IsHealthy() {
        return fmt.Errorf("服务异常：%s", health.GetStatusMessage())
    }

    fmt.Println("服务正常")
    return nil
}
```

### 前端调用（TypeScript）

```typescript
// 首先需要 Wails 生成绑定
// wails generate

import { AnalyzePCAP, CheckZeekHealth, GetZeekVersion } from '../wailsjs/go/main/App';

async function handleAnalyze() {
    try {
        // 1. 检查服务健康
        const health = await CheckZeekHealth();
        if (health.status !== 'ok') {
            alert('Zeek 服务不可用：' + health.status);
            return;
        }

        // 2. 发起分析任务
        const result = await AnalyzePCAP({
            taskID: 'task-001',
            uuid: 'user-uuid',
            onlyNotice: false,
            pcapID: 'pcap-123',
            pcapPath: '/data/sample.pcap',
            scriptID: '',
            scriptPath: '',
            extractedFilePath: '/tmp/extracted',
            extractedFileMinSize: 1024,
        });

        console.log('分析任务已启动:', result);
        alert('分析成功！');
    } catch (error) {
        console.error('分析失败:', error);
        alert('分析失败：' + error.message);
    }
}
```

## 🔧 API 参考

### AnalyzePCAPRequest

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| TaskID | string | ✅ | 任务唯一标识 |
| UUID | string | ✅ | 用户 ID |
| OnlyNotice | bool | ❌ | 是否只生成 notice 日志 |
| PcapID | string | ❌ | PCAP 文件 ID |
| PcapPath | string | ✅ | PCAP 文件路径 |
| ScriptID | string | ❌ | Zeek 脚本 ID |
| ScriptPath | string | ❌ | Zeek 脚本路径 |
| ExtractedFilePath | string | ❌ | 提取文件存储路径 |
| ExtractedFileMinSize | int32 | ❌ | 提取文件最小大小 (KB) |

### AnalyzeResult

| 字段 | 类型 | 说明 |
|------|------|------|
| TaskID | string | 任务 ID |
| Status | string | 任务状态（running/completed/failed） |
| StartTime | string | 任务开始时间 |

### HealthStatus

| 字段 | 类型 | 说明 |
|------|------|------|
| Status | string | 状态（ok/kafka_down） |
| PoolRunning | int | 协程池运行数量 |

## 🛠️ 高级功能

### 1. gRPC 失败自动降级

服务层实现了自动降级机制：

```go
func (s *Service) AnalyzePCAP(ctx context.Context, req AnalyzePCAPRequest) (*AnalyzeResult, error) {
    // 优先使用 gRPC
    resp, err := s.client.Analyze(ctx, grpcReq)
    if err != nil {
        // gRPC 失败时自动降级到 HTTP
        log.Warn("gRPC failed, fallback to HTTP", "error", err)
        return s.analyzeHTTP(ctx, req)
    }
    return resp, nil
}
```

### 2. 健康检查

定期检查服务状态：

```go
func monitorZeekService(service *zeek.Service) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        health, err := service.CheckHealth(context.Background())
        if err != nil {
            log.Error("Zeek health check failed", "error", err)
            continue
        }

        if !health.IsHealthy() {
            log.Warn("Zeek service unhealthy", "status", health.Status)
        }
    }
}
```

### 3. 超时控制

所有调用都设置了合理的超时：

```go
// gRPC 调用默认 5 分钟超时
ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
defer cancel()
```

## 📝 注意事项

1. **Proto 文件修改**
   - 修改 `proto/zeek_runner.proto` 后需重新生成代码
   - 使用命令：`protoc --go_out=. --go-grpc_out=. proto/zeek_runner.proto`

2. **服务依赖**
   - 确保 zeek_runner 服务已启动
   - 检查 Kafka 连接状态

3. **错误处理**
   - 所有公开方法都返回 error
   - 建议调用方进行适当的错误处理

4. **资源释放**
   - 使用 `defer zeekService.Close()` 确保连接关闭

## 🔍 故障排查

### 问题 1: gRPC 连接失败

```bash
# 检查 zeek_runner 服务是否运行
curl http://localhost:8080/api/v1/healthz

# 检查 gRPC 端口
netstat -tlnp | grep 50051
```

### 问题 2: 代码生成失败

```bash
# 确保 protoc 已安装
protoc --version

# 确保插件已安装
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 问题 3: 导入路径错误

确保 `go.mod` 中的 module 名称正确：

```go
module nebula
```

## 📚 相关文档

- [gRPC Go 官方文档](https://grpc.io/docs/languages/go/)
- [Protocol Buffers 文档](https://protobuf.dev/)
- [zeek_runner 服务文档](https://github.com/your-org/zeek_runner)

## 🤝 贡献

如有问题或改进建议，请提交 Issue 或 Pull Request。
