<script lang="ts">
    import { onMount } from 'svelte';
    import { GetPacketDetail, GetPacketsByPage, GetPacketHex, GetAllFrames, FollowStream } from '../../../wailsjs/go/main/App';
    import JsonTree from './JsonTree.svelte';
    import StreamModal from './StreamModal.svelte';

    export let file: any;

    // ==== 核心状态 ====
    let isLoadingData = false;
    let packets: any[] = [];
    let selectedDetail: any = null;
    let selectedRowIndex: number | null = null;
    $: selectedPacket = packets.find(p => p.index === selectedRowIndex);

    // ==== 过滤与分页状态 ====
    let globalFilter = "";
    let packetPage = 1;
    let packetPageSize = 10;
    let packetTotal = 0;
    $: packetTotalPages = packetTotal === -1 ? packetPage + 1 : (Math.ceil(packetTotal / packetPageSize) || 1);
    let jumpPageNum = 1;

    let filterProtocol = "";
    let filterIp = "";
    let filterInfo = "";

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

    // 初始化监听
    $: if (file) {
        packetPage = 1;
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
        selectedDetail = null;
        selectedRowIndex = null;
        loadAnalyzeData();
    }

    function clearGlobalFilter() {
        if (globalFilter === "") return;
        globalFilter = "";
        applyGlobalFilter();
    }

    // ==== 极速追踪流与重组逻辑 ====
    async function followStream(streamId: number, protocol: string) {
        if (streamId === null || streamId === undefined || isNaN(streamId)) return;

        streamCurrentId = streamId;
        streamCurrentProto = protocol;

        // 1. 触发底层列表过滤
        globalFilter = `${protocol}.stream == ${streamId}`;
        applyGlobalFilter();

        // 2. 呼出追踪流模态框
        const safeFileName = file && file.fileName ? file.fileName : 'Unknown File';
        streamModalTitle = `Follow ${protocol.toUpperCase()} Stream (${streamId})  |  Filter: ${globalFilter}  |  File: ${safeFileName}`;

        showStreamModal = true;
        isStreamLoading = true;
        streamPayloads = [];

        // 3. 重置统计元数据
        streamClientNode = "Client";
        streamServerNode = "Server";
        streamClientBytes = 0;
        streamServerBytes = 0;

        try {
            // 【核心优化】：将重负载全量拉取+海量解析的工作移交给 Go，仅获取聚合后的几 KB 成果！
            const resStr = await FollowStream(file.filePath, globalFilter, protocol);
            const streamData = JSON.parse(resStr);

            if (streamData) {
                streamClientNode = streamData.clientNode || "Client";
                streamServerNode = streamData.serverNode || "Server";
                streamClientBytes = streamData.clientBytes || 0;
                streamServerBytes = streamData.serverBytes || 0;
                streamPayloads = streamData.payloads || [{ dir: 'client', hexData: "" }];
            } else {
                streamPayloads = [{ dir: 'client', hexData: "" }];
            }
        } catch (err) {
            console.error("Failed to load stream data:", err);
            streamPayloads = [{ dir: 'client', hexData: "" }];
        } finally {
            isStreamLoading = false;
        }
    }

    function handleStreamModalClose() {
        showStreamModal = false;
        streamPayloads = [];
    }

    // ==== 数据加载核心逻辑 ====
    async function loadAnalyzeData() {
        if (!file) return;
        isLoadingData = true;
        packets = [];
        try {
            const resStr = await GetPacketsByPage(file.filePath, packetPage, packetPageSize, globalFilter);
            const resData = JSON.parse(resStr);

            if (resData) {
                packetTotal = resData.total !== undefined ? resData.total : 0;
                if (resData.list) {
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
                }
            }
        } catch (err) {
            console.error("加载数据包失败:", err);
        } finally {
            isLoadingData = false;
        }
    }

    function changePacketPage(page: number) {
        if (page >= 1 && page <= packetTotalPages) {
            packetPage = page;
            jumpPageNum = page;
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
                <span class="value highlight-blue" title={file?.filePath}>{file?.fileName}</span>
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
                <span class="label">Engine 追踪 ID</span>
                <span class="value sys-font">{file?.fileId}</span>
            </div>
        </div>
    </div>

    <div class="level-box level-2">
        <div class="level-header flex-between">
            <h3><span class="level-num">2</span> 文件帧提取 (Frames)
                <span class="sub-title">当前加载: {packets.length} 帧</span>
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
                <button class="action-btn" on:click={applyGlobalFilter}>查询</button>
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
                    引擎匹配: {packetTotal === -1 ? '计算中 (10000+)' : packetTotal} 帧
                    (第 {packetPage} / {packetTotal === -1 ? '?' : packetTotalPages} 页)
                </span>
                <select class="size-selector" bind:value={packetPageSize} on:change={handleSizeChange}>
                    <option value={10}>10 帧/页</option>
                    <option value={50}>50 帧/页</option>
                    <option value={100}>100 帧/页</option>
                    <option value={500}>500 帧/页</option>
                </select>
            </div>

            <div class="page-controls">
                <button disabled={packetPage <= 1} on:click={() => changePacketPage(1)}>首页</button>
                <button disabled={packetPage <= 1} on:click={() => changePacketPage(packetPage - 1)}>上一页</button>
                <span class="jump-box">
                    跳至 <input type="number" min="1" max={packetTotal === -1 ? 9999 : packetTotalPages} bind:value={jumpPageNum} on:keyup={(e) => e.key === 'Enter' && changePacketPage(jumpPageNum)} /> 页
                </span>
                <button disabled={packetPage >= packetTotalPages} on:click={() => changePacketPage(packetPage + 1)}>下一页</button>
                <button disabled={packetTotal === -1 || packetPage >= packetTotalPages} on:click={() => changePacketPage(packetTotalPages)}>尾页</button>
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
        padding-bottom: 24px;
        overflow-y: auto;
        padding-right: 8px;
    }

    .detail-container::-webkit-scrollbar { width: 8px; }
    .detail-container::-webkit-scrollbar-track { background: transparent; }
    .detail-container::-webkit-scrollbar-thumb { background: #334155; border-radius: 4px; }
    .detail-container::-webkit-scrollbar-thumb:hover { background: #475569; }

    .level-box {
        background: #111827;
        border: 1px solid #1e293b;
        border-radius: 8px;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        flex-shrink: 0;
    }

    .level-header {
        background: #0f172a;
        padding: 10px 16px;
        border-bottom: 1px solid #1e293b;
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .level-header.flex-between { justify-content: space-between; }
    .level-header h3 { margin: 0; font-size: 0.95rem; color: #f1f5f9; display: flex; align-items: center; gap: 8px; }
    .level-num { background: #4f46e5; color: white; width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; border-radius: 4px; font-size: 0.8rem; font-weight: bold; }
    .sub-title { font-size: 0.8rem; color: #64748b; font-weight: normal; margin-left: 8px; }

    /* ================= 层级 1: 基本信息 ================= */
    .level-1 { flex: none; }
    .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; padding: 16px; }
    .info-item { display: flex; flex-direction: column; gap: 4px; }
    .info-item .label { font-size: 0.75rem; color: #64748b; text-transform: uppercase; letter-spacing: 0.5px;}
    .info-item .value { font-size: 0.9rem; color: #e2e8f0; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;}
    .highlight-blue { color: #38bdf8 !important; }
    .sys-font { font-family: 'Fira Code', monospace; color: #94a3b8 !important;}
    .loading-spinner { color: #f59e0b; font-size: 0.8rem; font-weight: bold; }
    .status-badge.done { background: rgba(16, 185, 129, 0.1); color: #10b981; border: 1px solid rgba(16, 185, 129, 0.2); padding: 2px 8px; border-radius: 12px; font-size: 0.7rem;}

    /* ================= 层级 2: 文件帧流 ================= */
    .level-2 { min-height: 360px; flex: 1 0 auto; }
    .global-filter-bar { display: flex; align-items: center; gap: 8px; }
    .filter-label { font-size: 0.8rem; color: #38bdf8; font-weight: bold;}
    .filter-input-wrapper { position: relative; display: flex; align-items: center; }
    .filter-input-wrapper input { background: #1e293b; border: 1px solid #38bdf8; color: white; padding: 4px 28px 4px 8px; border-radius: 4px; outline: none; font-size: 0.8rem; width: 260px; transition: 0.2s; }
    .filter-input-wrapper input:focus { box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.3); }
    .clear-btn { position: absolute; right: 4px; background: transparent; border: none; color: #94a3b8; font-size: 1.2rem; cursor: pointer; line-height: 1; padding: 0 4px; transition: color 0.2s; }
    .clear-btn:hover { color: #f8fafc; }
    .action-btn { background: #1e293b; border: 1px solid #334155; color: white; padding: 4px 12px; border-radius: 4px; cursor: pointer; }
    .action-btn:hover { background: #3b82f6; }

    .packet-list-area { flex: 1; overflow: auto; }
    .data-table { width: 100%; border-collapse: collapse; text-align: left; font-size: 0.8rem; table-layout: fixed;}
    .data-table th { position: sticky; top: 0; background: #1e293b; padding: 6px 12px; color: #cbd5e1; z-index: 10; font-weight: normal;}
    .data-table td { padding: 4px 12px; border-bottom: 1px solid #1e293b; color: #e2e8f0; white-space: nowrap;}
    .data-table tbody tr:hover { background: #1f2937; cursor: pointer;}
    .active-row { background: rgba(79, 70, 229, 0.3) !important; }
    .info-cell { overflow: hidden; text-overflow: ellipsis; }
    .badge { background: #1e293b; padding: 2px 6px; border-radius: 4px; font-family: monospace; border: 1px solid #334155;}

    /* 分页器 */
    .pagination { display: flex; justify-content: space-between; align-items: center; padding: 6px 16px; background: #0f172a; border-top: 1px solid #1e293b; font-size: 0.8rem; color: #94a3b8; }
    .page-info { display: flex; align-items: center; gap: 12px; }
    .size-selector { background: #1e293b; border: 1px solid #334155; color: #cbd5e1; padding: 2px 6px; border-radius: 4px; outline: none; font-size: 0.75rem; cursor: pointer; }
    .page-controls { display: flex; align-items: center; gap: 6px; }
    .page-controls button { background: #1e293b; border: 1px solid #334155; color: white; padding: 4px 10px; border-radius: 4px; cursor: pointer; transition: 0.2s; }
    .page-controls button:hover:not(:disabled) { background: #3b82f6; border-color: #3b82f6; }
    .page-controls button:disabled { opacity: 0.4; cursor: not-allowed; }
    .jump-box { display: flex; align-items: center; gap: 4px; margin: 0 4px; }
    .jump-box input { background: #1e293b; border: 1px solid #334155; color: white; width: 40px; padding: 3px; border-radius: 4px; text-align: center; outline: none; font-size: 0.8rem; }
    .jump-box input:focus { border-color: #4f46e5; }
    .jump-box input::-webkit-outer-spin-button, .jump-box input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }

    /* ================= 层级 3: 帧系列透视 ================= */
    .level-3 { min-height: 550px; flex: 1 0 auto; }
    .detail-split-pane { display: flex; flex: 1; overflow: hidden; background: #0f172a;}
    .pane-box { display: flex; flex-direction: column; overflow: hidden; }
    .proto-tree { flex: 4; border-right: 1px solid #1e293b; }
    .hex-view { flex: 5; }
    .pane-title { padding: 6px 12px; background: #1e293b; font-size: 0.8rem; color: #94a3b8; border-bottom: 1px solid #334155; }
    .scroll-view { flex: 1; overflow: auto; padding: 12px; background: #111827;}
    .hex-scroll pre { white-space: pre; }
    pre { margin: 0; font-size: 0.8rem; color: #a5b4fc; font-family: 'Fira Code', Consolas, monospace; line-height: 1.4; }
    .placeholder { text-align: center; color: #475569; margin-top: 2rem; font-size: 0.85rem;}

    /* 追踪流按钮 */
    .stream-btn {
        background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.3); color: #38bdf8;
        padding: 2px 10px; border-radius: 4px; cursor: pointer; font-size: 0.75rem; margin-left: 16px;
        font-family: 'Fira Code', monospace; transition: all 0.2s ease-out;
    }
    .stream-btn:hover { background: rgba(56, 189, 248, 0.2); border-color: #38bdf8; transform: translateY(-1px); box-shadow: 0 2px 4px rgba(56, 189, 248, 0.2); }
</style>