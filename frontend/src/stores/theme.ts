import { writable } from 'svelte/store';

export interface ThemeConfig {
    mode: 'light' | 'dark';
    primaryColor: string;
}

const defaultTheme: ThemeConfig = {
    mode: 'light',
    primaryColor: '#3b82f6'
};

// 从 localStorage 加载主题
function loadTheme(): ThemeConfig {
    if (typeof window === 'undefined') return defaultTheme;
    
    const saved = localStorage.getItem('theme');
    if (saved) {
        try {
            return JSON.parse(saved);
        } catch (e) {
            return defaultTheme;
        }
    }
    
    // 根据系统偏好设置
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        return { ...defaultTheme, mode: 'dark' };
    }
    
    return defaultTheme;
}

// 保存主题到 localStorage
function saveTheme(theme: ThemeConfig) {
    if (typeof window === 'undefined') return;
    localStorage.setItem('theme', JSON.stringify(theme));
}

// 应用主题到 document
function applyTheme(theme: ThemeConfig) {
    if (typeof document === 'undefined') return;
    
    const root = document.documentElement;
    
    // 应用暗色模式
    if (theme.mode === 'dark') {
        root.setAttribute('data-theme', 'dark');
        root.classList.add('dark');
    } else {
        root.removeAttribute('data-theme');
        root.classList.remove('dark');
    }
    
    // 应用主题色
    root.style.setProperty('--color-primary', theme.primaryColor);
}

// 创建主题 store
function createThemeStore() {
    const theme = loadTheme();
    applyTheme(theme);
    
    const { subscribe, update, set } = writable<ThemeConfig>(theme);
    
    return {
        subscribe,
        
        // 切换亮暗模式
        toggleMode() {
            update(current => {
                const newMode: 'light' | 'dark' = current.mode === 'light' ? 'dark' : 'light';
                const newTheme: ThemeConfig = {
                    ...current,
                    mode: newMode
                };
                saveTheme(newTheme);
                applyTheme(newTheme);
                return newTheme;
            });
        },
        
        // 设置模式
        setMode(mode: 'light' | 'dark') {
            update(current => {
                const newTheme: ThemeConfig = { ...current, mode };
                saveTheme(newTheme);
                applyTheme(newTheme);
                return newTheme;
            });
        },
        
        // 设置主题色
        setPrimaryColor(color: string) {
            update(current => {
                const newTheme = { ...current, primaryColor: color };
                saveTheme(newTheme);
                applyTheme(newTheme);
                return newTheme;
            });
        },
        
        // 重置为主题
        reset() {
            set(defaultTheme);
            saveTheme(defaultTheme);
            applyTheme(defaultTheme);
        }
    };
}

export const themeStore = createThemeStore();

// 初始化主题
if (typeof window !== 'undefined') {
    applyTheme(loadTheme());
    
    // 监听系统主题变化
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        const current = loadTheme();
        if (!localStorage.getItem('theme')) {
            const newMode: 'light' | 'dark' = e.matches ? 'dark' : 'light';
            const newTheme: ThemeConfig = {
                ...current,
                mode: newMode
            };
            applyTheme(newTheme);
        }
    });
}
