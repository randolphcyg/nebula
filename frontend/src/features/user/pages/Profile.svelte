<script lang="ts">
    import { onMount } from 'svelte';
    import { auth } from '../../../stores/auth';
    import { themeStore } from '../../../stores/theme';
    import { preferencesStore } from '../../../stores/preferences';
    import { error as showError, success as showSuccess } from '../../../stores/toast';
    import { ChangePassword, UpdateUserProfile } from '../../../../wailsjs/go/main/App';
    import type { User } from '../../../types';

    let user: User | null = null;
    let isLoading = false;

    // Tab 切换
    let activeTab = 'profile';

    // 修改资料表单
    let editEmail = '';
    let isEditingProfile = false;

    // 修改密码表单
    let currentPassword = '';
    let newPassword = '';
    let confirmPassword = '';

    // 主题设置（从 store 初始化）
    let themeMode: 'light' | 'dark' = 'light';

    // 偏好设置（从 store 初始化）
    let pageSize = 10;
    let showNotifications = true;
    let compactMode = false;

    // 表单验证状态
    let emailError = '';
    let passwordError = '';

    onMount(() => {
        user = auth.getCurrentUser();
        if (user) {
            editEmail = user.email;
        }
        themeStore.subscribe(theme => {
            themeMode = theme.mode;
        });
        preferencesStore.subscribe(prefs => {
            pageSize = prefs.pageSize;
            showNotifications = prefs.showNotifications;
            compactMode = prefs.compactMode;
        });
    });

    async function handleChangeProfile() {
        if (!editEmail) {
            showError('邮箱不能为空');
            return;
        }

        if (!user) {
            showError('用户信息未加载');
            return;
        }

        isLoading = true;
        try {
            const token = localStorage.getItem('token');
            if (!token) {
                showError('未登录，请先登录');
                return;
            }

            await UpdateUserProfile(user.id, editEmail, token);
            showSuccess('邮箱修改成功');

            // 更新本地用户信息
            user.email = editEmail;
            auth.updateUser(user);

            isEditingProfile = false;
        } catch (err: any) {
            showError('修改邮箱失败：' + (err.message || err));
        } finally {
            isLoading = false;
        }
    }

    async function handleChangePassword() {
        if (!currentPassword || !newPassword || !confirmPassword) {
            showError('请填写所有密码字段');
            return;
        }

        if (newPassword.length < 6) {
            showError('密码长度至少为 6 位');
            return;
        }

        if (newPassword !== confirmPassword) {
            showError('两次输入的新密码不一致');
            return;
        }

        isLoading = true;
        try {
            const token = localStorage.getItem('token');
            if (!token) {
                showError('未登录，请先登录');
                return;
            }

            await ChangePassword(token, currentPassword, newPassword);
            showSuccess('密码修改成功，请重新登录');
            
            // 清空表单
            currentPassword = '';
            newPassword = '';
            confirmPassword = '';
            
            // 退出登录
            setTimeout(() => {
                auth.logout();
            }, 1500);
        } catch (err: any) {
            showError('修改密码失败：' + (err.message || err));
        } finally {
            isLoading = false;
        }
    }

    function handleThemeChange() {
        themeStore.toggleMode();
        showSuccess('主题已切换');
    }

    function handlePageSizeChange() {
        preferencesStore.setPageSize(pageSize);
        showSuccess('设置已保存');
    }

    function handleNotificationToggle() {
        preferencesStore.toggleNotifications();
        showSuccess(showNotifications ? '已开启通知' : '已关闭通知');
    }

    function handleCompactModeToggle() {
        preferencesStore.toggleCompactMode();
        showSuccess(compactMode ? '已开启紧凑模式' : '已关闭紧凑模式');
    }

    function switchTab(tab: string) {
        activeTab = tab;
    }
</script>

