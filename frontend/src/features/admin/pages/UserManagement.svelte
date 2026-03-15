<script lang="ts">
    import { auth } from '$lib/stores/auth';
    import { success as showSuccess, error as showError } from '$lib/stores/toast';
    import { logger } from '$lib/utils/logger';
    import { UserStatus, UserStatusMap, UserStatusColorMap } from '$lib/types';
    import { 
        GetUsers, 
        SearchUsers, 
        UpdateUserStatus, 
        BatchUpdateUserStatus,
        GetAuditLogs,
        GetAllRoles,
        UpdateUserRole,
        DeleteUser,
        UpdateUserProfile 
    } from '../../../../wailsjs/go/main/App';
    import { onMount } from 'svelte';

    interface User {
        id: number;
        username: string;
        email: string;
        role: string;
        roleCode: string;
        roleID: number;
        status: UserStatus;
        lastLogin: string;
        createdAt: string;
    }

    interface Role {
        id: number;
        name: string;
        code: string;
        description: string;
        permissions: string[];
    }

    interface AuditLog {
        id: number;
        userID: number;
        username: string;
        operator: string;
        action: string;
        oldStatus: number;
        newStatus: number;
        createdAt: string;
    }

    let users: User[] = [];
    let allUsers: User[] = [];
    let isLoading = false;
    let searchKeyword = '';
    let selectedUserIds: number[] = [];
    
    // 响应式计算
    $: isAllSelected = users.length > 0 && selectedUserIds.length === users.length;
    $: hasSelectedUsers = selectedUserIds.length > 0;
    
    let showAuditLogs = false;
    let auditLogs: AuditLog[] = [];
    let showRoleModal = false;
    let selectedUser: User | null = null;
    let roles: Role[] = [];
    let filter = 'all'; // all, pending, active, disabled
    let currentUserId = 0;
    
    // 编辑邮箱状态
    let editingUserId: number | null = null;
    let editEmail = '';

    // 处理搜索输入
    function handleSearchInput(e: Event) {
        const target = e.target as HTMLInputElement;
        searchKeyword = target.value;
    }

    // 加载角色列表
    async function loadRoles() {
        try {
            const result = await GetAllRoles();
            roles = result as Role[];
        } catch (err: any) {
            logger.error('加载角色失败:', err);
            showError('加载角色列表失败');
        }
    }

    // 加载用户列表
    async function loadUsers() {
        isLoading = true;
        try {
            const result = await GetUsers();
            allUsers = result as User[];
            // 获取当前用户 ID
            const user = auth.getCurrentUser();
            if (user) {
                currentUserId = user.id;
            }
            applyFilters();
        } catch (err: any) {
            logger.error('加载用户失败:', err);
            showError('加载用户列表失败');
        } finally {
            isLoading = false;
        }
    }

    // 搜索用户
    async function handleSearch() {
        isLoading = true;
        try {
            if (searchKeyword.trim() === '') {
                await loadUsers();
            } else {
                const result = await SearchUsers(searchKeyword);
                allUsers = result as User[];
                applyFilters();
            }
        } catch (err: any) {
            console.error('搜索失败:', err);
            showError('搜索失败');
        } finally {
            isLoading = false;
        }
    }

    // 应用过滤器
    function applyFilters() {
        switch (filter) {
            case 'pending':
                users = allUsers.filter(u => u.status === UserStatus.Pending);
                break;
            case 'active':
                users = allUsers.filter(u => u.status === UserStatus.Active);
                break;
            case 'disabled':
                users = allUsers.filter(u => u.status === UserStatus.Disabled);
                break;
            default:
                users = allUsers;
        }
    }

    // 审核用户
    async function approveUser(userId: number, status: UserStatus) {
        try {
            const token = auth.getToken();
            await UpdateUserStatus(userId, status, token);
            const statusText = getStatusAction(status);
            showSuccess(`用户${statusText}，审核成功！`);
            await loadUsers();
        } catch (err: any) {
            logger.error('审核失败:', err);
            showError(`审核失败：${err.message}`);
        }
    }

    // 禁用/启用用户
    async function toggleUserStatus(userId: number, newStatus: UserStatus) {
        try {
            const token = auth.getToken();
            await UpdateUserStatus(userId, newStatus, token);
            const statusText = newStatus === UserStatus.Disabled ? '禁用' : '启用';
            showSuccess(`用户${statusText}成功！`);
            await loadUsers();
        } catch (err: any) {
            logger.error('操作失败:', err);
            showError(`操作失败：${err.message}`);
        }
    }

    // 删除用户
    async function deleteUser(userId: number) {
        if (!confirm('确定要删除该用户吗？此操作不可恢复！')) {
            return;
        }

        try {
            const token = auth.getToken();
            await DeleteUser(userId, token);
            showSuccess('用户删除成功！');
            await loadUsers();
        } catch (err: any) {
            logger.error('删除失败:', err);
            showError(`删除失败：${err.message}`);
        }
    }

    // 批量审核
    async function batchApprove(status: number) {
        if (selectedUserIds.length === 0) {
            showError('请先选择用户');
            return;
        }

        if (!confirm(`确定要${status === 1 ? '通过' : '拒绝'}选中的 ${selectedUserIds.length} 个用户吗？`)) {
            return;
        }

        try {
            const token = auth.getToken();
            await BatchUpdateUserStatus(selectedUserIds, status, token);
            const statusText = status === 1 ? '通过' : '拒绝';
            showSuccess(`批量${statusText}成功！`);
            selectedUserIds = [];
            await loadUsers();
        } catch (err: any) {
            logger.error('批量审核失败:', err);
            showError(`批量审核失败：${err.message}`);
        }
    }

    // 批量删除
    async function batchDelete() {
        if (selectedUserIds.length === 0) {
            showError('请先选择用户');
            return;
        }

        try {
            const token = auth.getToken();
            
            // 逐个删除
            for (const userId of selectedUserIds) {
                await DeleteUser(userId, token);
            }
            
            showSuccess(`批量删除 ${selectedUserIds.length} 个用户成功！`);
            selectedUserIds = [];
            await loadUsers();
        } catch (err: any) {
            logger.error('批量删除失败:', err);
            showError(`批量删除失败：${err.message}`);
        }
    }

    // 批量禁用
    async function batchDisable() {
        if (selectedUserIds.length === 0) {
            showError('请先选择用户');
            return;
        }

        try {
            const token = auth.getToken();
            await BatchUpdateUserStatus(selectedUserIds, UserStatus.Disabled, token);
            showSuccess(`批量禁用 ${selectedUserIds.length} 个用户成功！`);
            selectedUserIds = [];
            await loadUsers();
        } catch (err: any) {
            logger.error('批量禁用失败:', err);
            showError(`批量禁用失败：${err.message}`);
        }
    }

    // 批量启用
    async function batchEnable() {
        if (selectedUserIds.length === 0) {
            showError('请先选择用户');
            return;
        }

        try {
            const token = auth.getToken();
            await BatchUpdateUserStatus(selectedUserIds, UserStatus.Active, token);
            showSuccess(`批量启用 ${selectedUserIds.length} 个用户成功！`);
            selectedUserIds = [];
            await loadUsers();
        } catch (err: any) {
            logger.error('批量启用失败:', err);
            showError(`批量启用失败：${err.message}`);
        }
    }

    // 切换选择
    function toggleSelection(userId: number) {
        const index = selectedUserIds.indexOf(userId);
        if (index > -1) {
            selectedUserIds.splice(index, 1);
        } else {
            selectedUserIds.push(userId);
        }
        // 强制触发响应式更新
        selectedUserIds = selectedUserIds;
    }

    // 全选/取消全选
    function toggleSelectAll() {
        selectedUserIds = isAllSelected ? [] : users.map(u => u.id);
    }

    // 加载审核日志
    async function loadAuditLogs() {
        try {
            const result = await GetAuditLogs(50);
            auditLogs = result as AuditLog[];
            showAuditLogs = true;
        } catch (err: any) {
            logger.error('加载日志失败:', err);
            showError('加载审核日志失败');
        }
    }

    // 处理角色变更
    async function handleRoleChange(e: Event, userId: number) {
        try {
            const target = e.target as HTMLSelectElement;
            const roleId = parseInt(target.value);
            await UpdateUserRole(userId, roleId);
            showSuccess('角色更新成功！');
            await loadUsers();
        } catch (err: any) {
            logger.error('更新角色失败:', err);
            showError(`更新角色失败：${err.message}`);
            await loadUsers(); // 恢复原角色
        }
    }

    // 开始编辑邮箱
    function startEditEmail(user: User) {
        editingUserId = user.id;
        editEmail = user.email;
    }

    // 取消编辑邮箱
    function cancelEditEmail() {
        editingUserId = null;
        editEmail = '';
    }

    // 保存邮箱修改
    async function saveEmail(userId: number) {
        if (!editEmail) {
            showError('邮箱不能为空');
            return;
        }

        try {
            const token = localStorage.getItem('token');
            if (!token) {
                showError('未登录，请先登录');
                return;
            }

            await UpdateUserProfile(userId, editEmail, token);
            showSuccess('邮箱修改成功！');
            
            // 更新本地用户列表
            const user = allUsers.find(u => u.id === userId);
            if (user) {
                user.email = editEmail;
            }
            
            editingUserId = null;
            editEmail = '';
        } catch (err: any) {
            showError(`修改邮箱失败：${err.message}`);
        }
    }

    // 状态文本
    function getStatusText(status: UserStatus): string {
        switch (status) {
            case UserStatus.Pending: return '⏳ 待审核';
            case UserStatus.Active: return '✅ 正常';
            case UserStatus.Disabled: return '🚫 禁用';
            default: return '未知';
        }
    }

    // 状态样式
    function getStatusClass(status: UserStatus): string {
        switch (status) {
            case UserStatus.Pending: return 'status-pending';
            case UserStatus.Active: return 'status-active';
            case UserStatus.Disabled: return 'status-disabled';
        }
    }

    // 状态操作文本
    function getStatusAction(status: UserStatus): string {
        switch (status) {
            case UserStatus.Active: return '通过';
            case UserStatus.Disabled: return '拒绝';
            default: return '';
        }
    }

    // 审核日志文本
    function getActionText(action: string): string {
        const map: Record<string, string> = {
            'approve': '✅ 审核通过',
            'reject': '🚫 审核拒绝',
            'disable': '🚫 禁用',
            'enable': '✅ 启用',
            'delete': '�️ 删除',
            'batch_approve': '✅ 批量通过',
            'batch_reject': '🚫 批量拒绝',
        };
        return map[action] || action;
    }

    onMount(() => {
        loadUsers();
        loadRoles();
    });
