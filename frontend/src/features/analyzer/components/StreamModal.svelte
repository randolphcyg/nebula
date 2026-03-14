<script lang="ts">
    import { createEventDispatcher } from 'svelte';

    export let show = false;
    export let title = "";
    export let isLoading = false;
    export let payloads: { dir: 'client' | 'server', hexData: string }[] = [];

    export let streamId: number = 0;
    export let clientNode: string = "Client";
    export let serverNode: string = "Server";
    export let clientBytes: number = 0;
    export let serverBytes: number = 0;
    export let packetCount: number = 0;

    const dispatch = createEventDispatcher();

    let displayFormat = "ASCII";
    let displayDirection = "both";

    // ==== 核心保护：最大显示限制 (500KB) ====
    // 500KB 字节 = 1000,000 个 Hex 字符
    const MAX_DISPLAY_HEX_CHARS = 500 * 1024 * 2;

    $: filteredPayloads = payloads.filter(p => {
        if (displayDirection === 'both') return true;
        return p.dir === displayDirection;
    });

    // ==== 截断逻辑 ====
    let isTruncated = false;
    $: displayedPayloads = (() => {
        isTruncated = false;
        let result = [];
        let currentLen = 0;

        for (let p of filteredPayloads) {
            if (currentLen >= MAX_DISPLAY_HEX_CHARS) {
                isTruncated = true;
                break;
            }
            let available = MAX_DISPLAY_HEX_CHARS - currentLen;
            if (p.hexData.length > available) {
                // 仅截断显示，不影响原始 filteredPayloads
                result.push({ ...p, hexData: p.hexData.substring(0, available) });
                isTruncated = true;
                break;
            } else {
                result.push(p);
                currentLen += p.hexData.length;
            }
        }
        return result;
    })();

    function close() {
        show = false;
        dispatch('close');
    }

    // ==== 核心引擎：V8 极致优化版格式化算法 ====
    function formatData(hexStr: string, format: string, dir: string, index: number) {
        if (!hexStr) return "";

        // 【核心修复 1】：必须剥离原始的冒号，否则不仅格式错乱，还会导致长度计算错误！
        hexStr = hexStr.replace(/:/g, '');

        if (format === '原始数据') {
            return hexStr;
        }

        if (format === 'Hex 转储') {
            let lines = [];
            let offset = 0;
            for (let i = 0; i < hexStr.length; i += 32) {
                let chunk = hexStr.substr(i, 32);
                let hexArr = [];
                let asciiArr = [];
                for(let j = 0; j < chunk.length; j += 2) {
                    let h = chunk.substr(j, 2);
                    hexArr.push(h);
                    let code = parseInt(h, 16);
                    asciiArr.push((code >= 32 && code <= 126) ? String.fromCharCode(code) : ".");
                    if (j === 14) hexArr.push(""); // 额外空格分隔
                }
                let offsetStr = offset.toString(16).padStart(8, '0');
                lines.push(`${offsetStr}  ${hexArr.join(' ').padEnd(50, ' ')}  ${asciiArr.join('')}`);
                offset += 16;
            }
            return lines.join('\n');
        }

        if (format === 'C 数组') {
            let lines = [`char peer_${dir}_${index}[] = {`];
            let currentLine = [];
            for (let i = 0; i < hexStr.length; i += 2) {
                currentLine.push(`0x${hexStr.substr(i, 2)}`);
                if (currentLine.length === 8 || i >= hexStr.length - 2) {
                    lines.push("  " + currentLine.join(", ") + (i >= hexStr.length - 2 ? "" : ","));
                    currentLine = [];
                }
            }
            lines.push("};");
            return lines.join('\n');
        }

        // 默认 ASCII
        // 使用 Math.floor 兜底，防止万一出现奇数长度导致 new Array(小数) 崩溃
        let asciiChars = new Array(Math.floor(hexStr.length / 2));
        for (let i = 0, j = 0; i < hexStr.length; i += 2, j++) {
            let code = parseInt(hexStr.substr(i, 2), 16);
            asciiChars[j] = ((code >= 32 && code <= 126) || code === 10 || code === 13 || code === 9) ? String.fromCharCode(code) : '.';
        }
        return asciiChars.join('');
    }

    // ==== 导出功能：另存为 (导出全量数据，无视截断) ====
    function saveAs() {
        if (filteredPayloads.length === 0) {
            alert("没有数据可供保存");
            return;
        }

        // 注意：导出使用的是未截断的 filteredPayloads
        let contentArray = filteredPayloads.map((p, i) => formatData(p.hexData, displayFormat, p.dir, i));
        let content = contentArray.join(displayFormat === 'ASCII' ? '' : '\n\n');

        const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `stream_${streamId}_${displayFormat.replace(' ', '')}.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    function formatBytes(bytes: number) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
</script>

{#if show}
    <div class="modal-overlay" on:click={close}>
        <div class="modal-content" on:click|stopPropagation>
            <div class="modal-header">
                <h3>🌊 {title}</h3>
                <button class="close-btn" on:click={close}>✕</button>
            </div>

            <div class="modal-body">
                {#if isLoading}
                    <div class="loading-state">
                        <span class="spinner">🔃</span> 引擎正在重组应用层载荷...
                    </div>
                {:else}
                    {#if isTruncated}
                        <div class="truncation-warning">
                            ⚠️ <strong>载荷体积过大</strong>：为防止渲染卡顿，当前视图仅展示前 500 KB 数据。请点击右下角“另存为...”导出完整数据包。
                        </div>
                    {/if}

                    <div class="stream-text-area">
                        {#if displayedPayloads.length === 0}
                            <div style="color: #64748b; text-align: center; margin-top: 2rem;">没有提取到对应的载荷数据</div>
                        {/if}
                        {#each displayedPayloads as payload, i}
                            <span class="stream-chunk {payload.dir === 'client' ? 'client-text' : 'server-text'}">
                                {formatData(payload.hexData, displayFormat, payload.dir, i)}
                            </span>
                        {/each}
                    </div>
                {/if}
            </div>

            <div class="modal-footer-wrapper">
                <div class="control-bar">
                    <div class="control-item">
                        <select class="ui-select" bind:value={displayDirection}>
                            <option value="both">整个对话 ({formatBytes(clientBytes + serverBytes)})</option>
                            <option value="client">{clientNode} ➔ {serverNode} ({formatBytes(clientBytes)})</option>
                            <option value="server">{serverNode} ➔ {clientNode} ({formatBytes(serverBytes)})</option>
                        </select>
                    </div>
                    <div class="control-item">
                        <span class="ctrl-label">显示和保存数据为:</span>
                        <select class="ui-select" bind:value={displayFormat}>
                            <option value="ASCII">ASCII</option>
                            <option value="Hex 转储">Hex 转储</option>
                            <option value="原始数据">原始数据</option>
                            <option value="C 数组">C 数组</option>
                        </select>
                    </div>
                    <div class="control-item stream-navigator">
                        <span class="ctrl-label">流</span>
                        <div class="spinner-group">
                            <button class="spin-btn"
                                    on:click={() => dispatch('switchStream', streamId - 1)}
                                    disabled={streamId <= 0}>▼</button>

                            <input type="number" class="stream-input" min="0" bind:value={streamId} />

                            <button class="spin-btn"
                                    on:click={() => dispatch('switchStream', streamId + 1)}
                                    disabled={packetCount === 0}>▲</button>
                        </div>
                    </div>
                </div>

                <div class="action-bar">
                    <div class="legend">
                        <span class="dot client-dot"></span> 发起方 (Client)
                        <span class="dot server-dot" style="margin-left: 12px;"></span> 响应方 (Server)
                    </div>
                    <div class="action-btns">
                        <button class="action-btn" on:click={saveAs}>另存为...</button>
                        <button class="action-btn outline" on:click={close}>关闭</button>
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    /* ... 保持原有样式不变，并在此处新增警告框的样式 ... */
    .modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(15, 23, 42, 0.85); backdrop-filter: blur(4px); z-index: 9999; display: flex; align-items: center; justify-content: center; animation: fadeIn 0.2s ease-out; }
    .modal-content { background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 8px; width: 85vw; max-width: 1200px; height: 85vh; display: flex; flex-direction: column; box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5); animation: slideUp 0.3s ease-out; }
    @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
    @keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }

    .modal-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 20px; border-bottom: 1px solid var(--border-color); background: var(--bg-tertiary); border-radius: 8px 8px 0 0; }
    .modal-header h3 { margin: 0; color: var(--text-primary); font-size: 1.05rem; }
    .close-btn { background: transparent; border: none; color: var(--text-secondary); font-size: 1.2rem; cursor: pointer; transition: 0.2s; }
    .close-btn:hover { color: #ef4444; transform: scale(1.1); }

    .modal-body { flex: 1; padding: 12px 20px; overflow: hidden; display: flex; flex-direction: column; background: var(--bg-primary); gap: 8px; }
    .loading-state { margin: auto; color: var(--color-info); font-family: monospace; font-size: 1.1rem; }
    .spinner { display: inline-block; animation: spin 1s linear infinite; }
    @keyframes spin { 100% { transform: rotate(360deg); } }

    /* 【新增】：黄色截断警告条 */
    .truncation-warning {
        background: rgba(245, 158, 11, 0.1); border: 1px solid rgba(245, 158, 11, 0.3); color: #fcd34d;
        padding: 8px 12px; border-radius: 4px; font-size: 0.85rem; display: flex; align-items: center; gap: 8px;
    }

    .stream-text-area { flex: 1; overflow-y: auto; background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 4px; padding: 12px; font-family: Consolas, 'Courier New', monospace; font-size: 0.85rem; line-height: 1.4; white-space: pre-wrap; word-break: break-all; text-align: left; }
    .client-text { color: #cc0000; }
    .server-text { color: #0000cc; }

    .modal-footer-wrapper { display: flex; flex-direction: column; background: var(--bg-tertiary); border-top: 1px solid var(--border-color); border-radius: 0 0 8px 8px; }
    .control-bar { display: flex; justify-content: space-between; align-items: center; padding: 10px 20px; border-bottom: 1px solid var(--border-color); background: var(--bg-secondary); }
    .control-item { display: flex; align-items: center; gap: 8px; }
    .ctrl-label { font-size: 0.85rem; color: var(--text-secondary); }
    .ui-select { background: var(--bg-tertiary); border: 1px solid var(--border-color); color: var(--text-primary); padding: 4px 10px; border-radius: 4px; outline: none; font-size: 0.85rem; cursor: pointer; transition: 0.2s; }
    .ui-select:hover { border-color: var(--border-color-light); }

    .stream-navigator { display: flex; align-items: center; }
    .spinner-group { display: flex; align-items: center; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 4px; overflow: hidden; }
    .spin-btn { background: var(--bg-secondary); border: none; color: var(--text-secondary); cursor: pointer; padding: 4px 8px; font-size: 0.75rem; transition: 0.2s; }
    .spin-btn:disabled { color: var(--text-muted); cursor: not-allowed; }
    .spin-btn:hover:not(:disabled) { background: var(--border-color); color: var(--text-primary); }
    .stream-input { background: transparent; border: none; border-left: 1px solid var(--border-color); border-right: 1px solid var(--border-color); color: var(--text-primary); width: 45px; text-align: center; padding: 4px; font-size: 0.85rem; outline: none; }
    .stream-input::-webkit-inner-spin-button { -webkit-appearance: none; }

    .action-bar { display: flex; justify-content: space-between; align-items: center; padding: 10px 20px; }
    .legend { display: flex; align-items: center; font-size: 0.85rem; color: var(--text-secondary); }
    .dot { width: 10px; height: 10px; border-radius: 50%; display: inline-block; margin-right: 6px; }
    .client-dot { background: #cc0000; box-shadow: 0 0 4px rgba(204,0,0,0.5); }
    .server-dot { background: #0000cc; box-shadow: 0 0 4px rgba(0,0,204,0.5); }
    .action-btns { display: flex; gap: 10px; }
    .action-btn { background: var(--color-primary); border: none; color: white; padding: 6px 16px; border-radius: 4px; cursor: pointer; font-size: 0.85rem; transition: 0.2s; }
    .action-btn:hover { background: var(--color-primary-hover); }
    .action-btn.outline { background: transparent; border: 1px solid var(--border-color); color: var(--text-primary); }
    .action-btn.outline:hover { background: var(--bg-tertiary); color: var(--text-primary); }
    /* 将行内元素改为块级元素，强制换行，并增加段落间距 */
    .stream-chunk {
        display: block;
        margin-bottom: 12px;
    }

    .client-text { color: #cc0000; }
    .server-text { color: #0000cc; }
</style>