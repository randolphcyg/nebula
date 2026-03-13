<script lang="ts">
    import { onMount } from 'svelte';
    import { auth, app } from './stores';
    import { showSuccess } from './utils';
    import { ToastContainer } from './components/ui';
    import { Sidebar } from './components/layout';
    import Login from './auth/Login.svelte';
    import Workspace from './features/workspace/Workspace.svelte';
    import { GetFileList, GetWiresharkVersion } from '../wailsjs/go/main/App';
    
    let isAuthenticated = false;
    let activeTab = 'home';
    
    // ==== 大屏统计状态 ====
    let dashStats = {
        pcapCount: 0,
        wsVersion: "探测中...",
        zeekVersion: "v8.1.0 (TODO)",
        zeekScripts: 142,
        aiModel: "qwen2.5-coder:3b (TODO)",
        aiTokens: "1.2M",
        systemUptime: "0 Days"
    };
    
    // 监听认证状态
    auth.subscribe(state => {
        isAuthenticated = state.isAuthenticated;
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
            const pcapResp = await GetFileList({ page: 1, pageSize: 1 });
            dashStats.pcapCount = pcapResp.total || 0;
        } catch(e) {
            console.error("大屏获取文件数失败:", e);
        }

        // 2. 获取 wireshark 引擎版本
        try {
            const wsRespStr = await GetWiresharkVersion();
            const wsObj = JSON.parse(wsRespStr);
            dashStats.wsVersion = wsObj.version || "版本未知";
        } catch(e) {
            dashStats.wsVersion = "引擎离线";
            console.error("大屏获取 Wireshark 版本失败:", e);
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
</script>

<div id="app-wrapper">
    <ToastContainer />
    
    {#if !isAuthenticated}
        <!-- 登录页面 -->
        <Login on:success={handleLoginSuccess} />
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
                        <div class="dash-card">
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
                                <div class="card-value">{dashStats.zeekVersion}</div>
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
            {:else if activeTab === 'analyzer'}
                <!-- 协议分析引擎 -->
                <Workspace />
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
    }
    
    /* 控制台首页样式 */
    .dashboard-container {
        display: flex;
        flex-direction: column;
        height: 100%;
        gap: 24px;
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
