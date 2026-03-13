<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { Login as LoginApi } from '../../wailsjs/go/main/App';
    import { auth } from '$lib/stores/auth';
    import { handleError } from '$lib/utils/helpers';
    
    let username = '';
    let password = '';
    let isLoading = false;
    let error = '';
    
    const dispatch = createEventDispatcher();
    
    async function handleLogin() {
        error = '';
        
        if (!username || !password) {
            error = '请输入用户名和密码';
            return;
        }
        
        isLoading = true;
        
        try {
            const response = await LoginApi(username, password);
            
            if (response && response.token) {
                auth.login(response.token, response.user);
                dispatch('success', response.user);
            } else {
                error = '登录失败：' + (response.msg || '未知错误');
            }
        } catch (err: any) {
            error = handleError(err, '登录失败，请稍后重试');
        } finally {
            isLoading = false;
        }
    }
    
    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === 'Enter') {
            handleLogin();
        }
    }
</script>

<div class="login-container">
    <div class="login-card">
        <div class="login-header">
            <h1 class="login-title">Nebula</h1>
            <p class="login-subtitle">网络流量分析平台</p>
        </div>
        
        <form class="login-form" on:submit|preventDefault={handleLogin}>
            {#if error}
                <div class="error-message">
                    <span>⚠️</span> {error}
                </div>
            {/if}
            
            <div class="form-group">
                <label for="username">用户名</label>
                <input
                    id="username"
                    type="text"
                    class="input"
                    bind:value={username}
                    on:keypress={handleKeyPress}
                    placeholder="请输入用户名"
                    autocomplete="username"
                    disabled={isLoading}
                />
            </div>
            
            <div class="form-group">
                <label for="password">密码</label>
                <input
                    id="password"
                    type="password"
                    class="input"
                    bind:value={password}
                    on:keypress={handleKeyPress}
                    placeholder="请输入密码"
                    autocomplete="current-password"
                    disabled={isLoading}
                />
            </div>
            
            <button 
                type="submit" 
                class="login-btn"
                disabled={isLoading}
            >
                {#if isLoading}
                    <span class="spinner"></span>
                    登录中...
                {:else}
                    登录
                {/if}
            </button>
        </form>
        
        <div class="login-footer">
            <p class="hint">默认账号：admin / admin123</p>
        </div>
    </div>
</div>

<style>
    .login-container {
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
        padding: var(--spacing-lg);
    }
    
    .login-card {
        width: 100%;
        max-width: 420px;
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-xl);
        padding: var(--spacing-2xl);
        box-shadow: var(--shadow-xl);
    }
    
    .login-header {
        text-align: center;
        margin-bottom: var(--spacing-2xl);
    }
    
    .login-title {
        font-size: var(--font-3xl);
        font-weight: 700;
        color: var(--text-primary);
        margin: 0 0 var(--spacing-sm) 0;
        background: linear-gradient(135deg, #4f46e5 0%, #10b981 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }
    
    .login-subtitle {
        color: var(--text-secondary);
        font-size: var(--font-base);
        margin: 0;
    }
    
    .login-form {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-lg);
    }
    
    .form-group {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-sm);
    }
    
    .form-group label {
        color: var(--text-secondary);
        font-size: var(--font-sm);
        font-weight: 500;
    }
    
    .input {
        padding: 0.75rem 1rem;
        font-size: var(--font-base);
        background-color: var(--bg-secondary);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-md);
        color: var(--text-primary);
        transition: var(--transition-base);
    }
    
    .input:focus {
        outline: none;
        border-color: var(--color-primary);
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    .input::placeholder {
        color: var(--text-muted);
    }
    
    .error-message {
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        padding: var(--spacing-md);
        background-color: var(--color-danger-light);
        border: 1px solid var(--color-danger);
        border-radius: var(--radius-md);
        color: var(--color-danger);
        font-size: var(--font-sm);
    }
    
    .login-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: var(--spacing-sm);
        padding: 0.875rem 1.5rem;
        font-size: var(--font-base);
        font-weight: 600;
        color: var(--color-white);
        background-color: var(--color-primary);
        border: none;
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: var(--transition-base);
        margin-top: var(--spacing-md);
    }
    
    .login-btn:hover:not(:disabled) {
        background-color: var(--color-primary-hover);
        transform: translateY(-1px);
        box-shadow: var(--shadow-lg);
    }
    
    .login-btn:disabled {
        opacity: 0.7;
        cursor: not-allowed;
    }
    
    .spinner {
        width: 16px;
        height: 16px;
        border: 2px solid rgba(255, 255, 255, 0.3);
        border-top-color: white;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }
    
    @keyframes spin {
        to { transform: rotate(360deg); }
    }
    
    .login-footer {
        margin-top: var(--spacing-xl);
        text-align: center;
    }
    
    .hint {
        color: var(--text-muted);
        font-size: var(--font-xs);
        margin: 0;
    }
</style>
