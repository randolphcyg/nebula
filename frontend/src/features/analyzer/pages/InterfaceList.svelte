<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { GetInterfaces } from '../../../../wailsjs/go/main/App';
    import { error as showError } from '../../../stores/toast';
    import { logger } from '../../../utils/logger';

    const dispatch = createEventDispatcher();

    interface NetworkInterface {
        name: string;
        description: string;
        addresses: Array<{
            addr: string;
            netmask: string;
        }>;
    }

    let interfaces: NetworkInterface[] = [];
    let isLoading = false;
    let searchName = '';

    $: filteredInterfaces = interfaces.filter(iface =>
        searchName === '' ||
        iface.name.toLowerCase().includes(searchName.toLowerCase()) ||
        (iface.description && iface.description.toLowerCase().includes(searchName.toLowerCase()))
    );

    onMount(async () => {
        await refreshInterfaces();
    });

    async function refreshInterfaces() {
        isLoading = true;
        try {
            const resStr = await GetInterfaces();
            const resData = JSON.parse(resStr);
            interfaces = resData.list || [];
            logger.debug('网卡列表加载成功，共', interfaces.length, '个接口');
        } catch (err) {
            logger.error('获取网卡列表失败:', err);
            showError('获取网卡列表失败：' + err);
        } finally {
            isLoading = false;
        }
    }

    function handleSearch() {
        // The reactive $: filteredInterfaces handles this automatically,
        // but we keep the function for the 'Enter' key and button consistency.
    }

    function handleReset() {
        searchName = '';
    }

    function handleStartCapture(iface: NetworkInterface) {
        // Dispatch event to parent (AnalyzerHub) to switch to capture view
        dispatch('capture', iface);
    }
</script>

