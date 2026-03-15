<script lang="ts">
    import PcapList from '../analyzer/pages/PcapList.svelte';
    import PcapDetail from '../analyzer/pages/PcapDetail.svelte';
    import InterfaceList from '../analyzer/pages/InterfaceList.svelte';
    import Tools from '../tools/pages/Tools.svelte';
    import JsonFormatter from '../tools/pages/JsonFormatter.svelte';
    import { info as showInfo } from '../../stores/toast';
    import { onMount, createEventDispatcher } from 'svelte';
    import { app } from '../../stores/app';
    import { WindowToggleMaximise } from '../../../wailsjs/runtime/runtime';

    let currentView: 'dashboard' | 'pcap-list' | 'pcap-detail' | 'live' | 'auto' | 'tools' | 'json-formatter' = 'dashboard';
    
    const dispatch = createEventDispatcher();

    let selectedFile: any = null;
    let selectedInterface: any = null;

    // 订阅当前激活的 tab
    app.subscribe(state => {
        // Workspace 只处理 analyzer tab 的内容
    });

    // 监听全局 analyze 事件
    onMount(() => {
        function handleAnalyzeEvent(event: Event) {
            const customEvent = event as CustomEvent;
            selectedFile = customEvent.detail;
            currentView = 'pcap-detail';
        }

        window.addEventListener('analyze', handleAnalyzeEvent);
        
        return () => {
            window.removeEventListener('analyze', handleAnalyzeEvent);
        };
    });

    function navigateTo(viewId: any) {
        currentView = viewId;
    }

    function handleAnalyze(event: CustomEvent) {
        selectedFile = event.detail;
        currentView = 'pcap-detail';
    }

    function handleStartCapture(event: CustomEvent) {
        selectedInterface = event.detail;
        showInfo(`即将启动网卡 [${selectedInterface.name}] 的实时抓包功能`);
    }

    function handleMaximize() {
        WindowToggleMaximise();
    }

    function handleOpenTool(event: CustomEvent) {
        const toolId = event.detail;
        if (toolId === 'json-formatter') {
            currentView = 'json-formatter';
        }
    }

    const features = [
        {
            id: 'pcap-list',
            title: '离线 PCAP 深度分析',
            desc: '调用 wireshark HTTP 接口，支持大型流量包的解析与分段检索。',
            icon: '📂',
            status: 'Ready'
        },
        { id: 'live', title: '网卡实时抓包', desc: '基于 libpcap 捕获网卡实时流量。', icon: '⚡', status: 'Ready' },
        {
            id: 'auto',
            title: '车载/工控协议专区',
            desc: '特定工业协议（如 CAN, Modbus）的时序与指令解析，适配硬件级调试。',
            icon: '🚗',
            status: 'Planning'
        }
    ];
</script>

<div class="hub-container">
    {#if currentView === 'dashboard'}
        <div class="sub-page">
            <div class="sub-header">
                <div class="breadcrumb">
                    <span class="breadcrumb-item active">
                        <span class="breadcrumb-icon">🔍</span>
                        协议分析引擎
                    </span>
                </div>
                
                <button class="window-control-btn" on:click={handleMaximize} title="最大化/还原窗口">
                    <span class="btn-icon">⛶</span>
                </button>
            </div>
            
            <div class="sub-content">
                <div class="section">
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
            </div>
        </div>

    {:else if currentView === 'tools' || currentView === 'json-formatter'}
        <div class="sub-page">
            <div class="sub-header">
                <div class="breadcrumb">
                    <a class="breadcrumb-item" on:click={() => app.setActiveTab('home')}>
                        <span class="breadcrumb-icon">🏠</span>
                        工作台
                    </a>
                    <span class="breadcrumb-separator">/</span>
                    <span class="breadcrumb-item active">通用工具</span>
                </div>
                
                <button class="window-control-btn" on:click={handleMaximize} title="最大化/还原窗口">
                    <span class="btn-icon">⛶</span>
                </button>
            </div>
            
            <div class="sub-content">
                {#if currentView === 'tools'}
                    <Tools on:openTool={handleOpenTool} />
                {:else if currentView === 'json-formatter'}
                    <JsonFormatter on:back={() => { currentView = 'tools' }} />
                {/if}
            </div>
        </div>

    {:else}
        <div class="sub-page">
            <div class="sub-header">
                <div class="breadcrumb">
                    <a class="breadcrumb-item" on:click={() => navigateTo('dashboard')}>
                        <span class="breadcrumb-icon">🔍</span>
                        协议分析引擎
                    </a>
                    
                    {#if currentView === 'pcap-detail'}
                        <span class="breadcrumb-separator">/</span>
                        <a class="breadcrumb-item" on:click={() => navigateTo('pcap-list')}>
                            PCAP 列表
                        </a>
                        <span class="breadcrumb-separator">/</span>
                        <span class="breadcrumb-item active">
                            {selectedFile?.fileName}
                        </span>
                    {:else if currentView === 'pcap-list'}
                        <span class="breadcrumb-separator">/</span>
                        <span class="breadcrumb-item active">PCAP 列表</span>
                    {:else if currentView === 'live'}
                        <span class="breadcrumb-separator">/</span>
                        <span class="breadcrumb-item active">网卡实时抓包</span>
                    {:else if currentView === 'auto'}
                        <span class="breadcrumb-separator">/</span>
                        <span class="breadcrumb-item active">车载/工控协议专区</span>
                    {/if}
                </div>
                
                <button class="window-control-btn" on:click={handleMaximize} title="最大化/还原窗口">
                    <span class="btn-icon">⛶</span>
                </button>
            </div>

            <div class="sub-content">
                {#if currentView === 'pcap-list'}
                    <PcapList on:analyze={handleAnalyze}/>
                {:else if currentView === 'pcap-detail'}
                    <PcapDetail file={selectedFile}/>
                {:else if currentView === 'live'}
                    <InterfaceList on:capture={handleStartCapture} />
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
        background-color: var(--bg-primary);
    }

    .dashboard {
        padding: 1.5rem;
        overflow-y: auto;
    }

    .section {
        margin-bottom: 2rem;
    }

    .section-title {
        font-size: 1.25rem;
        font-weight: 700;
        color: var(--text-primary);
        margin-bottom: 1rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 1rem;
    }

    .card {
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        padding: 1.5rem;
        cursor: pointer;
        transition: 0.2s;
        outline: none;
    }

    .card:hover, .card:focus {
        border-color: var(--color-primary);
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(79, 70, 229, 0.2);
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
        background: rgba(6, 78, 59, 0.2);
        color: #10b981;
    }

    .badge.planning {
        background: rgba(30, 58, 138, 0.2);
        color: #3b82f6;
    }

    h3 {
        margin: 0 0 0.5rem 0;
        font-size: 1.1rem;
        color: var(--text-primary);
    }

    p {
        margin: 0;
        font-size: 0.85rem;
        color: var(--text-secondary);
        line-height: 1.4;
    }

    .wip {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100%;
        color: var(--text-secondary);
    }

    .tools-page {
        height: 100%;
        overflow: hidden;
    }
</style>
