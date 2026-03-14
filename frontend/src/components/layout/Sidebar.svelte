<script lang="ts">
    import { auth } from '../../stores/auth';
    import { app } from '../../stores/app';
    import { showInfo } from '../../utils/helpers';
    
    let user;
    let sidebarOpen = true;
    let activeTab = 'home';
    let showUserMenu = false;
    
    auth.subscribe(state => {
        user = state.user;
    });
    
    // 订阅应用状态，个人中心将作为独立模态框显示
    app.subscribe(state => {
        sidebarOpen = state.sidebarOpen;
        activeTab = state.activeTab;
    });
    
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();
    
    const menuItems = [
        { id: 'home', label: '主页', icon: '🏠' },
        { id: 'analyzer', label: '协议分析引擎', icon: '🔍' },
        { id: 'zeek', label: 'Zeek 入侵检测', icon: '🛡️' },
        { id: 'ai', label: 'Dify 智能诊断', icon: '🧠' },
        { id: 'users', label: '用户管理', icon: '👥', adminOnly: true },
    ];
    
    // 检查是否是管理员
    function isAdmin(): boolean {
        return user && user.roleCode === 'admin';
    }
    
    // 过滤菜单项
    function getVisibleMenuItems() {
        return menuItems.filter(item => {
            if (item.adminOnly && !isAdmin()) {
                return false;
            }
            return true;
        });
    }
    
    function handleLogout() {
        showUserMenu = false;
        auth.logout();
        showInfo('已退出登录');
    }
    
    function goToProfile() {
        showUserMenu = false;
        switchTab('profile');
    }
    
    function toggleSidebar() {
        app.toggleSidebar();
    }
    
    function switchTab(tabId: string) {
        app.setActiveTab(tabId);
        dispatch('tabChange', tabId);
    }
    
    function toggleUserMenu() {
        showUserMenu = !showUserMenu;
    }
    
    // 点击外部关闭菜单
    function handleClickOutside(event: MouseEvent) {
        const target = event.target as HTMLElement;
        if (showUserMenu && !target.closest('.user-menu-container')) {
            showUserMenu = false;
        }
    }
</script>

