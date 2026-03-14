<script lang="ts">
    import PcapList from '../analyzer/pages/PcapList.svelte';
    import PcapDetail from '../analyzer/pages/PcapDetail.svelte';
    import InterfaceList from '../analyzer/pages/InterfaceList.svelte';
    import Profile from '../user/pages/Profile.svelte';
    import { info as showInfo } from '../../stores/toast';
    import { onMount, createEventDispatcher } from 'svelte';
    import { app } from '../../stores/app';

    let currentView: 'dashboard' | 'pcap-list' | 'pcap-detail' | 'live' | 'auto' | 'profile' = 'dashboard';
    let activeTab = 'home';
    
    const dispatch = createEventDispatcher();

    let selectedFile: any = null;
    let selectedInterface: any = null;

    // 订阅当前激活的 tab
    app.subscribe(state => {
        activeTab = state.activeTab;
        // 当切换到 profile tab 时，显示个人中心视图
        if (activeTab === 'profile') {
            currentView = 'profile';
        } else if (activeTab === 'analyzer' && currentView === 'profile') {
            // 从 profile 切换到 analyzer 时，回到 dashboard
            currentView = 'dashboard';
        }
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

    function navigateTo(viewId: any) {
        currentView = viewId;
        if (viewId === 'profile') {
            // 个人中心直接显示，不切换 tab
            // tab 已经在 Sidebar 中处理
        } else {
            // 其他功能保持 analyzer 标签激活
            dispatch('tabChange', 'analyzer');
        }
    }

    function handleAnalyze(event: CustomEvent) {
        selectedFile = event.detail;
        currentView = 'pcap-detail';
    }

    function handleStartCapture(event: CustomEvent) {
        selectedInterface = event.detail;
        showInfo(`即将启动网卡 [${selectedInterface.name}] 的实时抓包功能`);
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
                    <span class="title">离线流量包列表</span>
                {:else if currentView === 'pcap-detail'}
                    <button class="back-btn" on:click={() => navigateTo('pcap-list')}>← 返回文件列表</button>
                    <span class="title">正在分析：<strong class="highlight">{selectedFile?.fileName}</strong></span>
                {:else if currentView === 'live'}
                    <button class="back-btn" on:click={() => navigateTo('dashboard')}>← 返回工作台</button>
                    <span class="title">网卡列表</span>
                {:else if currentView === 'profile'}
                    <button class="back-btn" on:click={() => navigateTo('dashboard')}>← 返回工作台</button>
                    <span class="title">个人中心</span>
                {:else}
                    <button class="back-btn" on:click={() => navigateTo('dashboard')}>← 返回工作台</button>
                {/if}
            </div>

            <div class="sub-content">
                {#if currentView === 'pcap-list'}
                    <PcapList on:analyze={handleAnalyze}/>
                {:else if currentView === 'pcap-detail'}
                    <PcapDetail file={selectedFile}/>
                {:else if currentView === 'live'}
                    <InterfaceList on:capture={handleStartCapture} />
                {:else if currentView === 'profile'}
                    <Profile />
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
        padding: 1rem;
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
        border-bottom: 1px solid var(--border-color);
        margin-bottom: 12px;
        color: var(--text-primary);
        font-weight: bold;
    }

    .back-btn {
        background: transparent;
        border: 1px solid var(--border-color);
        color: var(--text-secondary);
        padding: 6px 14px;
        border-radius: 6px;
        cursor: pointer;
        transition: 0.2s;
        font-size: 0.9rem;
    }

    .back-btn:hover {
        background: var(--bg-tertiary);
        color: var(--text-primary);
    }

    .title {
        font-size: 1.05rem;
    }

    .highlight {
        color: var(--color-info);
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
        color: var(--text-muted);
    }
</style>