import { writable } from 'svelte/store';

interface ToastOptions {
    message: string;
    type?: 'success' | 'error' | 'warning' | 'info';
    duration?: number;
}

interface ToastInstance {
    id: number;
    options: ToastOptions;
}

const { subscribe, update } = writable<ToastInstance[]>([]);

let idCounter = 0;

function show(options: ToastOptions | string) {
    const opts = typeof options === 'string' 
        ? { message: options, type: 'info' as const }
        : options;
    
    const id = idCounter++;
    
    update(toasts => [...toasts, { id, options: opts }]);
    
    // 自动移除
    if (opts.duration !== 0) {
        setTimeout(() => {
            remove(id);
        }, opts.duration || 3000);
    }
    
    return id;
}

function remove(id: number) {
    update(toasts => toasts.filter(t => t.id !== id));
}

function clear() {
    update(() => []);
}

// 便捷方法
export function success(message: string, duration?: number) {
    return show({ message, type: 'success', duration });
}

export function error(message: string, duration?: number) {
    return show({ message, type: 'error', duration });
}

export function warning(message: string, duration?: number) {
    return show({ message, type: 'warning', duration });
}

export function info(message: string, duration?: number) {
    return show({ message, type: 'info', duration });
}

export { subscribe, show, remove, clear };
