<script lang="ts">
    import { auth } from '../../stores/auth';
    import { app } from '../../stores/app';
    import { showInfo } from '../../utils/helpers';
    
    let user;
    let sidebarOpen = true;
    let activeTab = 'home';
    
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
        { id: 'analyzer', label: '协议分析引擎', icon: '🔍' },
        { id: 'zeek', label: 'Zeek 入侵检测', icon: '🛡️' },
        { id: 'ai', label: 'Dify 智能诊断', icon: '🧠' }
    ];
    
    function handleLogout() {
        auth.logout();
        showInfo('已退出登录');
    }
    
    function toggleSidebar() {
        app.toggleSidebar();
    }
    
    function switchTab(tabId: string) {
        app.setActiveTab(tabId);
        dispatch('tabChange', tabId);
    }
</script>

<div class="layout">
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
            {#each menuItems as item}
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
        
        <div class="sidebar-footer">
            {#if sidebarOpen && user}
                <div class="user-info">
                    <div class="user-avatar">{user.username.charAt(0).toUpperCase()}</div>
                    <div class="user-details">
                        <div class="user-name">{user.username}</div>
                        <div class="user-role">{user.role}</div>
                    </div>
                    <button class="profile-btn" on:click={() => {
                        app.setActiveTab('analyzer');
                        setTimeout(() => {
                            const event = new CustomEvent('navigate', { detail: 'profile' });
                            window.dispatchEvent(event);
                        }, 100);
                    }} title="个人中心">
                        <span class="icon">⚙️</span>
                    </button>
                </div>
            {/if}
            
            <button class="logout-btn" on:click={handleLogout} title="退出登录">
                <span class="icon">🚪</span>
                {#if sidebarOpen}
                    <span>退出</span>
                {/if}
            </button>
        </div>
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
                    <span class="welcome-text">欢迎，{user.username}</span>
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
    
    .sidebar-footer {
        padding: var(--spacing-lg);
        border-top: 1px solid var(--border-color);
        display: flex;
        flex-direction: column;
        gap: var(--spacing-md);
    }
    
    .user-info {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-sm);
        background-color: var(--bg-tertiary);
        border-radius: var(--radius-md);
        flex: 1;
    }
    
    .user-avatar {
        width: 36px;
        height: 36px;
        border-radius: var(--radius-full);
        background: linear-gradient(135deg, #4f46e5 0%, #10b981 100%);
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 600;
        font-size: var(--font-base);
        color: white;
    }
    
    .user-details {
        flex: 1;
        overflow: hidden;
    }
    
    .user-name {
        font-size: var(--font-sm);
        font-weight: 600;
        color: var(--text-primary);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    
    .user-role {
        font-size: var(--font-xs);
        color: var(--text-muted);
    }
    
    .profile-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 36px;
        height: 36px;
        border: none;
        background: transparent;
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: all var(--transition-base);
        font-size: var(--font-lg);
    }
    
    .profile-btn:hover {
        background-color: var(--bg-hover);
        transform: scale(1.1);
    }
    
    .logout-btn {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-sm) var(--spacing-md);
        background: transparent;
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        color: var(--text-secondary);
        font-size: var(--font-sm);
        cursor: pointer;
        transition: var(--transition-fast);
        width: 100%;
    }
    
    .logout-btn:hover {
        background-color: var(--color-danger-light);
        border-color: var(--color-danger);
        color: var(--color-danger);
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
    
    .welcome-text {
        color: var(--text-secondary);
        font-size: var(--font-sm);
    }
    
    .content {
        flex: 1;
        overflow-y: auto;
        padding: var(--spacing-lg);
        background: var(--bg-primary);
    }
</style>
