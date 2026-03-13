<script lang="ts">
    import { createEventDispatcher, onDestroy, onMount } from 'svelte';
    import { EventsOff, EventsOn, OnFileDrop, OnFileDropOff } from '../../../../wailsjs/runtime/runtime';
    import {
        BatchDeleteFiles,
        DeleteFile,
        GetFileList,
        ImportFromPaths,
        ImportPcapsDialog
    } from '../../../../wailsjs/go/main/App';
    import { error as showError, success as showSuccess, warning as showWarning } from '../../../stores/toast';
    import { debounce } from '../../../utils/helpers';

    const dispatch = createEventDispatcher();

    let pcapFiles: any[] = [];
    let searchFileName = '';
    let searchFileSize = '';
    let searchStartDate = '';
    let searchEndDate = '';
    let currentPage = 1;
    let pageSize = 10;
    let totalCount = 0;
    $: totalPages = Math.ceil(totalCount / pageSize) || 1;

    let selectedIds: number[] = [];
    $: isAllSelected = pcapFiles.length > 0 && selectedIds.length === pcapFiles.length;

    let uploadProgressMap: Record<string, { fileName: string; percent: number }> = {};
    let isDragging = false;

    const debouncedRefresh = debounce(async () => {
        currentPage = 1;
        await refreshFileList();
    }, 300);

    onMount(async () => {
        await refreshFileList();

        EventsOn("pcap:progress", (data: any) => {
            const targetIndex = pcapFiles.findIndex(f => f.fileId === data.fileId);

            if (targetIndex !== -1) {
                if (data.percent === -1) {
                    pcapFiles[targetIndex].status = "导入失败";
                    pcapFiles[targetIndex]._error = data.error;
                    showError(`文件 [${data.fileName}] 导入失败：${data.error}`);
                } else if (data.percent === 100) {
                    pcapFiles[targetIndex].status = "导入成功";
                    pcapFiles[targetIndex]._progress = undefined;
                } else {
                    pcapFiles[targetIndex].status = "导入中";
                    pcapFiles[targetIndex]._progress = data.percent;
                }
                pcapFiles = [...pcapFiles];
            }
        });

        let isProcessingDrop = false;
        OnFileDrop((x, y, paths) => {
            isDragging = false;
            if (isProcessingDrop) return;
            isProcessingDrop = true;
            const validPaths = paths.filter(p => p.endsWith('.pcap') || p.endsWith('.pcapng'));
            if (validPaths.length === 0) {
                showWarning("仅支持拖入 .pcap 或 .pcapng 流量包！");
                isProcessingDrop = false;
                return;
            }

            ImportFromPaths(validPaths).then((newFiles) => {
                if (newFiles && newFiles.length > 0) {
                    pcapFiles = [...newFiles, ...pcapFiles];
                }
                isProcessingDrop = false;
            }).catch(err => {
                showError("拖拽导入失败：" + err);
                isProcessingDrop = false;
            });
        }, false);
    });

    onDestroy(() => {
        EventsOff("pcap:progress");
        OnFileDropOff();
    });

    function handleGlobalDragOver(e: DragEvent) {
        e.preventDefault();
        isDragging = true;
    }

    function handleGlobalDrop(e: DragEvent) {
        e.preventDefault();
        isDragging = false;
    }

    function handleGlobalDragLeave(e: DragEvent) {
        e.preventDefault();
        if (!e.relatedTarget) isDragging = false;
    }

    async function handleImportBtn() {
        try {
            await ImportPcapsDialog();
            await refreshFileList();
        } catch (err) {
            showError("导入失败：" + err);
        }
    }

    async function refreshFileList() {
        try {
            const resp = await GetFileList({
                fileName: searchFileName,
                fileSize: searchFileSize,
                startDate: searchStartDate,
                endDate: searchEndDate,
                page: currentPage,
                pageSize: pageSize
            });
            pcapFiles = resp.list || [];
            totalCount = resp.total || 0;
        } catch (err) {
            console.error("加载列表失败:", err);
        }
    }

    function handleSearch() {
        currentPage = 1;
        selectedIds = [];
        refreshFileList();
    }

    function handleReset() {
        searchFileName = "";
        searchFileSize = "";
        searchStartDate = "";
        searchEndDate = "";
        currentPage = 1;
        selectedIds = [];
        refreshFileList();
    }

    function changePage(page: number) {
        if (page >= 1 && page <= totalPages) {
            currentPage = page;
            selectedIds = [];
            refreshFileList();
        }
    }

    function toggleSelectAll() {
        selectedIds = isAllSelected ? [] : pcapFiles.map(f => f.id);
    }

    async function handleBatchDelete() {
        if (!confirm(`确定批量删除 ${selectedIds.length} 个流量包？`)) return;
        try {
            await BatchDeleteFiles(selectedIds);
            selectedIds = [];
            await refreshFileList();
        } catch (err) {
            showError("批量删除失败: " + err);
        }
    }

    async function handleDelete(id: number) {
        if (!confirm("确定要删除该记录及 PCAP 吗？")) return;
        try {
            await DeleteFile(id);
            await refreshFileList();
        } catch (err) {
            alert("删除失败: " + err);
        }
    }

    function formatDate(dateStr: string) {
        return !dateStr ? "-" : new Date(dateStr).toLocaleString('zh-CN', {hour12: false});
    }

    let copiedFileId: number | null = null;

    async function copyFileName(e: Event, file: any) {
        e.stopPropagation(); // 必须阻止冒泡，否则会触发单元格的点击选中复选框事件
        try {
            await navigator.clipboard.writeText(file.fileName);
            copiedFileId = file.id; // 记录当前复制成功的行 ID

            // 2 秒后恢复图标
            setTimeout(() => {
                if (copiedFileId === file.id) {
                    copiedFileId = null;
                }
            }, 2000);
        } catch (err) {
            console.error("复制失败:", err);
            alert("复制失败，请检查浏览器剪贴板权限");
        }
    }
