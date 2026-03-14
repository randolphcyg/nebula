<script lang="ts">
    import { onMount } from 'svelte';
    import { auth, app } from './stores';
    import { preferencesStore } from './stores/preferences';
    import { showSuccess } from './utils';
    import { logger } from './utils/logger';
    import { ToastContainer } from './components/ui';
    import { Sidebar } from './components/layout';
    import Login from './auth/Login.svelte';
    import Register from './auth/Register.svelte';
    import Workspace from './features/workspace/Workspace.svelte';
    import PcapList from './features/analyzer/pages/PcapList.svelte';
    import UserManagement from './features/admin/pages/UserManagement.svelte';
    import Profile from './features/user/pages/Profile.svelte';
    import { GetFileList, GetWiresharkVersion, GetZeekVersions, IsZeekEnabled } from '../wailsjs/go/main/App';
    
    let isAuthenticated = false;
    let activeTab = 'home';
    let showRegisterPage = false;
    let user = null;
    let compactMode = false;
    
    // 订阅紧凑模式
    onMount(() => {
        preferencesStore.subscribe(prefs => {
            compactMode = prefs.compactMode;
        });
    });
    
    // ==== 大屏统计状态 ====
    let dashStats = {
        pcapCount: 0,
        wsVersion: "探测中...",
        zeekVersion: "待检测",
        zeekKafkaVersion: "待检测",
        zeekScripts: 142,
        aiModel: "待配置",
        aiTokens: "1.2M",
        systemUptime: "0 Days"
    };
    
    // 监听认证状态
    auth.subscribe(state => {
        isAuthenticated = state.isAuthenticated;
        user = state.user;
    });
    
    // 监听标签页切换
    app.subscribe(state => {
        activeTab = state.activeTab;
    });
    
    onMount(async () => {
        // 如果已登录，加载数据
        if (isAuthenticated) {
            await loadDashboardData();
        }
    });
    
    async function loadDashboardData() {
        // 1. 获取 PCAP 文件总数
        try {
            const pcapResp = await GetFileList({ fileName: '', fileSize: '', startDate: '', endDate: '', page: 1, pageSize: 1 });
            dashStats.pcapCount = pcapResp.total || 0;
        } catch(e) {
            logger.error("大屏获取文件数失败:", e);
        }

        // 2. 获取 wireshark 引擎版本
        try {
            const wsVersion = await GetWiresharkVersion();
            // GetWiresharkVersion 直接返回版本字符串，不是 JSON
            dashStats.wsVersion = wsVersion || "版本未知";
        } catch(e) {
            dashStats.wsVersion = "引擎离线";
            logger.error("大屏获取 Wireshark 版本失败:", e);
        }
        
        // 3. 获取 Zeek 引擎和 Zeek-Kafka 版本
        try {
            const isEnabled = await IsZeekEnabled();
            logger.debug("Zeek 服务是否启用:", isEnabled);
            
            if (isEnabled) {
                const zeekData = await GetZeekVersions();
                logger.debug("Zeek 版本数据:", zeekData);
                
                // 确保字段存在且不是 HTML
                const zeekVer = zeekData.zeek_version;
                const zeekKafkaVer = zeekData.zeek_kafka_version;
                
                // 检查是否是 HTML 响应（错误情况）
                if (typeof zeekVer === 'string' && zeekVer.startsWith('<!doctype')) {
                    logger.error("Zeek 版本返回了 HTML 而不是纯文本");
                    dashStats.zeekVersion = "版本获取失败";
                    dashStats.zeekKafkaVersion = "-";
                } else {
                    dashStats.zeekVersion = zeekVer || "版本未知";
                    dashStats.zeekKafkaVersion = zeekKafkaVer || "版本未知";
                }
            } else {
                dashStats.zeekVersion = "服务未启用";
                dashStats.zeekKafkaVersion = "-";
            }
        } catch(e) {
            dashStats.zeekVersion = "引擎离线";
            dashStats.zeekKafkaVersion = "-";
            logger.error("大屏获取 Zeek 版本失败:", e);
        }
    }
    
    function handleLoginSuccess() {
        showSuccess('欢迎回来！');
        loadDashboardData();
    }
    
    function handleTabChange(event: CustomEvent) {
        activeTab = event.detail;
        if (activeTab === 'home') {
            loadDashboardData();
        }
    }

    function handlePcapAnalyze(event: CustomEvent) {
        const file = event.detail;
        app.setActiveTab('analyzer');
        // 触发 Workspace 的 analyze 事件
        setTimeout(() => {
            const workspaceEvent = new CustomEvent('analyze', { detail: file });
            window.dispatchEvent(workspaceEvent);
        }, 100);
    }
</script>

