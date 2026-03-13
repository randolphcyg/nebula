<script lang="ts">
    import PcapList from './analyzer/PcapList.svelte';
    import PcapDetail from './analyzer/PcapDetail.svelte';

    // 路由状态：仪表盘 -> 列表页 -> 详情页
    let currentView: 'dashboard' | 'pcap-list' | 'pcap-detail' | 'live' | 'auto' = 'dashboard';

    // 跨页面传递的文件对象
    let selectedFile: any = null;

    const features = [
        {
            id: 'pcap-list',
            title: '离线 PCAP 深度分析',
            desc: '调用 wireshark HTTP 接口，支持大型流量包的解析与分段检索。',
            icon: '📂',
            status: 'Ready'
        },
        {id: 'live', title: '网卡实时抓包', desc: '基于 libpcap 捕获网卡实时流量。', icon: '⚡', status: 'Planning'},
        {
            id: 'auto',
            title: '车载/工控协议专区',
            desc: '特定工业协议（如 CAN, Modbus）的时序与指令解析，适配硬件级调试。',
            icon: '🚗',
            status: 'Planning'
        }
    ];

    function navigateTo(viewId: any) {
        currentView = viewId;
    }

    // 接收列表页派发的 analyze 事件
    function handleAnalyze(event: CustomEvent) {
        selectedFile = event.detail;
        currentView = 'pcap-detail';
    }
</script>

<div class="hub-container">
    {#if currentView === 'dashboard'}
        <div class="dashboard">
            <div class="grid">
                {#each features as feature}
                    <div class="card" role="button" tabindex="0"
                         on:click={() => navigateTo(feature.id)}
                         on:keydown={(e) => e.key === 'Enter' && navigateTo(feature.id)}>
                        <div class="card-header">
                            <span class="icon">{feature.icon}</span>
                            <span class={`badge ${feature.status.toLowerCase()}`}>{feature.status}</span>
                        </div>
                        <h3>{feature.title}</h3>
                        <p>{feature.desc}</p>
                    </div>
                {/each}
            </div>
        </div>

    {:else}
        <div class="sub-page">
            <div class="sub-header">
                {#if currentView === 'pcap-list'}
                    <button class="back-btn" on:click={() => navigateTo('dashboard')}>← 返回工作台</button>
                    <span class="title">离线流量分析</span>
                {:else if currentView === 'pcap-detail'}
                    <button class="back-btn" on:click={() => navigateTo('pcap-list')}>← 返回文件列表</button>
                    <span class="title">正在分析: <strong class="highlight">{selectedFile?.fileName}</strong></span>
                {:else}
                    <button class="back-btn" on:click={() => navigateTo('dashboard')}>← 返回工作台</button>
                {/if}
            </div>

            <div class="sub-content">
                {#if currentView === 'pcap-list'}
                    <PcapList on:analyze={handleAnalyze}/>
                {:else if currentView === 'pcap-detail'}
                    <PcapDetail file={selectedFile}/>
                {:else}
                    <div class="wip"><h2>模块正在开发接入中...</h2></div>
                {/if}
            </div>
        </div>
    {/if}
</div>

<style>
    .hub-container {
        height: 100%;
        display: flex;
        flex-direction: column;
    }

    .dashboard {
        padding: 1rem;
    }

    .grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 1rem;
    }

    .card {
        background: #1e293b;
        border: 1px solid #334155;
        border-radius: 8px;
        padding: 1.5rem;
        cursor: pointer;
        transition: 0.2s;
        outline: none;
    }

    .card:hover, .card:focus {
        border-color: #6366f1;
        transform: translateY(-2px);
    }

    .card-header {
        display: flex;
        justify-content: space-between;
        margin-bottom: 1rem;
    }

    .icon {
        font-size: 1.5rem;
    }

    .badge {
        font-size: 0.7rem;
        padding: 2px 8px;
        border-radius: 12px;
        font-weight: bold;
    }

    .badge.ready {
        background: #064e3b;
        color: #34d399;
    }

    .badge.planning {
        background: #1e3a8a;
        color: #93c5fd;
    }

    h3 {
        margin: 0 0 0.5rem 0;
        font-size: 1.1rem;
        color: #f8fafc;
    }

    p {
        margin: 0;
        font-size: 0.85rem;
        color: #94a3b8;
        line-height: 1.4;
    }

    .sub-page {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
    }

    /* 统一导航栏样式 */
    .sub-header {
        display: flex;
        align-items: center;
        gap: 1rem;
        padding: 0 0 12px 0;
        border-bottom: 1px solid #334155;
        margin-bottom: 12px;
        color: #f1f5f9;
        font-weight: bold;
    }

    .back-btn {
        background: transparent;
        border: 1px solid #475569;
        color: #cbd5e1;
        padding: 6px 14px;
        border-radius: 6px;
        cursor: pointer;
        transition: 0.2s;
        font-size: 0.9rem;
    }

    .back-btn:hover {
        background: #1e293b;
        color: white;
    }

    .title {
        font-size: 1.05rem;
    }

    .highlight {
        color: #38bdf8;
        margin-left: 6px;
    }

    .sub-content {
        flex: 1;
        overflow: hidden;
        display: flex;
        flex-direction: column;
    }

    .wip {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 80%;
        color: #64748b;
    }
</style>