<div class="layout" on:click={handleClickOutside}>
    <!-- 侧边栏 -->
    <aside class="sidebar" class:collapsed={!sidebarOpen}>
        <div class="sidebar-top">
            <!-- Logo -->
            <div class="sidebar-header">
                <h1 class="logo">
                    <span class="logo-icon">🌌</span>
                    {#if sidebarOpen}
                        <span class="logo-text">Nebula</span>
                    {/if}
                </h1>
            </div>
            
            <!-- 折叠按钮 -->
            <button class="menu-toggle-sidebar" on:click={toggleSidebar}>
                {sidebarOpen ? '◀ 收起' : '▶'}
            </button>
            
            <!-- 导航菜单 -->
            <nav class="sidebar-nav">
                {#each getVisibleMenuItems() as item}
                    <a
                        href="#"
                        class="nav-item"
                        class:active={activeTab === item.id}
                        on:click|preventDefault={() => switchTab(item.id)}
                    >
                        <span class="nav-icon">{item.icon}</span>
                        {#if sidebarOpen}
                            <span class="nav-label">{item.label}</span>
                        {/if}
                    </a>
                {/each}
            </nav>
        </div>
        
        <!-- 底部用户信息 -->
        <div class="sidebar-footer">
            {#if user}
                <div class="user-menu-container">
                    <button class="user-menu-trigger" on:click|stopPropagation={toggleUserMenu}>
                        <div class="user-avatar">{user.username.charAt(0).toUpperCase()}</div>
                        {#if sidebarOpen}
                            <div class="user-info">
                                <div class="user-username">{user.username}</div>
                                <div class="user-role">{user.role}</div>
                            </div>
                            <span class="dropdown-arrow">{showUserMenu ? '▲' : '▼'}</span>
                        {/if}
                    </button>
                    
                    {#if showUserMenu && sidebarOpen}
                        <div class="user-menu-dropdown" on:click|stopPropagation>
                            <button class="dropdown-item" on:click={goToProfile}>
                                <span class="dropdown-icon">👤</span>
                                个人中心
                            </button>
                            <button class="dropdown-item logout" on:click={handleLogout}>
                                <span class="dropdown-icon">🚪</span>
                                退出登录
                            </button>
                        </div>
                    {/if}
                </div>
            {/if}
        </div>
    </aside>
    
    <!-- 主内容区 -->
    <main class="main-content">
        <!-- 内容区 -->
        <div class="content">
            <slot></slot>
        </div>
    </main>
</div>

<style>
    .layout {
        display: flex;
        height: 100vh;
        overflow: hidden;
    }
    
    /* 侧边栏 */
    .sidebar {
        width: 220px;
        background: var(--bg-secondary);
        border-right: 1px solid var(--border-color);
        display: flex;
        flex-direction: column;
        transition: width var(--transition-base);
        overflow: hidden;
    }
    
    .sidebar.collapsed {
        width: 70px;
    }
    
    .sidebar.collapsed .menu-toggle-sidebar {
        margin: 8px auto;
        width: 36px;
        height: 36px;
        padding: 0;
    }
    
    .sidebar.collapsed .sidebar-header {
        padding: 12px 8px;
    }
    
    .sidebar.collapsed .nav-item {
        margin: 0 auto;
        padding: 10px;
        justify-content: center;
    }
    
    .sidebar.collapsed .nav-label {
        display: none;
    }
    
    .sidebar-header {
        padding: 12px 16px;
        border-bottom: 1px solid var(--border-color);
        display: flex;
        justify-content: center;
        width: 100%;
    }
    
    .sidebar-top {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
        width: 100%;
    }
    
    .menu-toggle-sidebar {
        padding: 6px 10px;
        margin: 8px 4px;
        background: var(--bg-tertiary);
        border: 1px solid var(--border-color);
        color: var(--text-secondary);
        cursor: pointer;
        font-size: 0.8rem;
        font-weight: 500;
        transition: var(--transition-fast);
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: var(--radius-md);
        width: calc(100% - 8px);
        align-self: center;
    }
    
    .menu-toggle-sidebar:hover {
        background: var(--bg-secondary);
        color: var(--color-primary);
        border-color: var(--color-primary);
    }
    
    .logo {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        margin: 0;
        font-size: var(--font-xl);
        color: var(--text-primary);
    }
    
    .logo-icon {
        font-size: var(--font-2xl);
    }
    
    .logo-text {
        font-weight: 700;
        background: linear-gradient(135deg, #4f46e5 0%, #10b981 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }
    
    .sidebar-nav {
        flex: 1;
        padding: var(--spacing-md);
        overflow-y: auto;
    }
    
    .sidebar-footer {
        padding: var(--spacing-md);
        border-top: 1px solid var(--border-color);
        background: var(--bg-tertiary);
    }
    
    .user-menu-container {
        position: relative;
    }
    
    .user-menu-trigger {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-md);
        width: 100%;
        background: transparent;
        border: none;
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: var(--transition-fast);
        color: var(--text-primary);
    }
    
    .user-menu-trigger:hover {
        background: var(--bg-secondary);
    }
    
    .user-avatar {
        width: 36px;
        height: 36px;
        border-radius: 50%;
        background: var(--color-primary);
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 600;
        font-size: var(--font-base);
        flex-shrink: 0;
    }
    
    .user-info {
        flex: 1;
        text-align: left;
        overflow: hidden;
    }
    
    .user-username {
        font-size: var(--font-sm);
        font-weight: 600;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    
    .user-role {
        font-size: var(--font-xs);
        color: var(--text-secondary);
        margin-top: 2px;
    }
    
    .dropdown-arrow {
        font-size: var(--font-xs);
        color: var(--text-secondary);
    }
    
    .user-menu-dropdown {
        position: absolute;
        bottom: 100%;
        left: 0;
        right: 0;
        background: var(--bg-primary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.15);
        margin-bottom: var(--spacing-sm);
        overflow: hidden;
        z-index: 100;
    }
    
    .dropdown-divider {
        height: 1px;
        background: var(--border-color);
        margin: var(--spacing-sm) 0;
    }
    
    .dropdown-item {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-md);
        width: 100%;
        background: transparent;
        border: none;
        color: var(--text-primary);
        font-size: var(--font-sm);
        cursor: pointer;
        transition: var(--transition-fast);
        text-align: left;
    }
    
    .dropdown-item:hover {
        background: var(--bg-tertiary);
    }
    
    .dropdown-item.logout {
        color: var(--color-danger);
    }
    
    .dropdown-icon {
        font-size: var(--font-base);
    }
    
    .nav-item {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: var(--spacing-md);
        padding: 8px 10px;
        margin: 0 4px var(--spacing-sm) 4px;
        border-radius: var(--radius-md);
        color: var(--text-secondary);
        text-decoration: none;
        transition: var(--transition-fast);
        cursor: pointer;
        min-width: fit-content;
    }
    
    .nav-item:hover {
        background-color: var(--bg-tertiary);
        color: var(--text-primary);
    }
    
    .nav-item.active {
        background-color: var(--color-primary-light);
        color: var(--color-primary);
    }
    
    .nav-icon {
        font-size: var(--font-lg);
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    
    .nav-label {
        font-size: var(--font-sm);
        font-weight: 500;
        white-space: nowrap;
    }
    
    /* 主内容区 */
    .main-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }
    
    .content {
        flex: 1;
        overflow-y: auto;
        background: var(--bg-primary);
    }
</style>
