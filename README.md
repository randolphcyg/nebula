# NEBULA - 网络流量分析平台

基于 Wails + Svelte + Wireshark 的现代化网络流量分析工具。

## 📋 目录

- [功能特性](#-功能特性)
- [技术栈](#-技术栈)
- [项目结构](#-项目结构)
- [快速开始](#-快速开始)
- [认证与权限](#-认证与权限)
- [核心组件](#-核心组件)
- [开发指南](#-开发指南)

## ✨ 功能特性

- 📂 **离线 PCAP 深度分析** - 调用 Wireshark HTTP 接口，支持大型流量包解析与分段检索
- ⚡ **网卡实时抓包** - 基于 libpcap 捕获网卡实时流量
- 🚗 **车载/工控协议专区** - 特定工业协议（如 CAN, Modbus）的时序与指令解析
- 🔐 **用户认证与权限控制** - 基于 JWT 的无状态认证和 RBAC 权限管理
- 🎨 **现代化 UI** - 响应式设计，支持暗色模式

## 🛠️ 技术栈

### 前端
- **框架**: Svelte 5
- **语言**: TypeScript
- **构建工具**: Vite
- **UI 组件**: 自研组件库
- **状态管理**: Svelte Stores

### 后端
- **框架**: Wails v2
- **语言**: Go
- **数据库**: SQLite
- **认证**: JWT (JSON Web Token)
- **协议分析**: Wireshark (tshark)

## 📁 项目结构

```
nebula/
├── frontend/                      # 前端代码
│   ├── src/
│   │   ├── components/            # 通用 UI 组件
│   │   │   ├── layout/           # 布局组件
│   │   │   │   ├── Sidebar.svelte
│   │   │   │   └── Header.svelte
│   │   │   └── ui/               # 基础 UI 组件
│   │   │       ├── Button.svelte
│   │   │       ├── Modal.svelte
│   │   │       ├── Toast.svelte
│   │   │       └── ToastContainer.svelte
│   │   │   ├── features/             # 业务功能模块
│   │   │   ├── analyzer/         # 流量分析模块
│   │   │   │   ├── components/   # 分析器组件
│   │   │   │   │   ├── JsonTree.svelte
│   │   │   │   │   └── StreamModal.svelte
│   │   │   │   └── pages/        # 分析器页面
│   │   │   │       ├── PcapList.svelte
│   │   │   │       ├── PcapDetail.svelte
│   │   │   │       └── InterfaceList.svelte
│   │   │   ├── user/             # 用户模块
│   │   │   │   └── pages/        # 用户页面
│   │   │   │       └── Profile.svelte
│   │   │   └── workspace/        # 工作区模块
│   │   │       └── Workspace.svelte
│   │   ├── pages/                # 页面级组件
│   │   │   ├── Dashboard.svelte
│   │   │   ├── Login.svelte
│   │   │   └── NotFound.svelte
│   │   ├── stores/               # 状态管理
│   │   │   ├── app.ts           # 应用状态
│   │   │   ├── auth.ts          # 认证状态
│   │   │   ├── toast.ts         # Toast 通知
│   │   │   ├── theme.ts         # 主题设置
│   │   │   └── preferences.ts   # 用户偏好
│   │   │   └── index.ts
│   │   ├── utils/                # 工具函数
│   │   │   ├── helpers.ts       # 通用工具（防抖、节流等）
│   │   │   └── index.ts
│   │   ├── wailsjs/              # Wails 自动生成
│   │   │   ├── go/               # Go 函数绑定
│   │   │   └── runtime/          # Wails 运行时
│   │   ├── App.svelte
│   │   └── main.ts
│   ├── index.html
│   └── package.json
├── internal/                      # 后端代码
│   ├── auth/                      # 认证模块
│   │   ├── service.go            # 认证服务核心逻辑
│   │   └── permission.go         # 权限检查器
│   ├── database/                  # 数据库模块
│   │   ├── database.go           # 数据库连接管理
│   │   └── migrate.go            # 数据库迁移和种子数据
│   ├── models/                    # 数据模型
│   │   ├── auth.go               # 用户、角色、权限模型
│   │   └── pcap.go               # PCAP 文件模型
│   ├── services/                  # 业务服务
│   │   ├── analyzer/             # 分析器服务
│   │   └── pcap/                 # PCAP 服务
│   └── main.go
├── wails.json
├── go.mod
└── README.md
```

## 🚀 快速开始

### 环境要求

- Node.js >= 18
- Go >= 1.21
- Wails >= 2.7
- Wireshark (tshark 命令可用)

### 安装依赖

```bash
# 安装前端依赖
cd frontend
npm install

# 安装后端依赖
cd ..
go mod tidy
```

### 配置说明

项目使用 `config.yaml` 文件进行配置，支持以下配置项：

```yaml
# Wireshark HTTP 服务配置
wireshark:
  base_url: "http://127.0.0.1:18090/api/v1"
  container_mount_path: "/app/pcaps/"
  timeout: 60

# 数据库配置
database:
  type: "sqlite"
  sqlite_path: "./nebula.db"

# JWT 认证配置
auth:
  secret_key: "nebula-secret-key-change-in-production"
  token_expiry: 24

# PCAP 文件存储配置
pcap:
  storage_path: "./pcaps"
  max_upload_size: 500
```

**重要配置说明：**
- `wireshark.base_url`: Wireshark HTTP API 地址，可根据实际部署环境修改
- `wireshark.container_mount_path`: Docker 容器内 PCAP 文件路径，使用 Docker 时需与容器挂载路径一致
- `auth.secret_key`: JWT 密钥，**生产环境务必修改为随机字符串**
- `database.sqlite_path`: SQLite 数据库文件路径

### 开发模式

```bash
# 启动开发服务器
wails dev
```

启动成功后会显示：
```
启动 NEBULA v1.0.0
```

### 生产构建

```bash
# 构建生产版本
wails build
```

构建完成后，可执行文件和配置文件位于 `build/bin` 目录。

### 自定义图标

应用图标和元信息已在 `wails.json` 中配置：

```json
{
  "name": "NEBULA",
  "info": {
    "companyName": "NEBULA",
    "productName": "NEBULA Network Analyzer",
    "productVersion": "1.0.0",
    "copyright": "Copyright © 2024"
  }
}
```

如需自定义图标，可将图标文件放在项目根目录，并在 `wails.json` 中添加：
```json
{
  "assetfs": {
    "icon": "appicon.png"
  }
}
```

## 🔐 认证与权限

### 架构说明

系统采用 JWT (JSON Web Token) 进行无状态认证，支持基于角色的权限控制 (RBAC)。

### 数据模型

#### User (用户)
- `ID`: 用户 ID
- `Username`: 用户名（唯一）
- `Password`: 加密后的密码（bcrypt）
- `Email`: 邮箱
- `RoleID`: 角色 ID
- `Status`: 状态（1: 正常，0: 禁用）
- `LastLogin`: 最后登录时间

#### Role (角色)
- `ID`: 角色 ID
- `Name`: 角色名称
- `Code`: 角色代码（唯一）
- `Description`: 角色描述
- `Permissions`: 权限列表（多对多）

#### Permission (权限)
- `ID`: 权限 ID
- `Name`: 权限名称
- `Code`: 权限代码（唯一）
- `Resource`: 资源类型（如：pcap, analyzer, user）
- `Action`: 操作类型（如：read, write, delete）

### 默认账户

系统初始化时会自动创建以下默认账户：

| 用户名 | 密码 | 角色 | 权限 |
|--------|------|------|------|
| admin | admin123 | 超级管理员 | 所有权限 |
| user | user123 | 普通用户 | 读取权限 |

### API 接口

#### 1. Login (用户登录)

**函数签名**: `Login(username, password string) (map[string]interface{}, error)`

**请求参数**:
- `username`: 用户名
- `password`: 密码

**响应数据**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-03-14T10:00:00Z",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@nebula.local",
    "role": "超级管理员",
    "roleCode": "admin"
  }
}
```

**使用示例**:
```javascript
import { Login } from '../wailsjs/go/main/App';

async function handleLogin() {
    try {
        const result = await Login('admin', 'admin123');
        localStorage.setItem('token', result.token);
        localStorage.setItem('user', JSON.stringify(result.user));
    } catch (error) {
        console.error('登录失败:', error);
    }
}
```

#### 2. ValidateToken (验证 Token)

**函数签名**: `ValidateToken(token string) (map[string]interface{}, error)`

**响应数据**:
```json
{
  "userID": 1,
  "username": "admin",
  "role": "超级管理员",
  "roleCode": "admin"
}
```

**使用示例**:
```javascript
import { ValidateToken } from '../wailsjs/go/main/App';

async function checkAuth() {
    const token = localStorage.getItem('token');
    if (!token) return;
    
    try {
        const user = await ValidateToken(token);
        console.log('当前用户:', user);
    } catch (error) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
    }
}
```

#### 3. CheckPermission (检查权限)

**函数签名**: `CheckPermission(token, resource, action string) (bool, error)`

**使用示例**:
```javascript
import { CheckPermission } from '../wailsjs/go/main/App';

async function canDeletePcap() {
    const token = localStorage.getItem('token');
    const hasPermission = await CheckPermission(token, 'pcap', 'delete');
    
    if (hasPermission) {
        showDeleteButton();
    } else {
        hideDeleteButton();
    }
}
```

## 🧩 核心组件

### 通用 UI 组件

#### Toast.svelte
Toast 通知组件，用于显示操作反馈。

```typescript
// 使用方式
import { success, error, warning, info } from '../stores/toast';

success('操作成功');
error('操作失败');
warning('警告信息');
info('提示信息');
```

**参数**:
- `message`: 提示信息
- `duration`: 显示时长（毫秒），默认 3000ms

#### Modal.svelte
模态框组件，支持自定义内容和宽度。

```svelte
<script>
    import Modal from '../components/ui/Modal.svelte';
    
    let showModal = false;
</script>

<Modal 
    bind:visible={showModal}
    title="标题"
    width="500px"
    showClose={true}
    on:close={() => showModal = false}
>
    <p>模态框内容</p>
</Modal>
```

### 业务组件

#### JsonTree.svelte
JSON 数据可视化组件，支持展开/折叠。

```svelte
<script>
    import JsonTree from '../features/analyzer/components/JsonTree.svelte';
    
    let jsonData = { key: 'value' };
</script>

<JsonTree data={jsonData} />
```

#### StreamModal.svelte
网络流追踪与重组模态框。

```svelte
<script>
    import StreamModal from '../features/analyzer/components/StreamModal.svelte';
    
    let show = false;
    let streamData = { /* ... */ };
</script>

<StreamModal 
    bind:show={show}
    title="Follow TCP Stream"
    data={streamData}
/>
```

### 页面组件

#### PcapList.svelte
PCAP 文件列表页面，支持：
- 文件导入（拖拽/选择）
- 批量删除
- 搜索过滤
- 分页显示

**事件**:
- `analyze`: 点击分析按钮时触发

#### PcapDetail.svelte
PCAP 文件详情页面，支持：
- 数据包列表
- 协议过滤
- 数据包详情查看
- 流追踪（Follow Stream）

**属性**:
- `file`: PCAP 文件对象

#### InterfaceList.svelte
网卡列表页面，支持：
- 网卡信息展示
- 搜索过滤
- 启动抓包

**事件**:
- `startCapture`: 点击开始抓包时触发

#### Profile.svelte
个人中心页面，支持：
- 查看个人信息（用户名、邮箱、角色）
- 修改密码
- 主题切换（亮色/暗色模式）
- 偏好设置（页面大小、通知开关、紧凑模式）

**功能**:
- 基本信息展示
- 密码修改表单验证
- 主题实时切换
- 偏好设置持久化（localStorage）

## 📖 开发指南

### 代码规范

#### 变量声明
```typescript
// ✅ 推荐：一行一个变量
let searchFileName = '';
let searchFileSize = '';

// ❌ 不推荐：多个变量一行
let searchFileName = '', searchFileSize = '';
```

#### 字符串
```typescript
// ✅ 推荐：使用单引号
const message = '操作成功';

// ❌ 不推荐：使用双引号
const message = "操作成功";
```

#### 导入顺序
```typescript
// 1. Svelte API
import { onMount } from 'svelte';

// 2. Wails Runtime
import { EventsOn } from '../../../../wailsjs/runtime';

// 3. Wails Go Functions
import { GetFileList } from '../../../../wailsjs/go/main/App';

// 4. Stores/Utils
import { showError } from '../../../stores/toast';
```

#### 错误处理
```typescript
// ✅ 推荐：使用 Toast 通知
try {
    await someAsyncOperation();
} catch (err) {
    showError('操作失败：' + err);
}

// ❌ 不推荐：使用 alert
try {
    await someAsyncOperation();
} catch (err) {
    alert('操作失败：' + err);
}
```

### 性能优化

#### 防抖搜索
```typescript
import { debounce } from '../../../utils/helpers';

const debouncedSearch = debounce(async () => {
    await refreshFileList();
}, 300);

// 使用
<input on:input={debouncedSearch} />
```

#### 分页加载
```typescript
let packetPage = 1;
let packetPageSize = 10;
let hasMore = true;

async function loadMore() {
    if (!hasMore) return;
    packetPage++;
    await loadAnalyzeData();
}
```

### 添加新功能

1. **创建功能模块**
```bash
mkdir -p frontend/src/features/new-feature/{components,pages}
```

2. **创建 Store**
```typescript
// frontend/src/stores/new-feature.ts
import { writable } from 'svelte/store';

export const featureData = writable([]);
```

3. **创建页面组件**
```svelte
<!-- frontend/src/features/new-feature/pages/ListPage.svelte -->
<script lang="ts">
    import { onMount } from 'svelte';
    
    export let title = '';
</script>

<div class="page-container">
    <h1>{title}</h1>
</div>
```

4. **注册路由**
```svelte
// frontend/src/App.svelte
import NewFeaturePage from './features/new-feature/pages/ListPage.svelte';
```

## 📝 更新日志

### v1.0.0 (2024-03-13)
- ✨ 初始版本发布
- 🎨 完成代码规范统一
- ⚡ 性能优化（防抖、分页）
- 🔐 完善认证与权限系统

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
