<script lang="ts">
    import { auth } from '../stores/auth';
    import { success as showSuccess, error as showError } from '../stores/toast';
    import { Register, GetPublicKey } from '../../wailsjs/go/main/App';
    import { createEventDispatcher } from 'svelte';
    import JSEncrypt from 'jsencrypt';
    import { logger } from '../utils/logger';

    const dispatch = createEventDispatcher();

    let username = '';
    let email = '';
    let password = '';
    let confirmPassword = '';
    let isLoading = false;
    let publicKey = '';
    
    // 实时校验状态
    let usernameError = '';
    let emailError = '';
    let passwordError = '';
    let confirmPasswordError = '';
    
    // 校验规则
    function validateUsername(value: string): string {
        if (!value) return '';
        if (value.length < 3) return '用户名至少 3 位';
        if (value.length > 20) return '用户名最多 20 位';
        if (!/^[a-zA-Z0-9_]+$/.test(value)) return '只能包含字母、数字和下划线';
        return '';
    }
    
    function validateEmail(value: string): string {
        if (!value) return '';
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(value)) return '请输入有效的邮箱地址';
        return '';
    }
    
    function validatePassword(value: string): string {
        if (!value) return '';
        if (value.length < 6) return '密码至少 6 位';
        if (value.length > 50) return '密码最多 50 位';
        if (!/[a-zA-Z]/.test(value)) return '密码必须包含字母';
        if (!/[0-9]/.test(value)) return '密码必须包含数字';
        return '';
    }
    
    function validateConfirmPassword(value: string): string {
        if (!value) return '';
        if (value !== password) return '两次输入的密码不一致';
        return '';
    }
    
    // 实时校验触发
    $: usernameError = validateUsername(username);
    $: emailError = validateEmail(email);
    $: passwordError = validatePassword(password);
    $: confirmPasswordError = validateConfirmPassword(confirmPassword);
    
    // 表单是否有效（响应式）
    $: formValid = username && !usernameError && 
                   email && !emailError && 
                   password && !passwordError && 
                   confirmPassword && !confirmPasswordError;
    
    // 获取公钥（组件挂载时）
    async function loadPublicKey() {
        try {
            const base64Key = await GetPublicKey();
            // 后端返回的是 Base64 编码的 PEM，需要解码
            publicKey = atob(base64Key);
            logger.success('公钥已加载', publicKey.substring(0, 50) + '...');
        } catch (err) {
            logger.error('获取公钥失败:', err);
            showError('安全组件初始化失败，请刷新页面');
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

    async function handleRegister() {
        logger.debug('handleRegister called, formValid:', formValid);
        
        // 最终校验（防止绕过前端校验）
        if (!formValid) {
            showError('请修正表单中的错误');
            return;
        }

        // 加密密码
        let encryptedPassword: string;
        try {
            encryptedPassword = encryptPassword(password);
            logger.success('密码已加密');
        } catch (err: any) {
            showError(err.message);
            return;
        }

        logger.debug('Starting registration...', { username, email });
        isLoading = true;
        try {
            await Register(username, email, encryptedPassword);
            logger.info('注册成功');
            showSuccess('注册成功！请等待管理员审核，审核通过后会通知您');
            
            // 清空表单
            username = '';
            email = '';
            password = '';
            confirmPassword = '';
            
            // 切换到登录页面
            setTimeout(() => {
                dispatch('switchToLogin');
            }, 2000);
        } catch (err: any) {
            logger.error('注册失败:', err);
            
            let message = '注册失败，请稍后重试';
            
            // 详细错误信息处理
            if (err.message) {
                const errMsg = err.message.toLowerCase();
                
                if (errMsg.includes('用户名已存在') || errMsg.includes('user already exists')) {
                    message = '❌ 该用户名已被注册，请更换其他用户名';
                } else if (errMsg.includes('邮箱已被注册') || errMsg.includes('email already exists')) {
                    message = '❌ 该邮箱已被注册，请更换其他邮箱或使用其他邮箱注册';
                } else if (errMsg.includes('角色') || errMsg.includes('role')) {
                    message = '⚠️ 系统角色配置错误，请联系管理员初始化系统';
                } else if (errMsg.includes('数据库') || errMsg.includes('database')) {
                    message = '⚠️ 数据库连接失败，请稍后重试或联系管理员';
                } else if (errMsg.includes('网络') || errMsg.includes('network')) {
                    message = '⚠️ 网络连接失败，请检查网络后重试';
                } else {
                    // 其他错误，显示原始错误信息但加上前缀
                    message = `❌ 注册失败：${err.message}`;
                }
            }
            
            showError(message);
            isLoading = false;
        }
    }
    
    // 组件挂载时加载公钥
    import { onMount } from 'svelte';
    onMount(() => {
        loadPublicKey();
    });
</script>

<div class="register-container">
    <div class="register-card">
        <div class="register-header">
            <h1>🌌 创建账户</h1>
            <p>注册 Nebula 账户，开启安全分析之旅</p>
        </div>

        <form class="register-form" on:submit|preventDefault={handleRegister}>
            <div class="form-group">
                <label for="username">用户名</label>
                <input
                    id="username"
                    type="text"
                    bind:value={username}
                    placeholder="请输入用户名（3-20 位字母、数字、下划线）"
                    class:error={usernameError && username}
                    disabled={isLoading}
                />
                {#if username && usernameError}
                    <span class="error-message">❌ {usernameError}</span>
                {/if}
            </div>

            <div class="form-group">
                <label for="email">邮箱</label>
                <input
                    id="email"
                    type="email"
                    bind:value={email}
                    placeholder="example@domain.com"
                    class:error={emailError && email}
                    disabled={isLoading}
                />
                {#if email && emailError}
                    <span class="error-message">❌ {emailError}</span>
                {/if}
            </div>

            <div class="form-group">
                <label for="password">密码</label>
                <input
                    id="password"
                    type="password"
                    bind:value={password}
                    placeholder="请输入密码（6-50 位，包含字母和数字）"
                    class:error={passwordError && password}
                    disabled={isLoading}
                />
                {#if password && passwordError}
                    <span class="error-message">❌ {passwordError}</span>
                {/if}
                {#if password && !passwordError}
                    <span class="success-message">✅ 密码符合要求</span>
                {/if}
            </div>

            <div class="form-group">
                <label for="confirmPassword">确认密码</label>
                <input
                    id="confirmPassword"
                    type="password"
                    bind:value={confirmPassword}
                    placeholder="请再次输入密码"
                    class:error={confirmPasswordError && confirmPassword}
                    disabled={isLoading}
                />
                {#if confirmPassword && confirmPasswordError}
                    <span class="error-message">❌ {confirmPasswordError}</span>
                {/if}
                {#if confirmPassword && !confirmPasswordError && password}
                    <span class="success-message">✅ 密码一致</span>
                {/if}
            </div>

            <button type="submit" class="register-btn" disabled={isLoading || !formValid}>
                {#if isLoading}
                    注册中...
                {:else}
                    立即注册
                {/if}
            </button>

            <div class="register-footer">
                已有账户？
                <a href="#" on:click|preventDefault={() => dispatch('switchToLogin')}>
                    立即登录
                </a>
            </div>
        </form>
    </div>
</div>

<style>
    .register-container {
        display: flex;
        justify-content: center;
        align-items: center;
        min-height: 100vh;
        background: var(--bg-primary);
    }

    .register-card {
        width: 100%;
        max-width: 420px;
        padding: 2rem;
        background: var(--bg-secondary);
        border-radius: 12px;
        box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
    }

    .register-header {
        text-align: center;
        margin-bottom: 2rem;
    }

    .register-header h1 {
        font-size: 1.75rem;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 0.5rem;
    }

    .register-header p {
        font-size: 0.95rem;
        color: var(--text-secondary);
    }

    .register-form {
        display: flex;
        flex-direction: column;
        gap: 1.25rem;
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
        padding: 0.75rem 1rem;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        background: var(--bg-tertiary);
        color: var(--text-primary);
        font-size: 0.95rem;
        transition: all 0.2s;
    }

    .form-group input:focus {
        outline: none;
        border-color: var(--color-primary);
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }

    .form-group input:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .register-btn {
        width: 100%;
        padding: 0.875rem;
        background: var(--color-primary);
        color: white;
        border: none;
        border-radius: 6px;
        font-size: 1rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s;
        margin-top: 0.5rem;
    }

    .register-btn:hover:not(:disabled) {
        background: var(--color-primary-hover);
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
    }

    .register-btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .register-footer {
        text-align: center;
        font-size: 0.9rem;
        color: var(--text-secondary);
        margin-top: 0.5rem;
    }

    .register-footer a {
        color: var(--color-primary);
        text-decoration: none;
        font-weight: 500;
        margin-left: 0.25rem;
    }

    .register-footer a:hover {
        text-decoration: underline;
    }
    
    /* 错误和成功消息样式 */
    .error-message {
        display: block;
        margin-top: 0.5rem;
        font-size: 0.85rem;
        color: var(--color-danger);
        animation: shake 0.3s ease;
    }
    
    .success-message {
        display: block;
        margin-top: 0.5rem;
        font-size: 0.85rem;
        color: var(--color-success);
        animation: fadeIn 0.2s ease;
    }
    
    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        25% { transform: translateX(-4px); }
        75% { transform: translateX(4px); }
    }
    
    @keyframes fadeIn {
        from { opacity: 0; transform: translateY(-4px); }
        to { opacity: 1; transform: translateY(0); }
    }
    
    /* 输入框错误状态 */
    .form-group input.error {
        border-color: var(--color-danger);
        background-color: rgba(239, 68, 68, 0.05);
    }
    
    .form-group input.error:focus {
        box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
    }

    .register-tips {
        margin-top: 1.5rem;
        padding: 1rem;
        background: var(--bg-tertiary);
        border-radius: 6px;
        border-left: 3px solid var(--color-info);
    }

    .register-tips p {
        font-size: 0.85rem;
        color: var(--text-secondary);
        margin: 0;
    }
</style>
