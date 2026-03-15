<script lang="ts">
    import JsonFormatter from './JsonFormatter.svelte';
    import { app } from '../../../stores/app';
    import { WindowToggleMaximise } from '../../../../wailsjs/runtime/runtime';
    
    let currentTool: string | null = null;
    
    const tools = [
        {
            id: 'json-formatter',
            title: 'JSON 格式化',
            desc: '格式化、验证和美化 JSON 数据',
            icon: '📋',
            status: 'Ready'
        }
    ];
    
    function openTool(toolId: string) {
        currentTool = toolId;
    }
    
    function handleBack() {
        currentTool = null;
    }
    
    function handleMaximize() {
        WindowToggleMaximise();
    }
</script>

<div class="tools-container">
    {#if currentTool === null}
        <div class="sub-page">
            <div class="sub-header">
                <div class="breadcrumb">
                    <span class="breadcrumb-item active">
                        <span class="breadcrumb-icon">🛠️</span>
                        通用工具
                    </span>
                </div>
                
                <button class="window-control-btn" on:click={handleMaximize} title="最大化/还原窗口">
                    <span class="btn-icon">⛶</span>
                </button>
            </div>
            
            <div class="sub-content">
                <div class="tools-grid">
                    {#each tools as tool}
                        <div class="tool-card" 
                             role="button" 
                             tabindex="0"
                             on:click={() => openTool(tool.id)}
                             on:keydown={(e) => e.key === 'Enter' && openTool(tool.id)}>
                            <div class="tool-card-header">
                                <span class="tool-icon">{tool.icon}</span>
                                <span class={`tool-badge ${tool.status.toLowerCase()}`}>{tool.status}</span>
                            </div>
                            <h3>{tool.title}</h3>
                            <p>{tool.desc}</p>
                        </div>
                    {/each}
                </div>
            </div>
        </div>
    {:else if currentTool === 'json-formatter'}
        <div class="sub-page">
            <div class="sub-header">
                <div class="breadcrumb">
                    <a class="breadcrumb-item" on:click={handleBack}>
                        <span class="breadcrumb-icon">🛠️</span>
                        通用工具
                    </a>
                    <span class="breadcrumb-separator">/</span>
                    <span class="breadcrumb-item active">JSON 格式化</span>
                </div>
                
                <button class="window-control-btn" on:click={handleMaximize} title="最大化/还原窗口">
                    <span class="btn-icon">⛶</span>
                </button>
            </div>
            
            <div class="sub-content">
                <JsonFormatter on:back={handleBack} />
            </div>
        </div>
    {/if}
</div>

<style>
    .tools-container {
        height: 100%;
        display: flex;
        flex-direction: column;
    }
    
    /* 工具列表样式 */
    .tools-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 1rem;
    }
    
    .tool-card {
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        padding: 1.5rem;
        cursor: pointer;
        transition: var(--transition-base);
        outline: none;
    }
    
    .tool-card:hover, .tool-card:focus {
        border-color: var(--color-primary);
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(79, 70, 229, 0.2);
    }
    
    .tool-card-header {
        display: flex;
        justify-content: space-between;
        margin-bottom: 1rem;
    }
    
    .tool-icon {
        font-size: 1.5rem;
    }
    
    .tool-badge {
        font-size: 0.7rem;
        padding: 2px 8px;
        border-radius: 12px;
        font-weight: bold;
    }
    
    .tool-badge.ready {
        background: rgba(6, 78, 59, 0.2);
        color: #10b981;
    }
    
    .tool-badge.planning {
        background: rgba(30, 58, 138, 0.2);
        color: #3b82f6;
    }
    
    .tool-card h3 {
        margin: 0 0 0.5rem 0;
        font-size: 1.1rem;
        color: var(--text-primary);
    }
    
    .tool-card p {
        margin: 0;
        font-size: 0.85rem;
        color: var(--text-secondary);
        line-height: 1.4;
    }
</style>
