<script lang="ts">
    import { onMount } from 'svelte';
    // 引入 Wails 绑定的后端方法
    import { GetFileList, GetWiresharkVersion } from '../wailsjs/go/main/App';
    // 引入 Analyzer Hub
    import AnalyzerHub from './lib/AnalyzerHub.svelte';

    let activeTab: 'home' | 'analyzer' | 'zeek' | 'ai' = 'home';

    let status = {
        cgo: "API Connected",
        zeek: "Idle",
        ai: "Ready"
    };

    const tabs = [
        {id: 'home', label: '控制台', icon: '🏠'},
        {id: 'analyzer', label: '协议分析引擎', icon: '🔍'},
        {id: 'zeek', label: 'Zeek 入侵检测', icon: '🛡️'},
        {id: 'ai', label: 'Dify 智能诊断', icon: '🧠'}
    ];

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

    onMount(async () => {
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
            console.error("大屏获取Wireshark版本失败:", e);
        }
    });

</script>

<div id="app-container">
    <aside class="sidebar">
        <div class="brand">NEBULA</div>
        <nav>
            {#each tabs as tab}
                <button
                        class:active={activeTab === tab.id}
                        on:click={() => activeTab = tab.id}
                >
                    <span class="icon">{tab.icon}</span>
                    <span class="label">{tab.label}</span>
                </button>
            {/each}
        </nav>
        <div class="sys-status">
            <div class="status-item">Engine: <span class="online">{status.cgo}</span></div>
        </div>
    </aside>

    <main class="content-area">
        <header>
            <h2>{tabs.find(t => t.id === activeTab)?.label}</h2>
        </header>

        <section class="viewport">
            {#if activeTab === 'home'}

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
                                <div class="card-sub">已加载特征脚本: <span class="highlight">{dashStats.zeekScripts}</span> 个</div>
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
                            <p>点击左侧菜单栏进入 <strong>“协议分析引擎”</strong>，开启对底层流量的深度透视与多维检索之旅。</p>
                        </div>
                        <button class="start-btn" on:click={() => activeTab = 'analyzer'}>立即开始分析 ➔</button>
                    </div>
                </div>
            {:else if activeTab === 'analyzer'}
                <AnalyzerHub/>
            {:else if activeTab === 'zeek'}
                <div class="module-placeholder">Zeek-Runner 规则下发与实时日志分析引擎 (开发中...)</div>
            {:else if activeTab === 'ai'}
                <div class="module-placeholder">Dify Agent 交互面板：选中流量自动发送 AI 诊断 (开发中...)</div>
            {/if}
        </section>
    </main>
</div>

<style>
    :global(body) {
        margin: 0;
        font-family: 'Inter', system-ui, sans-serif;
        background-color: #0f172a;
        color: white;
        overflow: hidden;
    }

    #app-container {
        display: flex;
        height: 100vh;
        width: 100vw;
    }

    /* 侧边栏及原有布局样式保持不变 */
    .sidebar { width: 180px; background: #0b1120; display: flex; flex-direction: column; border-right: 1px solid #1e293b; }
    .brand { padding: 1.5rem; font-size: 1.2rem; font-weight: 800; letter-spacing: 2px; color: #6366f1; border-bottom: 1px solid #1e293b; }
    nav { flex: 1; padding: 1rem 0.5rem; }
    nav button { width: 100%; padding: 10px 12px; margin-bottom: 8px; border: none; background: transparent; color: #94a3b8; text-align: left; cursor: pointer; border-radius: 6px; transition: 0.2s; display: flex; align-items: center; font-size: 0.9rem; }
    nav button:hover { background: #1e293b; color: white; }
    nav button.active { background: #4f46e5; color: white; font-weight: bold; }
    .icon { margin-right: 10px; font-size: 1.1rem; }
    .content-area { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: #0f172a; }
    header { padding: 1rem 1.5rem; background: #0b1120; border-bottom: 1px solid #1e293b; }
    header h2 { margin: 0; font-size: 1.1rem; color: #f8fafc; }
    .viewport { flex: 1; padding: 1.5rem; overflow: hidden; display: flex; flex-direction: column; }
    .module-placeholder { border: 2px dashed #334155; height: 100%; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #64748b; font-size: 0.95rem; }
    .online { color: #10b981; font-weight: bold; }
    .sys-status { padding: 1rem; font-size: 0.75rem; background: #05080f; border-top: 1px solid #1e293b; }

    /* ================= 新增的大屏专属样式 ================= */
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
        border-bottom: 1px solid #1e293b;
    }

    .title-group h1 { margin: 0 0 8px 0; font-size: 1.8rem; color: #f8fafc; letter-spacing: 1px;}
    .title-group p { margin: 0; color: #64748b; font-size: 0.9rem; font-family: monospace;}

    .live-badge {
        display: flex;
        align-items: center;
        gap: 8px;
        background: rgba(16, 185, 129, 0.1);
        border: 1px solid rgba(16, 185, 129, 0.2);
        color: #10b981;
        padding: 6px 12px;
        border-radius: 20px;
        font-size: 0.85rem;
        font-weight: bold;
        font-family: monospace;
    }
    .dot { width: 8px; height: 8px; background: #10b981; border-radius: 50%; box-shadow: 0 0 8px #10b981; animation: pulse 1.5s infinite;}
    @keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.4; } 100% { opacity: 1; } }

    .dash-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
        gap: 20px;
    }

    .dash-card {
        background: linear-gradient(145deg, #111827 0%, #0b1120 100%);
        border: 1px solid #1e293b;
        border-radius: 12px;
        padding: 20px;
        display: flex;
        align-items: flex-start;
        gap: 16px;
        transition: transform 0.2s, box-shadow 0.2s, border-color 0.2s;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    }

    .dash-card:hover {
        transform: translateY(-4px);
        box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.2);
        border-color: #334155;
    }

    .card-icon {
        font-size: 2.2rem;
        padding: 12px;
        border-radius: 12px;
        background: #1e293b;
    }

    /* 不同卡片的专属渐变底色 */
    .pcap-icon { background: linear-gradient(135deg, rgba(99,102,241,0.2) 0%, rgba(99,102,241,0.05) 100%); color: #818cf8; }
    .ws-icon { background: linear-gradient(135deg, rgba(56,189,248,0.2) 0%, rgba(56,189,248,0.05) 100%); color: #38bdf8; }
    .zeek-icon { background: linear-gradient(135deg, rgba(245,158,11,0.2) 0%, rgba(245,158,11,0.05) 100%); color: #fbbf24; }
    .ai-icon { background: linear-gradient(135deg, rgba(168,85,247,0.2) 0%, rgba(168,85,247,0.05) 100%); color: #c084fc; }

    .card-content { display: flex; flex-direction: column; }
    .card-label { color: #94a3b8; font-size: 0.85rem; margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.5px;}
    .card-value { color: #f8fafc; font-size: 1.6rem; font-weight: 700; margin-bottom: 6px; font-family: 'Fira Code', monospace;}
    .card-value .unit { font-size: 0.9rem; color: #64748b; font-weight: normal; }
    .text-glow { text-shadow: 0 0 10px rgba(56,189,248, 0.4); color: #38bdf8;}

    .card-sub { color: #64748b; font-size: 0.8rem; }
    .highlight { color: #e2e8f0; font-weight: bold; }

    .dash-banner {
        margin-top: auto;
        background: linear-gradient(90deg, rgba(79, 70, 229, 0.1) 0%, rgba(15, 23, 42, 0) 100%);
        border: 1px solid #334155;
        border-left: 4px solid #4f46e5;
        border-radius: 8px;
        padding: 20px 24px;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .banner-text h3 { margin: 0 0 8px 0; color: #f1f5f9; font-size: 1.1rem; }
    .banner-text p { margin: 0; color: #94a3b8; font-size: 0.9rem; }

    .start-btn {
        background: #4f46e5;
        color: white;
        border: none;
        padding: 12px 24px;
        border-radius: 6px;
        font-size: 0.95rem;
        font-weight: bold;
        cursor: pointer;
        transition: 0.2s;
        box-shadow: 0 4px 6px rgba(79, 70, 229, 0.2);
    }
    .start-btn:hover {
        background: #4338ca;
        transform: translateY(-2px);
        box-shadow: 0 6px 12px rgba(79, 70, 229, 0.3);
    }
</style>