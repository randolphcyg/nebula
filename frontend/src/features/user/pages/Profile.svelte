<script lang="ts">
    import { onMount } from 'svelte';
    import { auth } from '../../../stores/auth';
    import { themeStore } from '../../../stores/theme';
    import { preferencesStore } from '../../../stores/preferences';
    import { error as showError, success as showSuccess } from '../../../stores/toast';
    import { ChangePassword } from '../../../../wailsjs/go/main/App';
    import type { User } from '../../../types';

    let user: User | null = null;
    let isLoading = false;

    // 修改密码表单
    let currentPassword = '';
    let newPassword = '';
    let confirmPassword = '';
    let showPasswordForm = false;

    // 主题设置
    let themeMode: 'light' | 'dark' = 'light';

    // 偏好设置
    let pageSize = 10;
    let showNotifications = true;
    let compactMode = false;

    onMount(() => {
        user = auth.getCurrentUser();
        themeStore.subscribe(theme => {
            themeMode = theme.mode;
        });
        preferencesStore.subscribe(prefs => {
            pageSize = prefs.pageSize;
            showNotifications = prefs.showNotifications;
            compactMode = prefs.compactMode;
        });
    });

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
            showPasswordForm = false;
            
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
</script>

<div class="profile-container">
    <div class="profile-header">
        <h1>个人中心</h1>
        <p class="subtitle">管理您的账户信息和偏好设置</p>
    </div>

    <div class="profile-content">
        <!-- 基本信息 -->
        <section class="profile-section">
            <h2>👤 基本信息</h2>
            <div class="info-grid">
                <div class="info-item">
                    <label>用户名</label>
                    <div class="value">{user?.username || '-'}</div>
                </div>
                <div class="info-item">
                    <label>邮箱</label>
                    <div class="value">{user?.email || '-'}</div>
                </div>
                <div class="info-item">
                    <label>角色</label>
                    <div class="value">
                        <span class="badge">{user?.role || '-'}</span>
                    </div>
                </div>
                <div class="info-item">
                    <label>上次登录</label>
                    <div class="value">{user?.lastLogin ? new Date(user.lastLogin).toLocaleString('zh-CN') : '-'}</div>
                </div>
            </div>
        </section>

        <!-- 主题设置 -->
        <section class="profile-section">
            <h2>🎨 主题设置</h2>
            <div class="setting-item">
                <div class="setting-info">
                    <span class="setting-label">外观模式</span>
                    <span class="setting-desc">切换亮色/暗色主题</span>
                </div>
                <button class="toggle-btn" on:click={handleThemeChange}>
                    {#if themeMode === 'light'}
                        ☀️ 亮色模式
                    {:else}
                        🌙 暗色模式
                    {/if}
                </button>
            </div>
        </section>

        <!-- 偏好设置 -->
        <section class="profile-section">
            <h2>⚙️ 偏好设置</h2>
            
            <div class="setting-item">
                <div class="setting-info">
                    <span class="setting-label">默认页面大小</span>
                    <span class="setting-desc">列表页面每页显示的数据条数</span>
                </div>
                <select bind:value={pageSize} on:change={handlePageSizeChange}>
                    <option value={10}>10 条</option>
                    <option value={20}>20 条</option>
                    <option value={50}>50 条</option>
                    <option value={100}>100 条</option>
                </select>
            </div>

            <div class="setting-item">
                <div class="setting-info">
                    <span class="setting-label">通知提醒</span>
                    <span class="setting-desc">显示操作结果通知</span>
                </div>
                <label class="switch">
                    <input type="checkbox" bind:checked={showNotifications} on:change={handleNotificationToggle}>
                    <span class="slider"></span>
                </label>
            </div>

            <div class="setting-item">
                <div class="setting-info">
                    <span class="setting-label">紧凑模式</span>
                    <span class="setting-desc">减小元素间距，显示更多内容</span>
                </div>
                <label class="switch">
                    <input type="checkbox" bind:checked={compactMode} on:change={handleCompactModeToggle}>
                    <span class="slider"></span>
                </label>
            </div>
        </section>

        <!-- 修改密码 -->
        <section class="profile-section">
            <h2>🔐 账户安全</h2>
            
            {#if !showPasswordForm}
                <button class="action-btn" on:click={() => showPasswordForm = true}>
                    修改密码
                </button>
            {:else}
                <div class="password-form">
                    <div class="form-group">
                        <label for="currentPassword">当前密码</label>
                        <input
                            id="currentPassword"
                            type="password"
                            bind:value={currentPassword}
                            placeholder="请输入当前密码"
                        />
                    </div>

                    <div class="form-group">
                        <label for="newPassword">新密码</label>
                        <input
                            id="newPassword"
                            type="password"
                            bind:value={newPassword}
                            placeholder="请输入新密码（至少 6 位）"
                        />
                    </div>

                    <div class="form-group">
                        <label for="confirmPassword">确认新密码</label>
                        <input
                            id="confirmPassword"
                            type="password"
                            bind:value={confirmPassword}
                            placeholder="请再次输入新密码"
                        />
                    </div>

                    <div class="form-actions">
                        <button class="action-btn" on:click={handleChangePassword} disabled={isLoading}>
                            {#if isLoading}
                                保存中...
                            {:else}
                                保存新密码
                            {/if}
                        </button>
                        <button class="action-btn outline" on:click={() => {
                            showPasswordForm = false;
                            currentPassword = '';
                            newPassword = '';
                            confirmPassword = '';
                        }}>
                            取消
                        </button>
                    </div>
                </div>
            {/if}
        </section>
    </div>
</div>

<style>
    .profile-container {
        padding: 1.5rem;
        max-width: 900px;
        margin: 0 auto;
    }

    .profile-header {
        margin-bottom: 2rem;
    }

    .profile-header h1 {
        font-size: 1.75rem;
        font-weight: 600;
        margin-bottom: 0.5rem;
        color: var(--color-text-primary);
    }

    .subtitle {
        color: var(--color-text-secondary);
        font-size: 0.95rem;
    }

    .profile-content {
        display: flex;
        flex-direction: column;
        gap: 1.5rem;
    }

    .profile-section {
        background: var(--color-bg-card);
        border-radius: 8px;
        padding: 1.5rem;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    }

    .profile-section h2 {
        font-size: 1.1rem;
        font-weight: 600;
        margin-bottom: 1rem;
        color: var(--color-text-primary);
    }

    .info-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 1rem;
    }

    .info-item {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
    }

    .info-item label {
        font-size: 0.85rem;
        color: var(--color-text-secondary);
        font-weight: 500;
    }

    .info-item .value {
        font-size: 0.95rem;
        color: var(--color-text-primary);
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
        justify-content: space-between;
        align-items: center;
        padding: 1rem 0;
        border-bottom: 1px solid var(--color-border);
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
    }

    .setting-label {
        font-size: 0.95rem;
        font-weight: 500;
        color: var(--color-text-primary);
    }

    .setting-desc {
        font-size: 0.85rem;
        color: var(--color-text-secondary);
    }

    .toggle-btn,
    .action-btn {
        padding: 0.5rem 1rem;
        background: var(--color-primary);
        color: white;
        border: none;
        border-radius: 6px;
        font-size: 0.9rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
    }

    .toggle-btn:hover,
    .action-btn:hover {
        opacity: 0.9;
        transform: translateY(-1px);
    }

    .action-btn.outline {
        background: transparent;
        border: 1px solid var(--color-border);
        color: var(--color-text-primary);
    }

    .action-btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    select {
        padding: 0.5rem 1rem;
        border: 1px solid var(--color-border);
        border-radius: 6px;
        background: var(--color-bg-input);
        color: var(--color-text-primary);
        font-size: 0.9rem;
        cursor: pointer;
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
        background-color: var(--color-border);
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
        background-color: white;
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
        background: var(--color-bg-input);
        border-radius: 6px;
    }

    .form-group {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .form-group label {
        font-size: 0.9rem;
        font-weight: 500;
        color: var(--color-text-primary);
    }

    .form-group input {
        padding: 0.6rem 1rem;
        border: 1px solid var(--color-border);
        border-radius: 6px;
        background: var(--color-bg-card);
        color: var(--color-text-primary);
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