</script>

<svelte:window on:dragover={handleGlobalDragOver} on:dragleave={handleGlobalDragLeave} on:drop={handleGlobalDrop}/>

<div class="list-container">
    {#if isDragging}
        <div class="drag-overlay">
            <div class="drag-content"><span class="icon">📥</span>
                <h2>松开鼠标导入 PCAP</h2></div>
        </div>
    {/if}

    <div class="filter-panel">
        <div class="filter-group"><label for="sName">文件名:</label><input id="sName" type="text"
                                                                           bind:value={searchFileName}
                                                                           on:input={debouncedRefresh}
                                                                           placeholder="搜文件名"/></div>
        <div class="filter-group"><label for="sSize">大小:</label><input id="sSize" type="text"
                                                                             bind:value={searchFileSize}
                                                                             on:keyup={(e) => e.key === 'Enter' && handleSearch()}
                                                                             placeholder="如 1.5 MB"/></div>
        <div class="filter-group"><label for="sDate">时间:</label><input id="sDate" type="date"
                                                                         bind:value={searchStartDate}/><span
                class="separator">-</span><input type="date" bind:value={searchEndDate}/></div>
        <div class="filter-actions">
            <button class="action-btn" on:click={handleSearch}>查询</button>
            <button class="action-btn outline" on:click={handleReset}>重置</button>
        </div>
        <div class="import-action">
            <button class="primary-btn" on:click={handleImportBtn}>+ 导入本地文件</button>
        </div>
    </div>

    {#if selectedIds.length > 0}
        <div class="batch-action-bar">
            <span>已选择 <strong>{selectedIds.length}</strong> 个流量包</span>
            <div class="batch-btns">
                <button class="batch-btn danger" on:click={handleBatchDelete}>🗑️ 批量删除</button>
                <button class="batch-btn" on:click={() => selectedIds = []}>取消选择</button>
            </div>
        </div>
    {/if}

    {#if Object.keys(uploadProgressMap).length > 0}
        <div class="progress-panel">
            <h4>正在拷贝文件入库...</h4>
            {#each Object.entries(uploadProgressMap) as [id, info]}
                <div class="progress-item">
                    <div class="prog-meta"><span class="prog-name">{info.fileName}</span><span>{info.percent}%</span>
                    </div>
                    <div class="prog-bar-bg">
                        <div class="prog-bar-fill" style="width: {info.percent}%"></div>
                    </div>
                </div>
            {/each}
        </div>
    {/if}

    <div class="table-wrapper">
        <table class="data-table">
            <thead>
            <tr>
                <th style="width: 40px; text-align: center;"><input type="checkbox" checked={isAllSelected}
                                                                    on:change={toggleSelectAll} class="checkbox-ui"/>
                </th>
                <th>文件名</th>
                <th style="width: 50px;">大小</th>
                <th style="width: 120px;">上传时间</th>
                <th style="width: 80px;">状态</th>
                <th class="sticky-col-header" style="width: 120px;">操作</th>
            </tr>
            </thead>
            <tbody>
            {#each pcapFiles as file}
                <tr class:selected-row={selectedIds.includes(file.id)}>

                    <td style="text-align: center;" on:click|stopPropagation>
                        <input type="checkbox"
                               bind:group={selectedIds}
                               value={file.id}
                               class="checkbox-ui"/>
                    </td>

                    <td class="filename-cell" title={file.fileName}
                        on:click={() => document.querySelector(`input[value="${file.id}"]`)?.click()}>

                        <div class="filename-layout">
                            <div class="filename-text">{file.fileName}</div>

                            <div class="copy-btn"
                                 on:click={(e) => copyFileName(e, file)}
                                 title="复制文件名">
                                {copiedFileId === file.id ? '✅' : '📋'}
                            </div>
                        </div>

                        <div class="file-id">{file.fileId}</div>
                    </td>

                    <td>{file.fileSize}</td>
                    <td class="time-col">{formatDate(file.createdAt)}</td>
                    <td class="status-cell">
                        {#if file.status === '导入中' && file._progress !== undefined}
                            <div class="inline-progress">
                                <div class="inline-meta">导入中 {file._progress}%</div>
                                <div class="inline-bar-bg">
                                    <div class="inline-bar-fill" style="width: {file._progress}%"></div>
                                </div>
                            </div>
                        {:else}
                            <span class={`status-dot
                                ${file.status === '导入失败' ? 'error' : ''}
                                ${file.status === '导入成功' ? 'done' : ''}
                            `}></span>
                            <span class:text-error={file.status === '导入失败'}
                                  title={file._error || ""}>{file.status}</span>
                        {/if}
                    </td>
                    <td class="sticky-col-body">
                        <button class="action-btn" on:click={() => dispatch('analyze', file)}>详情</button>
                        <button class="action-btn danger" on:click={() => handleDelete(file.id)}>删除</button>
                    </td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>

    <div class="pagination">
        <span>共计 {totalCount} 个流量包记录 (第 {currentPage} / {totalPages} 页)</span>
        <div class="page-controls">
            <button disabled={currentPage <= 1} on:click={() => changePage(currentPage - 1)}>上一页</button>
            <button disabled={currentPage >= totalPages} on:click={() => changePage(currentPage + 1)}>下一页</button>
        </div>
    </div>
</div>

<style>
    .list-container {
        height: 100%;
        display: flex;
        flex-direction: column;
        position: relative;
    }

    .drag-overlay {
        position: absolute;
        inset: 0;
        background: rgba(15, 23, 42, 0.9);
        z-index: 999;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 3px dashed #6366f1;
        border-radius: 8px;
    }

    .drag-content {
        text-align: center;
        color: white;
    }

    .filter-panel {
        display: flex;
        flex-wrap: wrap;
        gap: 16px;
        align-items: center;
        padding: 12px;
        background: #111827;
        border: 1px solid #1e293b;
        border-radius: 8px;
        margin-bottom: 12px;
    }

    .filter-group {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .filter-group label {
        color: #94a3b8;
        font-size: 0.85rem;
    }

    .filter-group input {
        background: #1e293b;
        border: 1px solid #334155;
        color: white;
        padding: 6px 10px;
        border-radius: 6px;
        outline: none;
        font-size: 0.85rem;
    }

    .import-action {
        margin-left: auto;
    }

    .primary-btn {
        background: #4f46e5;
        color: white;
        border: none;
        padding: 8px 14px;
        border-radius: 6px;
        cursor: pointer;
        transition: 0.2s;
    }

    .primary-btn:hover {
        background: #4338ca;
    }

    .action-btn {
        background: #1e293b;
        border: 1px solid #334155;
        color: white;
        padding: 6px 12px;
        border-radius: 4px;
        cursor: pointer;
    }

    .action-btn.danger:hover {
        background: #ef4444;
        border-color: #ef4444;
    }

    .action-btn:hover {
        background: #3b82f6;
    }

    .table-wrapper {
        flex: 1;
        overflow: auto;
        border: 1px solid #1e293b;
        border-radius: 8px 8px 0 0;
        background: #111827;
    }

    .data-table {
        width: 100%;
        min-width: 800px;
        border-collapse: collapse;
        text-align: left;
        font-size: 0.85rem;
        table-layout: fixed;
    }

    .data-table th {
        position: sticky;
        top: 0;
        background: #1e293b;
        padding: 12px;
        z-index: 10;
        color: #cbd5e1;
    }

    .data-table td {
        padding: 12px;
        border-bottom: 1px solid #1e293b;
    }

    .data-table tbody tr:hover {
        background: #1f2937;
    }

    .sticky-col-header {
        position: sticky !important;
        right: 0;
        z-index: 20 !important;
        background: #1e293b;
        box-shadow: -2px 0 4px rgba(0, 0, 0, 0.3);
    }

    /* 单元格进度条专属样式 */
    .status-cell {
        width: 120px;
    }

    .inline-progress {
        display: flex;
        flex-direction: column;
        justify-content: center;
        gap: 4px;
        width: 100px;
    }

    .inline-meta {
        font-size: 0.75rem;
        color: #38bdf8;
        font-weight: bold;
    }

    .inline-bar-bg {
        height: 4px;
        background: #0f172a;
        border-radius: 2px;
        overflow: hidden;
        width: 100%;
    }

    .inline-bar-fill {
        height: 100%;
        background: #38bdf8;
        transition: width 0.1s ease-out;
    }

    .status-dot.error {
        background: #ef4444;
    }

    .text-error {
        color: #ef4444;
    }

    .sticky-col-body {
        position: sticky;
        right: 0;
        z-index: 2;
        background: #111827;
        box-shadow: -2px 0 4px rgba(0, 0, 0, 0.3);
    }

    .data-table tbody tr:hover .sticky-col-body {
        background: #1f2937;
    }

    .filename-cell {
        overflow: hidden;
    }

    .filename-layout {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
    }

    .filename-text {
        color: #38bdf8;
        font-weight: 500;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        flex: 1; /* 撑满剩余空间 */
    }

    .copy-btn {
        opacity: 0; /* 默认隐藏，保持界面清爽 */
        cursor: pointer;
        font-size: 0.9rem;
        transition: opacity 0.2s, transform 0.1s;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    /* 只有当鼠标悬浮在这一格时，才显示复制按钮 */
    .filename-cell:hover .copy-btn {
        opacity: 1;
    }

    .copy-btn:hover {
        transform: scale(1.5); /* 鼠标指上去稍微放大 */
    }

    .file-id {
        font-size: 0.7rem;
        color: #64748b;
        margin-top: 4px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .file-id {
        font-size: 0.7rem;
        color: #64748b;
        margin-top: 4px;
    }

    .pagination {
        display: flex;
        justify-content: space-between;
        padding: 12px;
        background: #111827;
        border: 1px solid #1e293b;
        border-top: none;
        border-radius: 0 0 8px 8px;
    }

    .page-controls button {
        background: #1e293b;
        border: 1px solid #334155;
        color: white;
        padding: 4px 10px;
        border-radius: 4px;
        cursor: pointer;
        margin-left: 8px;
    }

    .page-controls button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .batch-action-bar {
        display: flex;
        justify-content: space-between;
        padding: 10px;
        background: #1e3a8a;
        border-radius: 6px;
        margin-bottom: 10px;
    }

    .batch-btn {
        background: #3b82f6;
        border: none;
        color: white;
        padding: 6px 12px;
        border-radius: 4px;
        cursor: pointer;
    }

    .batch-btn.danger {
        background: #ef4444;
    }

    .progress-panel {
        background: #1e293b;
        border: 1px solid #334155;
        padding: 12px;
        border-radius: 8px;
        margin-bottom: 12px;
    }

    .prog-bar-bg {
        height: 6px;
        background: #0f172a;
        border-radius: 3px;
        overflow: hidden;
        margin-top: 4px;
    }

    .prog-bar-fill {
        height: 100%;
        background: #10b981;
        transition: 0.2s;
    }

    .status-dot {
        display: inline-block;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        margin-right: 6px;
    }

    .status-dot.done {
        background: #10b981;
    }

    .status-dot.pending {
        background: #64748b;
    }
</style>