</script>

<div class="user-management">
    <!-- 工具栏 -->
    <div class="toolbar">
        <div class="toolbar-left">
            <div class="search-box">
                <input 
                    type="text" 
                    placeholder="搜索用户名或邮箱..."
                    value={searchKeyword}
                    on:keydown={(e) => e.key === 'Enter' && handleSearch()}
                    on:input={handleSearchInput}
                />
                <button class="btn-search" on:click={handleSearch}>搜索</button>
                <button class="btn-clear" on:click={() => { searchKeyword = ''; loadUsers(); }}>清空</button>
            </div>

            <div class="filter-buttons">
                <button 
                    class:active={filter === 'all'} 
                    on:click={() => { filter = 'all'; applyFilters(); }}
                >
                    全部 ({allUsers.length})
                </button>
                <button 
                    class:active={filter === 'pending'} 
                    on:click={() => { filter = 'pending'; applyFilters(); }}
                >
                    待审核 ({allUsers.filter(u => u.status === 0).length})
                </button>
                <button 
                    class:active={filter === 'active'} 
                    on:click={() => { filter = 'active'; applyFilters(); }}
                >
                    正常 ({allUsers.filter(u => u.status === 1).length})
                </button>
                <button 
                    class:active={filter === 'disabled'} 
                    on:click={() => { filter = 'disabled'; applyFilters(); }}
                >
                    禁用 ({allUsers.filter(u => u.status === 2).length})
                </button>
            </div>
        </div>
        
        <div class="toolbar-right">
            <button class="btn-log" on:click={loadAuditLogs}>
                📋 审核日志
            </button>
        </div>
    </div>

    <!-- 批量操作栏 -->
    {#if hasSelectedUsers}
        {@const selectedUsersList = allUsers.filter(u => selectedUserIds.includes(u.id))}
        {@const hasPending = selectedUsersList.some(u => u.status === UserStatus.Pending)}
        {@const hasActive = selectedUsersList.some(u => u.status === UserStatus.Active)}
        {@const hasDisabled = selectedUsersList.some(u => u.status === UserStatus.Disabled)}
        {@const hasSuperAdmin = selectedUsersList.some(u => u.roleCode === 'admin')}
        
        <div class="batch-bar">
            <span class="selected-count">已选择 {selectedUserIds.length} 个用户</span>
            <div class="batch-actions">
                {#if hasSuperAdmin}
                    <!-- 选中了超级管理员：显示警告，不显示任何操作按钮 -->
                    <span class="warning-text" style="color: #f59e0b;">⚠️ 包含超级管理员，无法执行批量操作</span>
                {:else}
                    <!-- 不包含超级管理员：根据状态显示操作按钮 -->
                    {#if hasPending && !hasActive && !hasDisabled}
                        <!-- 只选中了待审核用户 -->
                        <button class="btn-batch-approve" on:click={() => batchApprove(1)}>
                            ✅ 批量通过
                        </button>
                        <button class="btn-batch-reject" on:click={() => batchApprove(2)}>
                            🚫 批量拒绝
                        </button>
                    {:else if !hasPending && hasActive && !hasDisabled}
                        <!-- 只选中了正常用户 -->
                        <button class="btn-batch-disable" on:click={() => batchDisable()}>
                            ⛔ 批量禁用
                        </button>
                        <button class="btn-batch-delete" on:click={() => batchDelete()}>
                            🗑️ 批量删除
                        </button>
                    {:else if !hasPending && !hasActive && hasDisabled}
                        <!-- 只选中了禁用用户 -->
                        <button class="btn-batch-enable" on:click={() => batchEnable()}>
                            ✅ 批量启用
                        </button>
                        <button class="btn-batch-delete" on:click={() => batchDelete()}>
                            🗑️ 批量删除
                        </button>
                    {:else}
                        <!-- 混合状态：只显示删除按钮 -->
                        <button class="btn-batch-delete" on:click={() => batchDelete()}>
                            🗑️ 批量删除
                        </button>
                    {/if}
                {/if}
            </div>
        </div>
    {/if}

    <!-- 用户列表 -->
    <div class="user-list-container">
        {#if isLoading}
            <div class="loading">加载中...</div>
        {:else}
            <div class="table-wrapper">
                <table class="data-table">
                    <thead>
                        <tr>
                            <th>
                                <input 
                                    type="checkbox" 
                                    checked={isAllSelected}
                                    on:change={toggleSelectAll}
                                />
                            </th>
                            <th>用户名</th>
                            <th>邮箱</th>
                            <th>角色</th>
                            <th>状态</th>
                            <th>注册时间</th>
                            <th>最后登录</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each users as user}
                            <tr 
                                class:selected={selectedUserIds.includes(user.id)}
                                on:click={() => toggleSelection(user.id)}
                            >
                                <td>
                                    <input 
                                        type="checkbox"
                                        checked={selectedUserIds.includes(user.id)}
                                        on:change={() => toggleSelection(user.id)}
                                    />
                                </td>
                                <td>
                                    <div class="user-name">
                                        <span class="avatar">{user.username.charAt(0).toUpperCase()}</span>
                                        {user.username}
                                    </div>
                                </td>
                                <td>
                                    {#if editingUserId === user.id}
                                        <div class="email-edit">
                                            <input 
                                                type="email" 
                                                bind:value={editEmail}
                                                placeholder="请输入邮箱"
                                            />
                                            <div class="edit-actions">
                                                <button 
                                                    class="btn-save" 
                                                    on:click={() => saveEmail(user.id)}
                                                >
                                                    保存
                                                </button>
                                                <button 
                                                    class="btn-cancel" 
                                                    on:click={cancelEditEmail}
                                                >
                                                    取消
                                                </button>
                                            </div>
                                        </div>
                                    {:else}
                                        <div class="email-display">
                                            {user.email}
                                            <button 
                                                class="btn-edit" 
                                                on:click={() => startEditEmail(user)}
                                                title="编辑邮箱"
                                            >
                                                ✏️
                                            </button>
                                        </div>
                                    {/if}
                                </td>
                                <td>
                                    {#if user.status === 1 && user.id !== currentUserId}
                                        <select 
                                            class="role-select"
                                            value={user.roleID || ''}
                                            on:change={(e) => handleRoleChange(e, user.id)}
                                        >
                                            <option value="">选择角色</option>
                                            {#each roles as role}
                                                <option value={role.id}>{role.name}</option>
                                            {/each}
                                        </select>
                                    {:else}
                                        <span class="role-badge">{user.role || '无'}</span>
                                    {/if}
                                </td>
                                <td>
                                    <span class="status-badge {getStatusClass(user.status)}">
                                        {getStatusText(user.status)}
                                    </span>
                                </td>
                                <td>{user.createdAt ? new Date(user.createdAt).toLocaleString('zh-CN') : '-'}</td>
                                <td>
                                    {user.lastLogin ? new Date(user.lastLogin).toLocaleString('zh-CN') : '从未登录'}
                                </td>
                                <td>
                                    <div class="action-buttons">
                                        {#if user.id !== currentUserId}
                                            {#if user.status === 0}
                                                <button 
                                                    class="btn-approve" 
                                                    on:click={() => approveUser(user.id, 1)}
                                                    title="通过审核"
                                                >
                                                    ✅ 通过
                                                </button>
                                                <button 
                                                    class="btn-reject" 
                                                    on:click={() => approveUser(user.id, 2)}
                                                    title="拒绝并禁用"
                                                >
                                                    🚫 拒绝
                                                </button>
                                            {:else if user.status === 1}
                                                <button 
                                                    class="btn-disable" 
                                                    on:click={() => toggleUserStatus(user.id, 2)}
                                                    title="禁用用户"
                                                >
                                                    🚫 禁用
                                                </button>
                                            {:else}
                                                <button 
                                                    class="btn-enable" 
                                                    on:click={() => toggleUserStatus(user.id, 1)}
                                                    title="启用用户"
                                                >
                                                    ✅ 启用
                                                </button>
                                            {/if}
                                            <button 
                                                class="btn-delete" 
                                                on:click={() => deleteUser(user.id)}
                                                title="删除用户"
                                            >
                                                🗑️ 删除
                                            </button>
                                        {/if}
                                    </div>
                                </td>
                            </tr>
                        {/each}
                        
                        {#if users.length === 0}
                            <tr>
                                <td colspan="9" class="empty-state">
                                    <div class="empty-icon">📭</div>
                                    <div class="empty-text">暂无用户数据</div>
                                </td>
                            </tr>
                        {/if}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>
</div>

{#if showAuditLogs}
    <div class="modal-overlay" on:click={() => showAuditLogs = false}>
        <div class="modal-content" on:click|stopPropagation>
            <div class="modal-header">
                <h2>📋 审核日志</h2>
                <button class="btn-close" on:click={() => showAuditLogs = false}>×</button>
            </div>
            <div class="modal-body">
                <table class="log-table">
                    <thead>
                        <tr>
                            <th>时间</th>
                            <th>操作用户</th>
                            <th>操作员</th>
                            <th>操作</th>
                            <th>变更前</th>
                            <th>变更后</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each auditLogs as log}
                            <tr>
                                <td>{new Date(log.createdAt).toLocaleString('zh-CN')}</td>
                                <td>{log.username}</td>
                                <td>{log.operator}</td>
                                <td>{getActionText(log.action)}</td>
                                <td>{getStatusText(log.oldStatus)}</td>
                                <td>{log.newStatus === -1 ? '已删除' : getStatusText(log.newStatus)}</td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
{/if}

<style>
    .user-management {
        padding: 1.5rem;
        max-width: 1400px;
        margin: 0 auto;
    }

    .page-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 2rem;
    }

    .page-header h1 {
        font-size: 1.75rem;
        font-weight: 700;
        color: var(--text-primary);
        margin-bottom: 0.25rem;
    }

    .page-header p {
        font-size: 0.875rem;
        color: var(--text-secondary);
    }

    .btn-log {
        padding: 0.5rem 1rem;
        background: var(--bg-tertiary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        color: var(--text-primary);
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
    }

    .btn-log:hover {
        background: var(--color-primary);
        color: white;
    }

    /* 工具栏 */
    .toolbar {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 1rem;
        margin-bottom: 1rem;
    }

    .toolbar-left {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        flex: 1;
    }

    .toolbar-right {
        display: flex;
        align-items: center;
    }

    .search-box {
        display: flex;
        gap: 0.5rem;
        max-width: 500px;
    }

    .search-box input {
        flex: 1;
        padding: 0.5rem 1rem;
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        color: var(--text-primary);
        font-size: 0.875rem;
    }

    .btn-search,
    .btn-clear {
        padding: 0.5rem 1rem;
        background: var(--color-primary);
        border: none;
        border-radius: var(--radius-md);
        color: white;
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
    }

    .btn-clear {
        background: var(--bg-tertiary);
        color: var(--text-primary);
    }

    .filter-buttons {
        display: flex;
        gap: 0.5rem;
    }

    .filter-buttons button {
        padding: 0.5rem 1rem;
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        color: var(--text-primary);
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
    }

    .filter-buttons button.active {
        background: var(--color-primary);
        color: white;
        border-color: var(--color-primary);
    }

    .batch-actions {
        display: flex;
        align-items: center;
        gap: 1rem;
        padding: 1rem;
        background: var(--bg-tertiary);
        border-radius: var(--radius-md);
        border: 1px solid var(--border-color);
    }

    /* 批量操作栏 */
    .batch-bar {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1rem;
        background: var(--bg-tertiary);
        border-radius: var(--radius-md);
        border: 1px solid var(--border-color);
        margin-bottom: 1rem;
    }

    .batch-bar .selected-count {
        font-size: 0.875rem;
        color: var(--text-secondary);
        font-weight: 500;
    }

    .batch-bar .batch-actions {
        display: flex;
        gap: 0.75rem;
        padding: 0;
        background: transparent;
        border: none;
    }

    .selected-count {
        font-size: 0.875rem;
        color: var(--text-secondary);
    }

    .btn-batch-approve,
    .btn-batch-reject {
        padding: 0.5rem 1rem;
        border: none;
        border-radius: var(--radius-md);
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
    }

    .btn-batch-approve {
        background: rgba(16, 185, 129, 0.1);
        color: #10b981;
    }

    .btn-batch-approve:hover {
        background: #10b981;
        color: white;
    }

    .btn-batch-reject {
        background: rgba(239, 68, 68, 0.1);
        color: #ef4444;
    }

    .btn-batch-reject:hover {
        background: #ef4444;
        color: white;
    }

    .btn-batch-delete {
        padding: 0.5rem 1rem;
        border: none;
        border-radius: var(--radius-md);
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
        background: rgba(107, 114, 128, 0.1);
        color: #6b7280;
    }

    .btn-batch-delete:hover {
        background: #6b7280;
        color: white;
    }

    .btn-batch-disable {
        padding: 0.5rem 1rem;
        border: none;
        border-radius: var(--radius-md);
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
        background: rgba(239, 68, 68, 0.1);
        color: #ef4444;
    }

    .btn-batch-disable:hover {
        background: #ef4444;
        color: white;
    }

    .btn-batch-enable {
        padding: 0.5rem 1rem;
        border: none;
        border-radius: var(--radius-md);
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
        background: rgba(16, 185, 129, 0.1);
        color: #10b981;
    }

    .btn-batch-enable:hover {
        background: #10b981;
        color: white;
    }

    /* 表格容器 */
    .user-list-container {
        background: var(--bg-secondary);
        border-radius: var(--radius-lg);
        border: 1px solid var(--border-color);
        overflow: hidden;
    }

    .table-wrapper {
        overflow-x: auto;
        max-width: 100%;
    }

    .loading {
        padding: 3rem;
        text-align: center;
        color: var(--text-secondary);
    }

    /* 用户名列 */
    .user-name {
        display: flex;
        align-items: center;
        gap: 0.75rem;
    }

    .avatar {
        width: 2rem;
        height: 2rem;
        border-radius: var(--radius-full);
        background: var(--color-primary);
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 600;
        font-size: 0.875rem;
    }

    /* 角色徽章 */
    .role-badge {
        display: inline-flex;
        padding: 0.25rem 0.75rem;
        border-radius: var(--radius-full);
        font-size: 0.875rem;
        font-weight: 500;
        background: rgba(79, 70, 229, 0.1);
        color: #4f46e5;
    }

    .role-select {
        padding: 0.25rem 1.5rem 0.25rem 0.5rem;
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        background: var(--bg-secondary);
        color: var(--text-primary);
        font-size: 0.875rem;
        cursor: pointer;
        min-width: 100px;
        appearance: none;
        background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%236b7280' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
        background-repeat: no-repeat;
        background-position: right 0.5rem center;
        transition: var(--transition-fast);
    }

    .role-select:hover {
        border-color: var(--color-primary);
    }

    .role-select:focus {
        outline: none;
        border-color: var(--color-primary);
        box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
    }

    .email-display {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .btn-edit {
        padding: 0.25rem 0.5rem;
        background: transparent;
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        font-size: 0.75rem;
        cursor: pointer;
        transition: var(--transition-fast);
    }

    .btn-edit:hover {
        background: var(--bg-tertiary);
        border-color: var(--color-primary);
    }

    .email-edit {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        min-width: 200px;
    }

    .email-edit input {
        padding: 0.5rem;
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        background: var(--bg-secondary);
        color: var(--text-primary);
        font-size: 0.875rem;
    }

    .edit-actions {
        display: flex;
        gap: 0.5rem;
    }

    .btn-save,
    .btn-cancel {
        padding: 0.375rem 0.75rem;
        border: none;
        border-radius: var(--radius-md);
        font-size: 0.75rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
    }

    .btn-save {
        background: var(--color-primary);
        color: white;
    }

    .btn-save:hover {
        background: var(--color-primary-dark);
    }

    .btn-cancel {
        background: var(--bg-tertiary);
        color: var(--text-primary);
    }

    .btn-cancel:hover {
        background: var(--border-color);
    }

    /* 状态徽章 */
    .status-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.25rem;
        padding: 0.25rem 0.75rem;
        border-radius: var(--radius-full);
        font-size: 0.875rem;
        font-weight: 500;
    }

    .status-pending {
        background: rgba(245, 158, 11, 0.1);
        color: #f59e0b;
    }

    .status-active {
        background: rgba(16, 185, 129, 0.1);
        color: #10b981;
    }

    .status-disabled {
        background: rgba(239, 68, 68, 0.1);
        color: #ef4444;
    }

    /* 操作按钮 */
    .action-buttons {
        display: flex;
        gap: 0.5rem;
    }

    .action-buttons button {
        padding: 0.375rem 0.75rem;
        border: none;
        border-radius: var(--radius-md);
        font-size: 0.875rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
    }

    .btn-approve {
        background: rgba(16, 185, 129, 0.1);
        color: #10b981;
    }

    .btn-approve:hover {
        background: #10b981;
        color: white;
    }

    .btn-reject,
    .btn-disable {
        background: rgba(239, 68, 68, 0.1);
        color: #ef4444;
    }

    .btn-reject:hover,
    .btn-disable:hover {
        background: #ef4444;
        color: white;
    }

    .btn-enable {
        background: rgba(16, 185, 129, 0.1);
        color: #10b981;
    }

    .btn-enable:hover {
        background: #10b981;
        color: white;
    }

    .btn-delete {
        background: rgba(107, 114, 128, 0.1);
        color: #6b7280;
    }

    .btn-delete:hover {
        background: #ef4444;
        color: white;
    }

    /* 模态框 */
    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.5);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }

    .modal-content {
        background: var(--bg-primary);
        border-radius: var(--radius-lg);
        border: 1px solid var(--border-color);
        max-width: 800px;
        width: 90%;
        max-height: 80vh;
        overflow: hidden;
        display: flex;
        flex-direction: column;
    }

    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 1.5rem;
        border-bottom: 1px solid var(--border-color);
    }

    .modal-header h2 {
        font-size: 1.25rem;
        font-weight: 600;
        color: var(--text-primary);
        margin: 0;
    }

    .btn-close {
        width: 2rem;
        height: 2rem;
        border: none;
        background: transparent;
        color: var(--text-secondary);
        font-size: 1.5rem;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: var(--radius-md);
        transition: var(--transition-fast);
    }

    .btn-close:hover {
        background: var(--bg-tertiary);
        color: var(--text-primary);
    }

    .modal-body {
        padding: 1.5rem;
        overflow-y: auto;
        flex: 1;
    }

    .log-table {
        width: 100%;
        border-collapse: collapse;
    }

    .log-table th,
    .log-table td {
        padding: 0.75rem;
        text-align: left;
        font-size: 0.875rem;
        border-bottom: 1px solid var(--border-color);
    }

    .log-table th {
        font-weight: 600;
        color: var(--text-primary);
        background: var(--bg-tertiary);
    }

    .log-table td {
        padding: 0.75rem;
        font-size: 0.875rem;
        color: var(--text-secondary);
        border-bottom: 1px solid var(--border-color);
    }

    .log-table tbody tr:last-child td {
        border-bottom: none;
    }
</style>