<div class="list-container">
    <div class="filter-panel">
        <div class="filter-group">
            <label for="sName">Interface Name:</label>
            <input id="sName" type="text"
                   bind:value={searchName}
                   on:keyup={(e) => e.key === 'Enter' && handleSearch()}
                   placeholder="e.g. eth0, en0" />
        </div>

        <div class="filter-actions">
            <button class="action-btn" on:click={handleSearch}>Search</button>
            <button class="action-btn outline" on:click={handleReset}>Reset</button>
        </div>

        <div class="import-action">
            <button class="primary-btn" on:click={refreshInterfaces} disabled={isLoading}>
                {isLoading ? '🔄 Refreshing...' : '🔄 Refresh Interfaces'}
            </button>
        </div>
    </div>

    <div class="table-wrapper">
        <table class="data-table">
            <thead>
            <tr>
                <th style="width: 150px;">Name</th>
                <th style="width: 200px;">Description</th>
                <th>Addresses (IP / Netmask)</th>
                <th style="width: 80px;">Status</th>
                <th class="sticky-col-header" style="width: 120px;">Action</th>
            </tr>
            </thead>
            <tbody>
            {#each filteredInterfaces as iface}
                <tr>
                    <td class="filename-cell">
                        <div class="filename-text">{iface.name}</div>
                    </td>
                    <td>
                        <span style="color: #94a3b8; font-size: 0.8rem;">
                            {iface.description || "No description"}
                        </span>
                    </td>
                    <td>
                        {#if iface.addresses && iface.addresses.length > 0}
                            <div class="address-list">
                                {#each iface.addresses as addr}
                                    <div class="address-item">
                                        <span class="addr-ip">{addr.addr}</span>
                                        {#if addr.netmask}
                                            <span class="addr-mask">/ {addr.netmask}</span>
                                        {/if}
                                    </div>
                                {/each}
                            </div>
                        {:else}
                            <span style="color: #64748b;">-</span>
                        {/if}
                    </td>
                    <td class="status-cell">
                        <span class="status-dot done"></span>
                        <span style="font-size: 0.8rem; color: #10b981;">Up</span>
                    </td>
                    <td class="sticky-col-body">
                        <button class="action-btn" on:click={() => handleStartCapture(iface)}>
                            ⚡ Live Capture
                        </button>
                    </td>
                </tr>
            {/each}
            {#if filteredInterfaces.length === 0 && !isLoading}
                <tr>
                    <td colspan="5" style="text-align: center; padding: 2rem; color: #64748b;">
                        No network interfaces found
                    </td>
                </tr>
            {/if}
            </tbody>
        </table>
    </div>
</div>

<style>
    /* Styles are 100% matched with PcapList.svelte for visual consistency */
    .list-container {
        height: 100%;
        display: flex;
        flex-direction: column;
        position: relative;
        background-color: var(--bg-primary);
    }

    .filter-panel {
        display: flex;
        flex-wrap: wrap;
        gap: 16px;
        align-items: center;
        padding: 12px;
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        margin-bottom: 12px;
    }

    .filter-group { display: flex; align-items: center; gap: 8px; }
    .filter-group label { color: var(--text-secondary); font-size: 0.85rem; }
    .filter-group input { background: var(--bg-tertiary); border: 1px solid var(--border-color); color: var(--text-primary); padding: 6px 10px; border-radius: 6px; outline: none; font-size: 0.85rem; }

    .import-action { margin-left: auto; }
    .primary-btn { background: var(--color-primary); color: white; border: none; padding: 8px 14px; border-radius: 6px; cursor: pointer; transition: 0.2s; }
    .primary-btn:hover:not(:disabled) { background: #4338ca; }
    .primary-btn:disabled { opacity: 0.6; cursor: not-allowed; }

    .action-btn { background: var(--bg-tertiary); border: 1px solid var(--border-color); color: var(--text-primary); padding: 6px 12px; border-radius: 4px; cursor: pointer; }
    .action-btn:hover { background: var(--color-primary); border-color: var(--color-primary); color: white; }
    .action-btn.outline { background: transparent; border: 1px solid var(--border-color); color: var(--text-primary); }
    .action-btn.outline:hover { background: var(--bg-tertiary); color: var(--text-primary); }

    .table-wrapper { flex: 1; overflow: auto; border: 1px solid var(--border-color); border-radius: 8px; background: var(--bg-secondary); }
    .table-wrapper::-webkit-scrollbar {
        width: 8px;
        height: 8px;
    }
    .table-wrapper::-webkit-scrollbar-track {
        background: var(--bg-primary);
    }
    .table-wrapper::-webkit-scrollbar-thumb {
        background: var(--border-color);
        border-radius: 4px;
    }
    .table-wrapper::-webkit-scrollbar-thumb:hover {
        background: var(--border-color-light);
    }
    .data-table { width: 100%; min-width: 800px; border-collapse: collapse; text-align: left; font-size: 0.85rem; table-layout: fixed; }
    .data-table th { position: sticky; top: 0; background: var(--bg-tertiary); padding: 12px; z-index: 10; color: var(--text-primary); }
    .data-table td { padding: 12px; border-bottom: 1px solid var(--border-color); color: var(--text-secondary); }
    .data-table tbody tr:hover { background: var(--bg-tertiary); }

    .sticky-col-header { position: sticky !important; right: 0; z-index: 20 !important; background: var(--bg-tertiary); box-shadow: -2px 0 4px rgba(0, 0, 0, 0.1); }
    .sticky-col-body { position: sticky; right: 0; z-index: 2; background: var(--bg-secondary); box-shadow: -2px 0 4px rgba(0, 0, 0, 0.1); }
    .data-table tbody tr:hover .sticky-col-body { background: var(--bg-tertiary); }

    .filename-text { color: var(--color-info); font-weight: 500; font-family: 'Fira Code', monospace;}

    .status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }
    .status-dot.done { background: #10b981; }

    /* Address specific styles */
    .address-list { display: flex; flex-direction: column; gap: 4px; }
    .address-item { background: var(--bg-tertiary); display: inline-block; width: fit-content; padding: 2px 8px; border-radius: 4px; border: 1px solid var(--border-color);}
    .addr-ip { color: var(--text-primary); font-family: monospace;}
    .addr-mask { color: var(--text-secondary); font-size: 0.75rem; font-family: monospace;}
</style>