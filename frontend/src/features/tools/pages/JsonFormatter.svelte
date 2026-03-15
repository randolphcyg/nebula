<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { success, error } from '../../../stores/toast';
    
    const dispatch = createEventDispatcher();
    
    let inputJson = '';
    let outputJson = '';
    let indentSize = 4;
    let sortKeys = false;
    
    function formatJson() {
        try {
            if (!inputJson.trim()) {
                error('请输入 JSON 字符串');
                return;
            }
            
            const parsed = JSON.parse(inputJson);
            const replacer = sortKeys ? (key: string, value: any) => 
                value instanceof Object && !(value instanceof Array) 
                    ? Object.keys(value).sort().reduce((sorted, k) => {
                        sorted[k] = value[k];
                        return sorted;
                    }, {} as any)
                    : value
                : undefined;
            
            outputJson = JSON.stringify(parsed, replacer, indentSize);
            success('JSON 格式化成功');
        } catch (e) {
            error(`JSON 格式错误：${(e as Error).message}`);
            outputJson = '';
        }
    }
    
    function clearAll() {
        inputJson = '';
        outputJson = '';
    }
    
    function copyOutput() {
        if (!outputJson) {
            error('没有可复制的内容');
            return;
        }
        
        navigator.clipboard.writeText(outputJson).then(() => {
            success('已复制到剪贴板');
        }).catch(() => {
            error('复制失败');
        });
    }
    
    function goBack() {
        dispatch('back');
    }
</script>

<div class="json-formatter">
    <div class="formatter-header">
        <div class="header-left">
            <button class="back-btn" on:click={goBack} title="返回">
                <span>←</span>
            </button>
            <h2 class="formatter-title">
                <span class="title-icon">📋</span>
                JSON 格式化
            </h2>
        </div>
        
        <div class="formatter-options">
            <div class="option-group">
                <label class="option-label">
                    <input type="checkbox" bind:checked={sortKeys} />
                    排序键名
                </label>
            </div>
            <div class="option-group">
                <label class="option-label">
                    缩进：
                    <select bind:value={indentSize} class="indent-select">
                        <option value={2}>2 空格</option>
                        <option value={4}>4 空格</option>
                        <option value={8}>8 空格</option>
                        <option value={0}>Tab</option>
                    </select>
                </label>
            </div>
            <button class="action-btn clear" on:click={clearAll}>
                <span class="btn-icon">🗑️</span>
                清空
            </button>
            <button class="action-btn primary" on:click={formatJson}>
                <span class="btn-icon">✨</span>
                格式化
            </button>
        </div>
    </div>
    
    <div class="formatter-content">
        <div class="formatter-panel">
            <div class="panel-header">
                <span class="panel-title">输入 JSON</span>
                <button class="panel-btn" on:click={() => { inputJson = ''; }} title="清空">
                    🗑️
                </button>
            </div>
            <textarea 
                class="input-area" 
                bind:value={inputJson}
                placeholder="请粘贴 JSON 字符串到这里..."
                spellcheck="false"
            ></textarea>
        </div>
        
        <div class="formatter-panel">
            <div class="panel-header">
                <span class="panel-title">格式化结果</span>
                <button class="panel-btn" on:click={copyOutput} title="复制">
                    📋
                </button>
            </div>
            <textarea 
                class="output-area" 
                bind:value={outputJson}
                placeholder="格式化后的 JSON 将显示在这里..."
                readonly
                spellcheck="false"
            ></textarea>
        </div>
    </div>
</div>

<style>
    .json-formatter {
        display: flex;
        flex-direction: column;
        height: 100%;
    }
    
    .formatter-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
        padding-bottom: 1rem;
        border-bottom: 1px solid var(--border-color);
    }
    
    .header-left {
        display: flex;
        align-items: center;
        gap: 1rem;
    }
    
    .back-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        padding: 0;
        background: transparent;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        color: var(--text-secondary);
        cursor: pointer;
        transition: var(--transition-fast);
        font-size: 1.25rem;
    }
    
    .back-btn:hover {
        background: var(--bg-tertiary);
        border-color: var(--color-primary);
        color: var(--color-primary);
    }
    
    .formatter-title {
        font-size: 1.25rem;
        font-weight: 600;
        color: var(--text-primary);
        margin: 0;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    
    .title-icon {
        font-size: 1.5rem;
    }
    
    .formatter-options {
        display: flex;
        align-items: center;
        gap: 1rem;
    }
    
    .option-group {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    
    .option-label {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 0.875rem;
        color: var(--text-secondary);
        cursor: pointer;
    }
    
    .option-label input[type="checkbox"] {
        cursor: pointer;
    }
    
    .indent-select {
        padding: 4px 8px;
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: 4px;
        color: var(--text-primary);
        font-size: 0.875rem;
        cursor: pointer;
    }
    
    .action-btn {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        padding: 6px 12px;
        background: transparent;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        font-size: 0.875rem;
        color: var(--text-secondary);
        cursor: pointer;
        transition: var(--transition-fast);
    }
    
    .action-btn:hover {
        background: var(--bg-tertiary);
        border-color: var(--color-primary);
        color: var(--color-primary);
    }
    
    .action-btn.primary {
        background: var(--color-primary);
        border-color: var(--color-primary);
        color: white;
    }
    
    .action-btn.primary:hover {
        background: var(--color-primary-dark);
        border-color: var(--color-primary-dark);
    }
    
    .action-btn.clear:hover {
        background: var(--bg-tertiary);
        border-color: var(--color-danger);
        color: var(--color-danger);
    }
    
    .btn-icon {
        font-size: 1rem;
    }
    
    .formatter-content {
        display: flex;
        gap: 1rem;
        flex: 1;
        min-height: 0;
    }
    
    .formatter-panel {
        flex: 1;
        display: flex;
        flex-direction: column;
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        overflow: hidden;
    }
    
    .panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem 1rem;
        border-bottom: 1px solid var(--border-color);
        background: var(--bg-tertiary);
    }
    
    .panel-title {
        font-size: 0.875rem;
        font-weight: 600;
        color: var(--text-primary);
    }
    
    .panel-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 28px;
        padding: 0;
        background: transparent;
        border: none;
        border-radius: 4px;
        color: var(--text-secondary);
        cursor: pointer;
        transition: var(--transition-fast);
        font-size: 1rem;
    }
    
    .panel-btn:hover {
        background: var(--bg-primary);
        color: var(--color-primary);
    }
    
    .input-area, .output-area {
        flex: 1;
        padding: 1rem;
        background: transparent;
        border: none;
        color: var(--text-primary);
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 0.875rem;
        line-height: 1.5;
        resize: none;
        outline: none;
    }
    
    .input-area {
        min-height: 200px;
    }
    
    .output-area {
        min-height: 200px;
        background: var(--bg-primary);
    }
</style>