<div id="app-wrapper" class:compact={compactMode}>
    <ToastContainer />
    
    {#if !isAuthenticated}
        <!-- 登录/注册页面 -->
        {#if showRegisterPage}
            <Register on:switchToLogin={() => showRegisterPage = false} />
        {:else}
            <Login on:success={handleLoginSuccess} on:showRegister={() => showRegisterPage = true} />
        {/if}
    {:else}
        <!-- 主应用布局 -->
        <Sidebar on:tabChange={handleTabChange}>
            {#if activeTab === 'home'}
                <!-- 控制台首页 -->
                <div class="dashboard-container">
                    <div class="dash-header">
                        <div class="title-group">
                            <h1>NEBULA 核心侦测网络总控台</h1>
                            <p>Nebula Security Analysis & Intelligence SOC</p>
                        </div>
                        <div class="live-badge"><span class="dot"></span> SYSTEM ONLINE</div>
                    </div>

                    <div class="dash-grid">
                        <div class="dash-card clickable" on:click={() => app.setActiveTab('pcapList')}>
                            <div class="card-icon pcap-icon">📦</div>
                            <div class="card-content">
                                <div class="card-label">本地流量包沉淀</div>
                                <div class="card-value">{dashStats.pcapCount} <span class="unit">Files</span></div>
                                <div class="card-sub">SQLite 索引映射正常</div>
                            </div>
                        </div>

                        <div class="dash-card">
                            <div class="card-icon ws-icon">🦈</div>
                            <div class="card-content">
                                <div class="card-label">协议分析引擎 (wireshark)</div>
                                <div class="card-value text-glow">{dashStats.wsVersion}</div>
                                <div class="card-sub">HTTP/RPC 通信链路畅通</div>
                            </div>
                        </div>

                        <div class="dash-card">
                            <div class="card-icon zeek-icon">🛡️</div>
                            <div class="card-content">
                                <div class="card-label">Zeek 威胁感知引擎</div>
                                <div class="card-value zeek-versions">
                                    <div class="version-row">
                                        <span class="version-label">Zeek:</span>
                                        <span class="version-value">{dashStats.zeekVersion}</span>
                                    </div>
                                    <div class="version-row">
                                        <span class="version-label">Zeek-Kafka:</span>
                                        <span class="version-value">{dashStats.zeekKafkaVersion}</span>
                                    </div>
                                </div>
                                <div class="card-sub">已加载特征脚本：<span class="highlight">{dashStats.zeekScripts}</span> 个</div>
                            </div>
                        </div>

                        <div class="dash-card">
                            <div class="card-icon ai-icon">🧠</div>
                            <div class="card-content">
                                <div class="card-label">Dify 智能诊断基座</div>
                                <div class="card-value">{dashStats.aiModel}</div>
                                <div class="card-sub">历史消耗 Context: <span class="highlight">{dashStats.aiTokens}</span> Tks</div>
                            </div>
                        </div>
                    </div>

                    <div class="dash-banner">
                        <div class="banner-text">
                            <h3>🚀 工作站全模块已就绪</h3>
                            <p>点击左侧菜单栏进入 <strong>"协议分析引擎"</strong>，开启对底层流量的深度透视与多维检索之旅。</p>
                        </div>
                        <button class="start-btn" on:click={() => app.setActiveTab('analyzer')}>立即开始分析 ➔</button>
                    </div>
                </div>
            {:else if activeTab === 'pcapList'}
                <!-- PCAP 流量包列表 -->
                <PcapList on:analyze={handlePcapAnalyze} />
            {:else if activeTab === 'analyzer'}
                <!-- 协议分析引擎 -->
                <Workspace />
            {:else if activeTab === 'profile'}
                <!-- 个人中心 -->
                <Profile />
            {:else if activeTab === 'users'}
                <!-- 用户管理（管理员功能） -->
                {#if user && user.roleCode === 'admin'}
                    <UserManagement />
                {:else}
                    <div class="module-placeholder">
                        <div class="placeholder-content">
                            <span class="placeholder-icon">🔒</span>
                            <h3>权限不足</h3>
                            <p>只有超级管理员可以访问用户管理页面</p>
                        </div>
                    </div>
                {/if}
            {:else if activeTab === 'zeek'}
                <!-- Zeek 入侵检测（待开发） -->
                <div class="module-placeholder">
                    <div class="placeholder-content">
                        <span class="placeholder-icon">🛡️</span>
                        <h3>Zeek-Runner 规则下发与实时日志分析引擎</h3>
                        <p>功能开发中...</p>
                    </div>
                </div>
            {:else if activeTab === 'ai'}
                <!-- Dify 智能诊断（待开发） -->
                <div class="module-placeholder">
                    <div class="placeholder-content">
                        <span class="placeholder-icon">🧠</span>
                        <h3>Dify Agent 交互面板</h3>
                        <p>选中流量自动发送 AI 诊断，功能开发中...</p>
                    </div>
                </div>
            {/if}
        </Sidebar>
    {/if}
</div>

<style>
    :global(body) {
        margin: 0;
        font-family: var(--font-family);
        background-color: var(--bg-primary);
        color: var(--text-primary);
        overflow: hidden;
    }
    
    #app-wrapper {
        height: 100vh;
        width: 100vw;
        background-color: var(--bg-primary);
    }
    
    /* 控制台首页样式 */
    .dashboard-container {
        display: flex;
        flex-direction: column;
        height: 100%;
        gap: 24px;
        padding: 1.5rem;
        animation: fadeIn 0.5s ease-out;
    }
    
    @keyframes fadeIn {
        from { opacity: 0; transform: translateY(10px); }
        to { opacity: 1; transform: translateY(0); }
    }
    
    .dash-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-end;
        padding-bottom: 16px;
        border-bottom: 1px solid var(--border-color);
    }
    
    .title-group h1 {
        margin: 0 0 8px 0;
        font-size: var(--font-2xl);
        background: linear-gradient(135deg, #4f46e5, #10b981);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }
    
    .title-group p {
        margin: 0;
        color: var(--text-secondary);
        font-size: var(--font-sm);
    }
    
    .live-badge {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 6px 12px;
        background: var(--color-success-light);
        color: var(--color-success);
        border-radius: var(--radius-full);
        font-size: var(--font-xs);
        font-weight: 600;
        letter-spacing: 0.5px;
    }
    
    .dot {
        width: 8px;
        height: 8px;
        background: var(--color-success);
        border-radius: 50%;
        animation: pulse 2s infinite;
    }
    
    @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
    }
    
    .dash-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
        gap: 20px;
    }
    
    .dash-card {
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-lg);
        padding: var(--spacing-lg);
        display: flex;
        gap: var(--spacing-lg);
        align-items: center;
        transition: var(--transition-base);
    }
    
    .dash-card:hover {
        border-color: var(--border-color-light);
        box-shadow: var(--shadow-lg);
        transform: translateY(-2px);
    }

    .dash-card.clickable {
        cursor: pointer;
    }

    .dash-card.clickable:hover {
        background: var(--bg-tertiary);
    }
    
    .card-icon {
        font-size: 3rem;
        flex-shrink: 0;
    }
    
    .card-content {
        flex: 1;
    }
    
    .card-label {
        color: var(--text-secondary);
        font-size: var(--font-xs);
        margin-bottom: 8px;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }
    
    .card-value {
        font-size: var(--font-2xl);
        font-weight: 700;
        color: var(--text-primary);
        margin-bottom: 4px;
    }
    
    .card-value.text-glow {
        text-shadow: 0 0 20px rgba(79, 70, 229, 0.5);
    }
    
    .card-value.zeek-versions {
        display: flex;
        flex-direction: column;
        gap: 4px;
        font-size: var(--font-xl);
    }
    
    .card-value .version-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }
    
    .card-value .version-label {
        font-size: var(--font-sm);
        font-weight: 600;
        color: var(--text-secondary);
        min-width: 50px;
    }
    
    .card-value .version-value {
        font-size: var(--font-lg);
        font-weight: 700;
        color: var(--color-primary);
        font-family: var(--font-mono);
    }
    
    .card-value .unit {
        font-size: var(--font-sm);
        font-weight: 400;
        color: var(--text-muted);
        margin-left: 4px;
    }
    
    .card-sub {
        color: var(--text-muted);
        font-size: var(--font-xs);
    }
    
    .card-sub .highlight {
        color: var(--color-info);
        font-weight: 600;
    }
    
    .dash-banner {
        background: linear-gradient(135deg, var(--color-primary-light), var(--color-info-light));
        border: 1px solid var(--border-color-light);
        border-radius: var(--radius-lg);
        padding: var(--spacing-xl);
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: auto;
    }
    
    .banner-text h3 {
        margin: 0 0 8px 0;
        color: var(--text-primary);
        font-size: var(--font-lg);
    }
    
    .banner-text p {
        margin: 0;
        color: var(--text-secondary);
        font-size: var(--font-sm);
    }
    
    .start-btn {
        padding: 12px 24px;
        background: var(--color-primary);
        color: white;
        border: none;
        border-radius: var(--radius-md);
        font-weight: 600;
        cursor: pointer;
        transition: var(--transition-base);
        white-space: nowrap;
    }
    
    .start-btn:hover {
        background: var(--color-primary-hover);
        transform: translateY(-2px);
        box-shadow: var(--shadow-lg);
    }
    
    /* 模块占位符样式 */
    .module-placeholder {
        border: 2px dashed var(--border-color-light);
        height: 100%;
        border-radius: var(--radius-lg);
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-secondary);
    }
    
    .placeholder-content {
        text-align: center;
        color: var(--text-secondary);
    }
    
    .placeholder-icon {
        font-size: 4rem;
        display: block;
        margin-bottom: var(--spacing-lg);
        opacity: 0.5;
    }
    
    .placeholder-content h3 {
        margin: 0 0 var(--spacing-md) 0;
        color: var(--text-primary);
        font-size: var(--font-xl);
    }
    
    .placeholder-content p {
        margin: 0;
        font-size: var(--font-sm);
    }
</style>
