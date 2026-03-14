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
    
    app.subscribe(state => {
        sidebarOpen = state.sidebarOpen;
        activeTab = state.activeTab;
    });
    
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();
    
    const menuItems = [
        { id: 'home', label: '控制台', icon: '🏠' },
        { id: 'pcapList', label: 'PCAP 流量包', icon: '📦' },
        { id: 'analyzer', label: '协议分析引擎', icon: '🔍' },
        { id: 'zeek', label: 'Zeek 入侵检测', icon: '🛡️' },
        { id: 'ai', label: 'Dify 智能诊断', icon: '🧠' },
        { id: 'users', label: '用户管理', icon: '👥', adminOnly: true },
        { id: 'profile', label: '个人中心', icon: '👤' }
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
        <div class="sidebar-header">
            <h1 class="logo">
                <span class="logo-icon">🌌</span>
                {#if sidebarOpen}
                    <span class="logo-text">Nebula</span>
                {/if}
            </h1>
        </div>
        
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
    </aside>
    
    <!-- 主内容区 -->
    <main class="main-content">
        <!-- 顶部栏 -->
        <header class="top-bar">
            <button class="menu-toggle" on:click={toggleSidebar}>
                {sidebarOpen ? '◀' : '▶'}
            </button>
            
            <div class="top-bar-actions">
                {#if user}
                    <div class="user-menu-container">
                        <button class="user-menu-trigger" on:click|stopPropagation={toggleUserMenu}>
                            <div class="user-avatar-small">{user.username.charAt(0).toUpperCase()}</div>
                            <span class="user-name-small">{user.username}</span>
                            <span class="dropdown-arrow">{showUserMenu ? '▲' : '▼'}</span>
                        </button>
                        
                        {#if showUserMenu}
                            <div class="user-menu-dropdown" on:click|stopPropagation>
                                <div class="dropdown-header">
                                    <div class="dropdown-avatar">{user.username.charAt(0).toUpperCase()}</div>
                                    <div class="dropdown-info">
                                        <div class="dropdown-username">{user.username}</div>
                                        <div class="dropdown-role">{user.role}</div>
                                    </div>
                                </div>
                                <div class="dropdown-divider"></div>
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
        </header>
        
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
        width: 260px;
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
    
    .sidebar-header {
        padding: var(--spacing-lg);
        border-bottom: 1px solid var(--border-color);
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
    
    .nav-item {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-md);
        margin-bottom: var(--spacing-sm);
        border-radius: var(--radius-md);
        color: var(--text-secondary);
        text-decoration: none;
        transition: var(--transition-fast);
        cursor: pointer;
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
        text-align: center;
    }
    
    .nav-label {
        font-size: var(--font-sm);
        font-weight: 500;
    }
    
    /* 主内容区 */
    .main-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }
    
    .top-bar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--spacing-md) var(--spacing-lg);
        background: var(--bg-secondary);
        border-bottom: 1px solid var(--border-color);
    }
    
    .menu-toggle {
        background: transparent;
        border: none;
        color: var(--text-secondary);
        font-size: var(--font-lg);
        cursor: pointer;
        padding: var(--spacing-sm);
        border-radius: var(--radius-md);
        transition: var(--transition-fast);
    }
    
    .menu-toggle:hover {
        background-color: var(--bg-tertiary);
        color: var(--text-primary);
    }
    
    .top-bar-actions {
        display: flex;
        align-items: center;
        gap: var(--spacing-lg);
    }
    
    /* 用户菜单样式 */
    .user-menu-container {
        position: relative;
    }
    
    .user-menu-trigger {
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        padding: var(--spacing-sm) var(--spacing-md);
        background: var(--bg-tertiary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: var(--transition-fast);
    }
    
    .user-menu-trigger:hover {
        background: var(--bg-secondary);
        border-color: var(--color-primary);
    }
    
    .user-avatar-small {
        width: 28px;
        height: 28px;
        border-radius: var(--radius-full);
        background: linear-gradient(135deg, #4f46e5 0%, #10b981 100%);
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.85rem;
        font-weight: 600;
    }
    
    .user-name-small {
        color: var(--text-primary);
        font-size: var(--font-sm);
        font-weight: 500;
    }
    
    .dropdown-arrow {
        color: var(--text-secondary);
        font-size: 0.7rem;
        margin-left: var(--spacing-xs);
    }
    
    .user-menu-dropdown {
        position: absolute;
        top: calc(100% + var(--spacing-sm));
        right: 0;
        width: 280px;
        background: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-lg);
        box-shadow: var(--shadow-xl);
        z-index: var(--z-dropdown);
        overflow: hidden;
        animation: slideDown 0.2s ease;
    }
    
    @keyframes slideDown {
        from {
            opacity: 0;
            transform: translateY(-8px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    
    .dropdown-header {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-lg);
        background: var(--bg-tertiary);
        border-bottom: 1px solid var(--border-color);
    }
    
    .dropdown-avatar {
        width: 48px;
        height: 48px;
        border-radius: var(--radius-full);
        background: linear-gradient(135deg, #4f46e5 0%, #10b981 100%);
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.25rem;
        font-weight: 600;
        flex-shrink: 0;
    }
    
    .dropdown-info {
        flex: 1;
        overflow: hidden;
    }
    
    .dropdown-username {
        color: var(--text-primary);
        font-size: var(--font-base);
        font-weight: 600;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    
    .dropdown-role {
        color: var(--text-secondary);
        font-size: var(--font-sm);
        margin-top: 2px;
    }
    
    .dropdown-divider {
        height: 1px;
        background: var(--border-color);
    }
    
    .dropdown-item {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        width: 100%;
        padding: var(--spacing-md) var(--spacing-lg);
        background: transparent;
        border: none;
        color: var(--text-primary);
        font-size: var(--font-sm);
        text-align: left;
        cursor: pointer;
        transition: var(--transition-fast);
    }
    
    .dropdown-item:hover {
        background: var(--bg-tertiary);
        color: var(--color-primary);
    }
    
    .dropdown-item.logout {
        color: var(--text-secondary);
    }
    
    .dropdown-item.logout:hover {
        background: rgba(239, 68, 68, 0.1);
        color: var(--color-danger);
    }
    
    .dropdown-icon {
        font-size: var(--font-base);
    }
    
    .content {
        flex: 1;
        overflow-y: auto;
        padding: var(--spacing-lg);
        background: var(--bg-primary);
    }
</style>
