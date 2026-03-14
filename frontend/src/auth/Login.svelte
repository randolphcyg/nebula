<script lang="ts">
    import { createEventDispatcher, onMount } from 'svelte';
    import { Login as LoginApi, GetPublicKey } from '../../wailsjs/go/main/App';
    import { auth } from '$lib/stores/auth';
    import { handleError } from '$lib/utils';
    import { logger } from '$lib/utils/logger';
    import JSEncrypt from 'jsencrypt';
    
    let username = '';
    let password = '';
    let isLoading = false;
    let error = '';
    let publicKey = '';
    
    const dispatch = createEventDispatcher();
    
    // 获取公钥
    async function loadPublicKey() {
        try {
            const base64Key = await GetPublicKey();
            // 后端返回的是 Base64 编码的 PEM，需要解码
            publicKey = atob(base64Key);
            logger.success('公钥已加载', publicKey.substring(0, 50) + '...');
        } catch (err) {
            logger.error('获取公钥失败:', err);
            error = '安全组件初始化失败，请刷新页面';
        }
    }
    
    // 加密密码
    function encryptPassword(pwd: string): string {
        if (!publicKey) {
            throw new Error('公钥未加载');
        }
        const encryptor = new JSEncrypt();
        encryptor.setPublicKey(publicKey);
        const encrypted = encryptor.encrypt(pwd);
        if (!encrypted) {
            throw new Error('密码加密失败');
        }
        logger.success('密码已加密');
        return encrypted;
    }
    
    async function handleLogin() {
        error = '';
        
        if (!username || !password) {
            error = '请输入用户名和密码';
            return;
        }
        
        // 加密密码
        let encryptedPassword: string;
        try {
            encryptedPassword = encryptPassword(password);
            console.log('✅ 密码已加密');
        } catch (err: any) {
            error = err.message;
            return;
        }
        
        isLoading = true;
        
        try {
            const response = await LoginApi(username, encryptedPassword);
            
            if (response && response.token) {
                auth.login(response.token, response.user);
                dispatch('success', response.user);
            } else {
                error = '登录失败：' + (response.msg || '未知错误');
            }
        } catch (err: any) {
            // 检查是否是特定的错误信息
            const errorMsg = err.message || err.error || String(err);
            if (errorMsg.includes('用户已被禁用')) {
                error = '该用户已被禁用，请联系管理员';
            } else if (errorMsg.includes('用户待审核')) {
                error = '该用户正在等待审核，请等待管理员批准';
            } else {
                error = handleError(err, '登录失败，请稍后重试');
            }
        } finally {
            isLoading = false;
        }
    }
    
    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === 'Enter') {
            handleLogin();
        }
    }
    
    // 组件挂载时加载公钥
    onMount(() => {
        loadPublicKey();
    });
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
            <p class="register-link">
                还没有账户？
                <a href="#" on:click|preventDefault={() => dispatch('showRegister')}>
                    立即注册
                </a>
            </p>
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
