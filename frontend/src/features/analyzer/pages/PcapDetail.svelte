<script lang="ts">
    import { onMount } from 'svelte';
    import { GetPacketDetail, GetPacketsByPage, GetPacketHex, GetAllFrames, FollowStream } from '../../../../wailsjs/go/main/App';
    import JsonTree from '../components/JsonTree.svelte';
    import StreamModal from '../components/StreamModal.svelte';
    import { success as showSuccess } from '../../../stores/toast';
    import { logger } from '../../../utils/logger';
    import type { PcapFile } from '../../../types';

    export let file: PcapFile;

    let isLoadingData = false;
    let packets: any[] = [];
    let selectedDetail: any = null;
    let selectedRowIndex: number | null = null;
    $: selectedPacket = packets.find(p => p.index === selectedRowIndex);

    let globalFilter = '';
    let packetPage = 1;
    let packetPageSize = 10;
    let hasMore = true;

    let filterProtocol = '';
    let filterIp = '';
    let filterInfo = '';

    $: filteredPackets = packets.filter(p => {
        const matchProto = filterProtocol === "" || p.protocol.toLowerCase().includes(filterProtocol.toLowerCase());
        const matchIp = filterIp === "" || p.source.includes(filterIp) || p.dest.includes(filterIp);
        const matchInfo = filterInfo === "" || p.info.toLowerCase().includes(filterInfo.toLowerCase());
        return matchProto && matchIp && matchInfo;
    });

    // ==== 追踪流模态框状态 ====
    let showStreamModal = false;
    let streamModalTitle = "";
    let isStreamLoading = false;
    let streamPayloads: { dir: 'client' | 'server', hexData: string }[] = [];
    let streamCurrentId = 0;
    let streamCurrentProto = "tcp";
    let streamClientNode = "Client";
    let streamServerNode = "Server";
    let streamClientBytes = 0;
    let streamServerBytes = 0;
    let streamPacketCount = 0;

    // 初始化监听
    $: if (file) {
        packetPage = 1;
        hasMore = true;
        selectedDetail = null;
        selectedRowIndex = null;
        resetFilters();
        loadAnalyzeData();
    }

    function resetFilters() {
        filterProtocol = "";
        filterIp = "";
        filterInfo = "";
    }

    function applyGlobalFilter() {
        packetPage = 1;
        hasMore = true;
        selectedDetail = null;
        selectedRowIndex = null;
        loadAnalyzeData();
    }

    function clearGlobalFilter() {
        if (globalFilter === "") return;
        globalFilter = "";
        applyGlobalFilter();
    }

    async function copyToClipboard(text: string) {
        if (!text) return;
        try {
            await navigator.clipboard.writeText(text);
            showSuccess('已复制到剪贴板');
        } catch (err) {
            logger.error('无法复制:', err);
        }
    }

    // ==== 极速追踪流与重组逻辑 ====
    async function followStream(streamId: number, protocol: string) {
        if (streamId === null || streamId === undefined || isNaN(streamId)) return;

        streamCurrentId = streamId;
        streamCurrentProto = protocol;

        globalFilter = `${protocol}.stream == ${streamId}`;
        applyGlobalFilter();

        const safeFileName = file && file.fileName ? file.fileName : 'Unknown File';
        streamModalTitle = `Follow ${protocol.toUpperCase()} Stream (${streamId})  |  Filter: ${globalFilter}  |  File: ${safeFileName}`;

        showStreamModal = true;
        isStreamLoading = true;
        streamPayloads = [];

        streamClientNode = 'Client';
        streamServerNode = 'Server';
        streamClientBytes = 0;
        streamServerBytes = 0;
        streamPacketCount = 0;

        try {
            const resStr = await FollowStream(file.filePath, globalFilter, protocol);
            const streamData = JSON.parse(resStr);

            if (streamData) {
                streamClientNode = streamData.clientNode || 'Client';
                streamServerNode = streamData.serverNode || 'Server';
                streamClientBytes = streamData.clientBytes || 0;
                streamServerBytes = streamData.serverBytes || 0;
                streamPacketCount = streamData.packetCount || 0;
                streamPayloads = streamData.payloads || [{ dir: 'client', hexData: '' }];
            } else {
                streamPayloads = [{ dir: 'client', hexData: '' }];
            }
        } catch (err) {
            logger.error('Failed to load stream data:', err);
            streamPayloads = [{ dir: 'client', hexData: '' }];
        } finally {
            isStreamLoading = false;
        }
    }

    function handleStreamModalClose() {
        showStreamModal = false;
        streamPayloads = [];
    }

    // ==== 数据加载核心逻辑 (✨ 核心修改区) ====
    async function loadAnalyzeData() {
        if (!file) return;
        isLoadingData = true;

        try {
            const resStr = await GetPacketsByPage(file.filePath, packetPage, packetPageSize, globalFilter);
            const resData = JSON.parse(resStr);

            if (resData) {
                const backendHasMore = resData.has_more ?? resData.hasMore ?? resData.HasMore;
                hasMore = backendHasMore === true;

                if (resData.list && resData.list.length > 0) {
                    packets = resData.list.map((frame: any, idx: number) => {
                        const bl = frame.BaseLayers || {};
                        const col = bl.WsCol || {};
                        const frm = bl.Frame || {};

                        const tcp = bl.Tcp || null;
                        const udp = bl.Udp || null;

                        let timeStr = frm['frame.time'] || '-';
                        if (timeStr.includes('T')) timeStr = timeStr.split('T')[1].replace('Z', '');

                        return {
                            _rawId: idx,
                            index: frm['frame.number'] || col['_ws.col.number'] || (idx + 1),
                            time: timeStr,
                            source: col['_ws.col.def_src'] || '-',
                            dest: col['_ws.col.def_dst'] || '-',
                            protocol: col['_ws.col.protocol'] || '-',
                            info: col['_ws.col.info'] || '-',
                            tcpStream: tcp ? tcp['tcp.stream'] : null,
                            udpStream: udp ? udp['udp.stream'] : null
                        };
                    });

                    if (packets.length > 0) {
                        selectPacket(packets[0].index);
                    }
                } else {
                    // 如果 list 为空，说明真的没数据了
                    packets = [];
                    hasMore = false;
                }
            }
        } catch (err) {
            console.error("加载数据包失败:", err);
            hasMore = false;
        } finally {
            isLoadingData = false;
        }
    }

    function changePacketPage(page: number) {
        if (page >= 1) {
            packetPage = page;
            selectedDetail = null;
            selectedRowIndex = null;
            resetFilters();
            loadAnalyzeData();
        }
    }

    function handleSizeChange() {
        packetPage = 1;
        changePacketPage(1);
    }

    async function selectPacket(index: number) {
        selectedRowIndex = index;
        try {
            const [treeResStr, hexResStr] = await Promise.all([
                GetPacketDetail(file.filePath, index),
                GetPacketHex(file.filePath, index)
            ]);
            const treeRes = JSON.parse(treeResStr);
            const hexRes = JSON.parse(hexResStr);

            let formattedHex = "未获取到 Hex 数据";
            if (hexRes && hexRes.offset) {
                let lines = [];
                for (let i = 0; i < hexRes.offset.length; i++) {
                    const offset = (hexRes.offset[i] || "").padEnd(6, " ");
                    const hexStr = (hexRes.hex[i] || "").padEnd(49, " ");
                    const ascii = hexRes.ascii[i] || "";
                    lines.push(`${offset}  ${hexStr}  ${ascii}`);
                }
                formattedHex = lines.join('\n');
            }

            if (treeRes && treeRes.list && treeRes.list.length > 0) {
                const frameData = treeRes.list[0];
                selectedDetail = {
                    tree: frameData.BaseLayers || frameData,
                    hex: formattedHex
                };
            }
        } catch (err) {
            console.error("获取包详情失败:", err);
        }
    }

    function formatDate(dateStr: string) {
        if (!dateStr) return "-";
        return new Date(dateStr).toLocaleString('zh-CN', {hour12: false});
    }