<div class="profile-container">

    <!-- Tab 导航 -->
    <div class="profile-tabs">
        <button 
            class="tab-btn" 
            class:active={activeTab === 'profile'}
            on:click={() => switchTab('profile')}
        >
            👤 基本信息
        </button>
        <button 
            class="tab-btn" 
            class:active={activeTab === 'settings'}
            on:click={() => switchTab('settings')}
        >
            ⚙️ 偏好设置
        </button>
    </div>

    <div class="profile-content">
        <!-- 基本信息 Tab -->
        {#if activeTab === 'profile'}
            <section class="profile-section">
                <div class="info-list">
                    <div class="info-row">
                        <span class="info-label">用户名</span>
                        <span class="info-value">{user?.username || '-'}</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">邮箱</span>
                        <span class="info-value">
                            {#if isEditingProfile}
                                <div class="edit-field">
                                    <input 
                                        type="email" 
                                        bind:value={editEmail}
                                        placeholder="请输入邮箱"
                                        class={emailError ? 'error' : ''}
                                    />
                                    <div class="field-actions">
                                        <button class="btn-sm btn-primary" on:click={handleChangeProfile} disabled={isLoading}>
                                            {#if isLoading}保存中...{:else}保存{/if}
                                        </button>
                                        <button class="btn-sm btn-outline" on:click={() => { 
                                            isEditingProfile = false; 
                                            emailError = '';
                                            if (user) editEmail = user.email; 
                                        }}>
                                            取消
                                        </button>
                                    </div>
                                </div>
                            {:else}
                                {user?.email || '-'}
                                <button class="btn-icon" on:click={() => { isEditingProfile = true; emailError = ''; }}>
                                    ✏️
                                </button>
                            {/if}
                        </span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">角色</span>
                        <span class="info-value"><span class="badge">{user?.role || '-'}</span></span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">上次登录</span>
                        <span class="info-value">{user?.lastLogin ? new Date(user.lastLogin).toLocaleString('zh-CN') : '从未登录'}</span>
                    </div>
                </div>

                <div class="security-section">
                    <div class="password-form">                        
                        <div class="password-inputs">
                            <input
                                id="currentPassword"
                                type="password"
                                bind:value={currentPassword}
                                placeholder="当前密码"
                                class="password-input"
                            />

                            <input
                                id="newPassword"
                                type="password"
                                bind:value={newPassword}
                                placeholder="新密码（至少 6 位）"
                                class="password-input"
                                class:error={passwordError}
                            />

                            <input
                                id="confirmPassword"
                                type="password"
                                bind:value={confirmPassword}
                                placeholder="确认新密码"
                                class="password-input"
                            />
                        </div>

                        {#if passwordError}
                            <div class="error-message">{passwordError}</div>
                        {/if}

                        <div class="form-actions">
                            <button class="action-btn primary" on:click={handleChangePassword} disabled={isLoading}>
                                {#if isLoading}保存中...{:else}保存{/if}
                            </button>
                            <button class="action-btn outline" on:click={() => {
                                passwordError = '';
                                currentPassword = '';
                                newPassword = '';
                                confirmPassword = '';
                            }}>
                                取消
                            </button>
                        </div>
                    </div>
                </div>
            </section>
        {/if}

        <!-- 偏好设置 Tab -->
        {#if activeTab === 'settings'}
            <section class="profile-section">
                <div class="settings-group">
                    <div class="setting-item">
                        <div class="setting-info">
                            <span class="setting-label">外观模式</span>
                            <span class="setting-desc">切换亮色/暗色主题</span>
                        </div>
                        <div class="setting-control">
                            <button class="toggle-btn" on:click={handleThemeChange}>
                                {#if themeMode === 'light'}
                                    ☀️ 亮色模式
                                {:else}
                                    🌙 暗色模式
                                {/if}
                            </button>
                        </div>
                    </div>
                </div>

                <div class="settings-group">
                    <div class="setting-item">
                        <div class="setting-info">
                            <span class="setting-label">默认页面大小</span>
                            <span class="setting-desc">列表页面每页显示的数据条数</span>
                        </div>
                        <div class="setting-control">
                            <select bind:value={pageSize} on:change={handlePageSizeChange}>
                                <option value={10}>10 条</option>
                                <option value={20}>20 条</option>
                                <option value={50}>50 条</option>
                                <option value={100}>100 条</option>
                            </select>
                        </div>
                    </div>

                    <div class="setting-item">
                        <div class="setting-info">
                            <span class="setting-label">通知提醒</span>
                            <span class="setting-desc">显示操作结果通知</span>
                        </div>
                        <div class="setting-control">
                            <label class="switch">
                                <input type="checkbox" bind:checked={showNotifications} on:change={handleNotificationToggle}>
                                <span class="slider"></span>
                            </label>
                        </div>
                    </div>

                    <div class="setting-item">
                        <div class="setting-info">
                            <span class="setting-label">紧凑模式</span>
                            <span class="setting-desc">减小元素间距，显示更多内容</span>
                        </div>
                        <div class="setting-control">
                            <label class="switch">
                                <input type="checkbox" bind:checked={compactMode} on:change={handleCompactModeToggle}>
                                <span class="slider"></span>
                            </label>
                        </div>
                    </div>
                </div>
            </section>
        {/if}
    </div>
</div>

<style>
    .profile-container {
        padding: 1.5rem;
        max-width: 900px;
        margin: 0;
        height: calc(100vh - 100px);
        overflow-y: auto;
        scroll-behavior: smooth;
        background-color: var(--bg-primary);
    }

    .profile-container::-webkit-scrollbar {
        width: 8px;
    }

    .profile-container::-webkit-scrollbar-track {
        background: transparent;
    }

    .profile-container::-webkit-scrollbar-thumb {
        background: var(--border-color);
        border-radius: 4px;
    }

    .profile-container::-webkit-scrollbar-thumb:hover {
        background: var(--border-color-light);
    }

    .profile-header {
        margin-bottom: 2rem;
    }

    .profile-header h1 {
        font-size: 1.75rem;
        font-weight: 600;
        margin-bottom: 0.5rem;
        color: var(--text-primary);
    }

    .subtitle {
        color: var(--text-secondary);
        font-size: 0.95rem;
    }

    /* Tab 导航 */
    .profile-tabs {
        display: flex;
        gap: 0.5rem;
        margin-bottom: 1.5rem;
        border-bottom: 2px solid var(--border-color);
        padding-bottom: 0;
    }

    .tab-btn {
        padding: 0.75rem 1.5rem;
        background: transparent;
        border: none;
        border-bottom: 2px solid transparent;
        border-radius: var(--radius-md) var(--radius-md) 0 0;
        color: var(--text-secondary);
        font-size: 0.95rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
        margin-bottom: -2px;
    }

    .tab-btn:hover {
        color: var(--text-primary);
        background: var(--bg-tertiary);
    }

    .tab-btn.active {
        color: var(--color-primary);
        border-bottom-color: var(--color-primary);
        background: transparent;
    }

    /* 基本信息列表 */
    .info-list {
        display: flex;
        flex-direction: column;
        gap: 0;
        margin-bottom: 2rem;
    }

    .info-row {
        display: flex;
        align-items: center;
        gap: 1.5rem;
        padding: 1rem 0;
        border-bottom: 1px solid var(--border-color-light);
    }

    .info-row:first-child {
        padding-top: 0;
    }

    .info-row:last-child {
        border-bottom: none;
        padding-bottom: 0;
    }

    .info-label {
        min-width: 100px;
        font-weight: 500;
        color: var(--text-secondary);
        font-size: 0.95rem;
        flex-shrink: 0;
    }

    .info-value {
        flex: 1;
        color: var(--text-primary);
        font-size: 0.95rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .edit-field {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex: 1;
        max-width: 500px;
    }

    .edit-field input {
        flex: 1;
        padding: 0.5rem 0.75rem;
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        background: var(--bg-primary);
        color: var(--text-primary);
        font-size: 0.95rem;
    }

    .edit-field input.error {
        border-color: var(--color-danger);
    }

    .field-actions {
        display: flex;
        gap: 0.5rem;
        flex-shrink: 0;
    }

    .btn-sm {
        padding: 0.5rem 1rem;
        font-size: 0.9rem;
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: var(--transition-fast);
        font-weight: 500;
    }

    .btn-sm.btn-primary {
        background: var(--color-primary);
        color: white;
        border: none;
    }

    .btn-sm.btn-primary:hover:not(:disabled) {
        background: var(--color-primary-dark);
    }

    .btn-sm.btn-primary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .btn-sm.btn-outline {
        background: transparent;
        color: var(--text-primary);
        border: 1px solid var(--border-color);
    }

    .btn-sm.btn-outline:hover {
        background: var(--bg-tertiary);
    }

    .btn-icon {
        background: transparent;
        border: none;
        cursor: pointer;
        font-size: 1rem;
        padding: 0.25rem;
        border-radius: var(--radius-sm);
        transition: var(--transition-fast);
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .btn-icon:hover {
        background: var(--bg-tertiary);
    }

    /* 账户安全区域 */
    .security-section {
        margin-top: 2rem;
    }

    .security-section .section-title {
        font-size: 1.1rem;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 1rem;
    }

    .password-form {
        max-width: 500px;
    }

    .password-inputs {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .password-input {
        width: 100%;
        padding: 0.75rem 1rem;
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        background: var(--bg-primary);
        color: var(--text-primary);
        font-size: 0.9rem;
        transition: var(--transition-fast);
    }

    .password-input:hover {
        border-color: var(--border-color-light);
    }

    .password-input:focus {
        border-color: var(--color-primary);
        outline: none;
        box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
    }

    .password-input.error {
        border-color: var(--color-danger);
    }

    .error-message {
        color: var(--color-danger);
        font-size: 0.85rem;
        margin-top: 0.75rem;
    }

    .password-form .form-actions {
        display: flex;
        gap: 0.75rem;
        margin-top: 1.25rem;
        padding-top: 1.25rem;
        border-top: 1px solid var(--border-color-light);
    }

    .password-form .action-btn {
        padding: 0.625rem 1.5rem;
        font-size: 0.9rem;
        font-weight: 500;
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: var(--transition-fast);
        border: none;
    }

    .password-form .action-btn.primary {
        background: var(--color-primary);
        color: white;
    }

    .password-form .action-btn.primary:hover:not(:disabled) {
        background: var(--color-primary-dark);
    }

    .password-form .action-btn.outline {
        background: transparent;
        color: var(--text-primary);
        border: 1px solid var(--border-color);
    }

    .password-form .action-btn.outline:hover {
        background: var(--bg-tertiary);
    }

    .password-form .action-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    /* 偏好设置分组 */
    .settings-group {
        margin-bottom: 2rem;
    }

    .settings-group:last-child {
        margin-bottom: 0;
    }

    .profile-content {
        display: flex;
        flex-direction: column;
        gap: 1.5rem;
    }

    .profile-section {
        background: var(--bg-secondary);
        border-radius: 8px;
        padding: 1.5rem;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        border: 1px solid var(--border-color);
    }

    .badge {
        display: inline-block;
        padding: 0.25rem 0.75rem;
        background: var(--color-primary);
        color: white;
        border-radius: 999px;
        font-size: 0.85rem;
        font-weight: 500;
    }

    .setting-item {
        display: flex;
        justify-content: flex-start;
        align-items: center;
        padding: 1rem 0;
        border-bottom: 1px solid var(--border-color-light);
        gap: 1rem;
    }

    .setting-item:last-child {
        border-bottom: none;
        padding-bottom: 0;
    }

    .setting-item:first-child {
        padding-top: 0;
    }

    .setting-info {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        min-width: 200px;
    }

    .setting-label {
        font-size: 0.95rem;
        font-weight: 500;
        color: var(--text-primary);
    }

    .setting-desc {
        font-size: 0.85rem;
        color: var(--text-secondary);
    }

    .setting-control {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: flex-start;
        min-width: 150px;
    }

    .toggle-btn {
        padding: 0.5rem 1rem;
        background: var(--bg-tertiary);
        color: var(--text-primary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        font-size: 0.9rem;
        font-weight: 500;
        cursor: pointer;
        transition: var(--transition-fast);
        min-width: 120px;
        text-align: center;
    }

    .toggle-btn:hover {
        background: var(--bg-secondary);
        border-color: var(--color-primary);
        color: var(--color-primary);
    }

    select {
        padding: 0.5rem 1rem;
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        background: var(--bg-primary);
        color: var(--text-primary);
        font-size: 0.9rem;
        font-weight: 500;
        cursor: pointer;
        min-width: 120px;
    }

    select:hover {
        border-color: var(--color-primary);
    }

    /* Switch 开关样式 */
    .switch {
        position: relative;
        display: inline-block;
        width: 48px;
        height: 26px;
    }

    .switch input {
        opacity: 0;
        width: 0;
        height: 0;
    }

    .slider {
        position: absolute;
        cursor: pointer;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background-color: var(--border-color);
        transition: 0.3s;
        border-radius: 26px;
    }

    .slider:before {
        position: absolute;
        content: "";
        height: 20px;
        width: 20px;
        left: 3px;
        bottom: 3px;
        background-color: var(--bg-secondary);
        transition: 0.3s;
        border-radius: 50%;
    }

    input:checked + .slider {
        background-color: var(--color-primary);
    }

    input:checked + .slider:before {
        transform: translateX(22px);
    }

    /* 密码表单 */
    .password-form {
        display: flex;
        flex-direction: column;
        gap: 1rem;
        padding: 1rem;
        background: var(--bg-tertiary);
        border-radius: 6px;
        border: 1px solid var(--border-color);
    }

    .form-group {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .form-group label {
        font-size: 0.9rem;
        font-weight: 500;
        color: var(--text-primary);
    }

    .form-group input {
        padding: 0.6rem 1rem;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        background: var(--bg-secondary);
        color: var(--text-primary);
        font-size: 0.9rem;
    }

    .form-group input:focus {
        outline: none;
        border-color: var(--color-primary);
    }

    .form-actions {
        display: flex;
        gap: 0.5rem;
        margin-top: 0.5rem;
    }
</style>
