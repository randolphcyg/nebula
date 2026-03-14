import { app } from '../stores/app';
import { error as toastError, success as toastSuccess, warning as toastWarning, info as toastInfo } from '../stores/toast';
import { logger } from './logger';

/**
 * 统一的错误处理函数
 */
export function handleError(error: any, customMessage?: string) {
    logger.error('Error:', error);
    
    const message = customMessage || (typeof error === 'string' ? error : '操作失败，请稍后重试');
    app.setError(message);
    
    // 使用 Toast 显示错误提示
    toastError(message, 5000);
    
    return message;
}

/**
 * 显示成功提示
 */
export function showSuccess(message: string, duration?: number) {
    toastSuccess(message, duration);
}

/**
 * 显示警告提示
 */
export function showWarning(message: string, duration?: number) {
    toastWarning(message, duration);
}

/**
 * 显示信息提示
 */
export function showInfo(message: string, duration?: number) {
    toastInfo(message, duration);
}

/**
 * 清除错误状态
 */
export function clearError() {
    app.setError(null);
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * 格式化日期时间
 */
export function formatDateTime(dateString: string): string {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

/**
 * 格式化相对时间
 */
export function formatRelativeTime(dateString: string): string {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);
    
    if (diffSecs < 60) return '刚刚';
    if (diffMins < 60) return `${diffMins} 分钟前`;
    if (diffHours < 24) return `${diffHours} 小时前`;
    if (diffDays < 7) return `${diffDays} 天前`;
    
    return formatDateTime(dateString);
}

/**
 * 防抖函数
 */
export function debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number
): (...args: Parameters<T>) => void {
    let timeout: ReturnType<typeof setTimeout> | null = null;
    
    return function executedFunction(...args: Parameters<T>) {
        const later = () => {
            timeout = null;
            func(...args);
        };
        
        if (timeout) {
            clearTimeout(timeout);
        }
        timeout = setTimeout(later, wait);
    };
}

/**
 * 节流函数
 */
export function throttle<T extends (...args: any[]) => any>(
    func: T,
    limit: number
): (...args: Parameters<T>) => void {
    let inThrottle: boolean;
    
    return function(...args: Parameters<T>) {
        if (!inThrottle) {
            func(...args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

/**
 * 验证邮箱格式
 */
export function isValidEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

/**
 * 复制到剪贴板
 */
export async function copyToClipboard(text: string): Promise<boolean> {
    if (!text) return false;
    
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch (err) {
        logger.error('无法复制:', err);
        return false;
    }
}

/**
 * 下载文件
 */
export function downloadFile(content: string, filename: string, mimeType: string = 'text/plain') {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}

/**
 * 解析 JWT Token（简单解析，不验证签名）
 * ⚠️ 仅用于解码 Token 中的非敏感信息（如用户名、过期时间）
 * 不可用于权限验证！
 */
export function parseJwt(token: string): any {
    try {
        const base64Url = token.split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const jsonPayload = decodeURIComponent(
            atob(base64)
                .split('')
                .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
                .join('')
        );
        return JSON.parse(jsonPayload);
    } catch (e) {
        logger.error('解析 Token 失败:', e);
        return null;
    }
}

/**
 * 检查 Token 是否过期
 */
export function isTokenExpired(token: string): boolean {
    const payload = parseJwt(token);
    if (!payload || !payload.exp) return true;
    
    const expiryDate = new Date(payload.exp * 1000);
    return expiryDate <= new Date();
}
