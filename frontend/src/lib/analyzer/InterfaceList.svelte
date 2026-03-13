<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { GetInterfaces } from '../../../wailsjs/go/main/App';

    const dispatch = createEventDispatcher();

    let interfaces: any[] = [];
    let isLoading = false;
    let searchName = "";

    // Derived store for filtering
    $: filteredInterfaces = interfaces.filter(iface =>
        searchName === "" ||
        iface.name.toLowerCase().includes(searchName.toLowerCase()) ||
        (iface.description && iface.description.toLowerCase().includes(searchName.toLowerCase()))
    );

    onMount(async () => {
        await refreshInterfaces();
    });

    // Fetch interfaces from the Gin API
    async function refreshInterfaces() {
        isLoading = true;
        try {
            const resStr = await GetInterfaces();
            const resData = JSON.parse(resStr);

            // 提取网卡列表数据
            interfaces = resData.list || [];

        } catch (err) {
            console.error("获取网卡列表失败:", err);
            alert("获取网卡列表失败: " + err);
        } finally {
            isLoading = false;
        }
    }

    function handleSearch() {
        // The reactive $: filteredInterfaces handles this automatically,
        // but we keep the function for the 'Enter' key and button consistency.
    }

    function handleReset() {
        searchName = "";
    }

    function handleStartCapture(iface: any) {
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

    .filter-group { display: flex; align-items: center; gap: 8px; }
    .filter-group label { color: #94a3b8; font-size: 0.85rem; }
    .filter-group input { background: #1e293b; border: 1px solid #334155; color: white; padding: 6px 10px; border-radius: 6px; outline: none; font-size: 0.85rem; }

    .import-action { margin-left: auto; }
    .primary-btn { background: #4f46e5; color: white; border: none; padding: 8px 14px; border-radius: 6px; cursor: pointer; transition: 0.2s; }
    .primary-btn:hover:not(:disabled) { background: #4338ca; }
    .primary-btn:disabled { opacity: 0.6; cursor: not-allowed; }

    .action-btn { background: #1e293b; border: 1px solid #334155; color: white; padding: 6px 12px; border-radius: 4px; cursor: pointer; }
    .action-btn:hover { background: #3b82f6; }
    .action-btn.outline { background: transparent; border: 1px solid #475569; color: #cbd5e1; }
    .action-btn.outline:hover { background: #334155; color: white; }

    .table-wrapper { flex: 1; overflow: auto; border: 1px solid #1e293b; border-radius: 8px; background: #111827; }
    .data-table { width: 100%; min-width: 800px; border-collapse: collapse; text-align: left; font-size: 0.85rem; table-layout: fixed; }
    .data-table th { position: sticky; top: 0; background: #1e293b; padding: 12px; z-index: 10; color: #cbd5e1; }
    .data-table td { padding: 12px; border-bottom: 1px solid #1e293b; }
    .data-table tbody tr:hover { background: #1f2937; }

    .sticky-col-header { position: sticky !important; right: 0; z-index: 20 !important; background: #1e293b; box-shadow: -2px 0 4px rgba(0, 0, 0, 0.3); }
    .sticky-col-body { position: sticky; right: 0; z-index: 2; background: #111827; box-shadow: -2px 0 4px rgba(0, 0, 0, 0.3); }
    .data-table tbody tr:hover .sticky-col-body { background: #1f2937; }

    .filename-text { color: #38bdf8; font-weight: 500; font-family: 'Fira Code', monospace;}

    .status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }
    .status-dot.done { background: #10b981; }

    /* Address specific styles */
    .address-list { display: flex; flex-direction: column; gap: 4px; }
    .address-item { background: #1e293b; display: inline-block; width: fit-content; padding: 2px 8px; border-radius: 4px; border: 1px solid #334155;}
    .addr-ip { color: #e2e8f0; font-family: monospace;}
    .addr-mask { color: #64748b; font-size: 0.75rem; font-family: monospace;}
</style>