</script>

<div class="detail-container">

    <div class="level-box level-1">
        <div class="level-header">
            <h3><span class="level-num">1</span> 文件基本信息 (Metadata)</h3>
            {#if isLoadingData}
                <span class="loading-spinner">🔃 引擎抽取中...</span>
            {:else}
                <span class="status-badge done">引擎就绪</span>
            {/if}
        </div>
        <div class="info-grid">
            <div class="info-item">
                <span class="label">正在分析文件</span>
                <span
                        class="value highlight-blue"
                        style="cursor: pointer;"
                        title={file?.filePath}
                        on:click={() => copyToClipboard(file?.filePath)}
                >
                    {file?.fileName}
                </span>
            </div>
            <div class="info-item">
                <span class="label">文件物理大小</span>
                <span class="value">{file?.fileSize}</span>
            </div>
            <div class="info-item">
                <span class="label">入库时间</span>
                <span class="value">{formatDate(file?.createdAt)}</span>
            </div>
            <div class="info-item">
                <span class="label">ID</span>
                <span class="value sys-font">{file?.fileId}</span>
            </div>
        </div>
    </div>

    <div class="level-box level-2">
        <div class="level-header flex-between">
            <h3><span class="level-num">2</span> 文件帧提取 (Frames)
                <span class="sub-title">当前页渲染: {packets.length} 帧</span>
            </h3>

            <div class="global-filter-bar">
                <span class="filter-label" title="支持 Wireshark Display Filter 语法，例如 tcp.port == 80">
                    🦈 Filter:
                </span>
                <div class="filter-input-wrapper">
                    <input type="text" bind:value={globalFilter}
                           on:keyup={(e) => e.key === 'Enter' && applyGlobalFilter()}
                           placeholder="如: ip.addr==1.1.1.1 或 tcp" />
                    {#if globalFilter !== ""}
                        <button class="clear-btn" title="清空过滤器" on:click={clearGlobalFilter}>×</button>
                    {/if}
                </div>
                <button class="action-btn" on:click={applyGlobalFilter} disabled={isLoadingData}>查询</button>
            </div>
        </div>

        <div class="packet-list-area">
            <table class="data-table">
                <thead>
                <tr>
                    <th style="width: 60px;">No.</th>
                    <th style="width: 140px;">Time</th>
                    <th style="width: 140px;">Source</th>
                    <th style="width: 140px;">Dest</th>
                    <th style="width: 90px;">Protocol</th>
                    <th>Info</th>
                </tr>
                </thead>
                <tbody>
                {#each filteredPackets as p (p.index || p._rawId)}
                    <tr class:active-row={selectedRowIndex === p.index} on:click={() => selectPacket(p.index)}>
                        <td>{p.index}</td>
                        <td class="time-col">{p.time}</td>
                        <td>{p.source}</td>
                        <td>{p.dest}</td>
                        <td><span class="badge">{p.protocol}</span></td>
                        <td class="info-cell" title={p.info}>{p.info}</td>
                    </tr>
                {/each}
                {#if filteredPackets.length === 0 && !isLoadingData}
                    <tr><td colspan="6" style="text-align: center; padding: 2rem; color: #64748b;">无匹配的数据包</td></tr>
                {/if}
                </tbody>
            </table>
        </div>

        <div class="pagination">
            <div class="page-info">
                <span>
                    当前第 <strong style="color: #38bdf8;">{packetPage}</strong> 页
                    {#if !hasMore} <span style="color: #94a3b8; font-size: 0.75rem;">(已到底部)</span> {/if}
                </span>
                <select class="size-selector" bind:value={packetPageSize} on:change={handleSizeChange}>
                    <option value={10}>10 帧/页</option>
                    <option value={50}>50 帧/页</option>
                    <option value={100}>100 帧/页</option>
                    <option value={500}>500 帧/页</option>
                </select>
            </div>

            <div class="page-controls">
                <button
                    disabled={packetPage <= 1 || isLoadingData}
                    on:click={() => changePacketPage(1)}
                >首页</button>

                <button
                    disabled={packetPage <= 1 || isLoadingData}
                    on:click={() => changePacketPage(packetPage - 1)}
                >上一页</button>

                <button
                    disabled={!hasMore || isLoadingData}
                    on:click={() => changePacketPage(packetPage + 1)}
                >下一页</button>
            </div>
        </div>
    </div>

    <div class="level-box level-3">
        <div class="level-header">
            <h3><span class="level-num">3</span> 帧系列透视 (Series)
                <span class="sub-title">当前选中帧: {selectedRowIndex || '?'}</span>

                {#if selectedPacket}
                    {#if selectedPacket.tcpStream !== null && selectedPacket.tcpStream !== undefined}
                        <button class="stream-btn" on:click={() => followStream(selectedPacket.tcpStream, 'tcp')}>
                            🌊 追踪 TCP 流 ({selectedPacket.tcpStream})
                        </button>
                    {:else if selectedPacket.udpStream !== null && selectedPacket.udpStream !== undefined}
                        <button class="stream-btn" on:click={() => followStream(selectedPacket.udpStream, 'udp')}>
                            🌊 追踪 UDP 流 ({selectedPacket.udpStream})
                        </button>
                    {/if}
                {/if}
            </h3>
        </div>

        <div class="detail-split-pane">
            <div class="pane-box proto-tree">
                <div class="pane-title">协议树 (Protocol Tree)</div>
                <div class="scroll-view">
                    {#if selectedDetail}
                        <JsonTree data={selectedDetail.tree} keyName="Frame Data" defaultExpanded={true} />
                    {:else}
                        <p class="placeholder">正在加载数据...</p>
                    {/if}
                </div>
            </div>

            <div class="pane-box hex-view">
                <div class="pane-title">底层 Hex 流 (Hex Dump)</div>
                <div class="scroll-view hex-scroll">
                    {#if selectedDetail}
                        <pre>{selectedDetail.hex}</pre>
                    {:else}
                        <p class="placeholder">正在加载数据...</p>
                    {/if}
                </div>
            </div>
        </div>
    </div>
</div>

<StreamModal
        bind:show={showStreamModal}
        title={streamModalTitle}
        isLoading={isStreamLoading}
        payloads={streamPayloads}
        streamId={streamCurrentId}
        clientNode={streamClientNode}
        serverNode={streamServerNode}
        clientBytes={streamClientBytes}
        serverBytes={streamServerBytes}
        packetCount={streamPacketCount}
        on:close={handleStreamModalClose}
        on:switchStream={(e) => followStream(e.detail, streamCurrentProto)}
/>

<style>
    /* ================= 布局与容器 ================= */
    .detail-container {
        display: flex;
        flex-direction: column;
        height: 100%;
        gap: 20px;
        padding: 1.5rem;
        padding-bottom: 24px;
        overflow-y: auto;
        padding-right: 8px;
        background-color: var(--bg-primary);
    }

    .detail-container::-webkit-scrollbar { width: 8px; }
    .detail-container::-webkit-scrollbar-track { background: transparent; }
    .detail-container::-webkit-scrollbar-thumb { background: var(--border-color); border-radius: 4px; }
    .detail-container::-webkit-scrollbar-thumb:hover { background: var(--border-color-light); }

    .level-box {
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        flex-shrink: 0;
    }

    .level-header {
        background: var(--bg-tertiary);
        padding: 10px 16px;
        border-bottom: 1px solid var(--border-color);
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .level-header.flex-between { justify-content: space-between; }
    .level-header h3 { margin: 0; font-size: 0.95rem; color: var(--text-primary); display: flex; align-items: center; gap: 8px; }
    .level-num { background: var(--color-primary); color: white; width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; border-radius: 4px; font-size: 0.8rem; font-weight: bold; }
    .sub-title { font-size: 0.8rem; color: var(--text-muted); font-weight: normal; margin-left: 8px; }

    /* ================= 层级 1: 基本信息 ================= */
    .level-1 { flex: none; }
    .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; padding: 16px; }
    .info-item { display: flex; flex-direction: column; gap: 4px; }
    .info-item .label { font-size: 0.75rem; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.5px;}
    .info-item .value { font-size: 0.9rem; color: var(--text-primary); font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;}
    .highlight-blue { color: var(--color-info) !important; }
    .sys-font { font-family: 'Fira Code', monospace; color: var(--text-tertiary) !important;}
    .loading-spinner { color: #f59e0b; font-size: 0.8rem; font-weight: bold; }
    .status-badge.done { background: rgba(16, 185, 129, 0.1); color: #10b981; border: 1px solid rgba(16, 185, 129, 0.2); padding: 2px 8px; border-radius: 12px; font-size: 0.7rem;}

    /* ================= 层级 2: 文件帧流 ================= */
    .level-2 { min-height: 360px; flex: 1 0 auto; }
    .global-filter-bar { display: flex; align-items: center; gap: 8px; }
    .filter-label { font-size: 0.8rem; color: var(--color-info); font-weight: bold;}
    .filter-input-wrapper { position: relative; display: flex; align-items: center; }
    .filter-input-wrapper input { background: var(--bg-tertiary); border: 1px solid var(--color-info); color: var(--text-primary); padding: 4px 28px 4px 8px; border-radius: 4px; outline: none; font-size: 0.8rem; width: 260px; transition: 0.2s; }
    .filter-input-wrapper input:focus { box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.3); }
    .clear-btn { position: absolute; right: 4px; background: transparent; border: none; color: var(--text-tertiary); font-size: 1.2rem; cursor: pointer; line-height: 1; padding: 0 4px; transition: color 0.2s; }
    .clear-btn:hover { color: var(--text-primary); }
    .action-btn { background: var(--bg-tertiary); border: 1px solid var(--border-color); color: var(--text-primary); padding: 4px 12px; border-radius: 4px; cursor: pointer; }
    .action-btn:hover:not(:disabled) { background: var(--color-primary); border-color: var(--color-primary); color: white; }
    .action-btn:disabled { opacity: 0.5; cursor: not-allowed; }

    .packet-list-area { flex: 1; overflow: auto; }
    .data-table { width: 100%; border-collapse: collapse; text-align: left; font-size: 0.8rem; table-layout: fixed;}
    .data-table th { position: sticky; top: 0; background: var(--bg-tertiary); padding: 6px 12px; color: var(--text-primary); z-index: 10; font-weight: normal;}
    .data-table td { padding: 4px 12px; border-bottom: 1px solid var(--border-color); color: var(--text-secondary); white-space: nowrap;}
    .data-table tbody tr:hover { background: var(--bg-tertiary); cursor: pointer;}
    .active-row { background: rgba(79, 70, 229, 0.3) !important; }
    .info-cell { overflow: hidden; text-overflow: ellipsis; }
    .badge { background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px; font-family: monospace; border: 1px solid var(--border-color);}

    /* 分页器 (✨ 更新样式) */
    .pagination { display: flex; justify-content: space-between; align-items: center; padding: 6px 16px; background: var(--bg-tertiary); border-top: 1px solid var(--border-color); font-size: 0.85rem; color: var(--text-secondary); }
    .page-info { display: flex; align-items: center; gap: 12px; }
    .size-selector { background: var(--bg-tertiary); border: 1px solid var(--border-color); color: var(--text-primary); padding: 2px 6px; border-radius: 4px; outline: none; font-size: 0.75rem; cursor: pointer; }
    .page-controls { display: flex; align-items: center; gap: 6px; }
    .page-controls button { background: var(--bg-tertiary); border: 1px solid var(--border-color); color: var(--text-primary); padding: 4px 12px; border-radius: 4px; cursor: pointer; transition: 0.2s; font-size: 0.8rem;}
    .page-controls button:hover:not(:disabled) { background: var(--color-primary); border-color: var(--color-primary); color: white; }
    .page-controls button:disabled { opacity: 0.4; cursor: not-allowed; }

    /* ================= 层级 3: 帧系列透视 ================= */
    .level-3 { min-height: 550px; flex: 1 0 auto; }
    .detail-split-pane { display: flex; flex: 1; overflow: hidden; background: var(--bg-primary);}
    .pane-box { display: flex; flex-direction: column; overflow: hidden; }
    .proto-tree { flex: 4; border-right: 1px solid var(--border-color); }
    .hex-view { flex: 5; }
    .pane-title { padding: 6px 12px; background: var(--bg-tertiary); font-size: 0.8rem; color: var(--text-secondary); border-bottom: 1px solid var(--border-color); }
    .scroll-view { flex: 1; overflow: auto; padding: 12px; background: var(--bg-secondary);}
    .hex-scroll pre { white-space: pre; }
    pre { margin: 0; font-size: 0.8rem; color: var(--text-tertiary); font-family: 'Fira Code', Consolas, monospace; line-height: 1.4; }
    .placeholder { text-align: center; color: var(--text-muted); margin-top: 2rem; font-size: 0.85rem;}

    /* 追踪流按钮 */
    .stream-btn {
        background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.3); color: #38bdf8;
        padding: 2px 10px; border-radius: 4px; cursor: pointer; font-size: 0.75rem; margin-left: 16px;
        font-family: 'Fira Code', monospace; transition: all 0.2s ease-out;
    }
    .stream-btn:hover { background: rgba(56, 189, 248, 0.2); border-color: #38bdf8; transform: translateY(-1px); box-shadow: 0 2px 4px rgba(56, 189, 248, 0.2); }
</style>