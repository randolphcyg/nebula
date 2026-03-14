/**
 * 日志级别枚举
 */
export enum LogLevel {
    DEBUG = 0,
    INFO = 1,
    WARN = 2,
    ERROR = 3,
    NONE = 4
}

/**
 * 根据环境设置日志级别
 * 生产环境：只显示错误
 * 开发环境：显示所有日志
 */
const getLogLevel = (): LogLevel => {
    if (import.meta.env.PROD) {
        return LogLevel.ERROR;
    }
    return import.meta.env.DEV ? LogLevel.DEBUG : LogLevel.INFO;
};

const CURRENT_LOG_LEVEL = getLogLevel();

/**
 * 日志工具类
 * 
 * 使用示例:
 * logger.debug('调试信息', data);
 * logger.info('普通信息');
 * logger.warn('警告信息');
 * logger.error('错误信息', error);
 */
export const logger = {
    /**
     * 调试日志（仅开发环境）
     */
    debug(...args: any[]): void {
        if (CURRENT_LOG_LEVEL <= LogLevel.DEBUG) {
            console.log('[DEBUG]', ...args);
        }
    },

    /**
     * 信息日志
     */
    info(...args: any[]): void {
        if (CURRENT_LOG_LEVEL <= LogLevel.INFO) {
            console.log('[INFO]', ...args);
        }
    },

    /**
     * 警告日志
     */
    warn(...args: any[]): void {
        if (CURRENT_LOG_LEVEL <= LogLevel.WARN) {
            console.warn('[WARN]', ...args);
        }
    },

    /**
     * 错误日志（始终显示）
     */
    error(...args: any[]): void {
        if (CURRENT_LOG_LEVEL <= LogLevel.ERROR) {
            console.error('[ERROR]', ...args);
        }
    },

    /**
     * 成功日志（仅开发环境）
     */
    success(...args: any[]): void {
        if (CURRENT_LOG_LEVEL <= LogLevel.DEBUG) {
            console.log('[SUCCESS]', ...args);
        }
    }
};
