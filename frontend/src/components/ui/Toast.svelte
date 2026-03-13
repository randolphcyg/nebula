<script lang="ts">
    export let message: string = '';
    export let type: 'success' | 'error' | 'warning' | 'info' = 'info';
    export let duration: number = 3000;
    
    let visible = false;
    let timeoutId: ReturnType<typeof setTimeout>;
    
    import { onMount, onDestroy } from 'svelte';
    
    onMount(() => {
        visible = true;
        
        if (duration > 0) {
            timeoutId = setTimeout(() => {
                close();
            }, duration);
        }
    });
    
    onDestroy(() => {
        if (timeoutId) {
            clearTimeout(timeoutId);
        }
    });
    
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();
    
    function close() {
        visible = false;
        setTimeout(() => {
            dispatch('close');
        }, 300);
    }
    
    function getIcon() {
        switch (type) {
            case 'success': return '✓';
            case 'error': return '✕';
            case 'warning': return '⚠';
            case 'info': return 'ℹ';
            default: return 'ℹ';
        }
    }
</script>

<div class="toast {type} {visible ? 'visible' : ''}" role="alert">
    <span class="toast-icon">{getIcon()}</span>
    <span class="toast-message">{message}</span>
    <button class="toast-close" on:click={close}>×</button>
</div>

<style>
    .toast {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        padding: var(--spacing-md) var(--spacing-lg);
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-lg);
        box-shadow: var(--shadow-lg);
        min-width: 300px;
        max-width: 500px;
        transform: translateX(calc(100% + var(--spacing-lg)));
        opacity: 0;
        transition: all var(--transition-slow);
    }
    
    .toast.visible {
        transform: translateX(0);
        opacity: 1;
    }
    
    .toast.success {
        border-left: 4px solid var(--color-success);
    }
    
    .toast.error {
        border-left: 4px solid var(--color-danger);
    }
    
    .toast.warning {
        border-left: 4px solid var(--color-warning);
    }
    
    .toast.info {
        border-left: 4px solid var(--color-info);
    }
    
    .toast-icon {
        font-size: var(--font-lg);
        flex-shrink: 0;
    }
    
    .toast.success .toast-icon { color: var(--color-success); }
    .toast.error .toast-icon { color: var(--color-danger); }
    .toast.warning .toast-icon { color: var(--color-warning); }
    .toast.info .toast-icon { color: var(--color-info); }
    
    .toast-message {
        flex: 1;
        color: var(--text-primary);
        font-size: var(--font-sm);
        line-height: 1.5;
    }
    
    .toast-close {
        background: transparent;
        border: none;
        color: var(--text-secondary);
        font-size: var(--font-lg);
        cursor: pointer;
        padding: var(--spacing-xs);
        border-radius: var(--radius-md);
        transition: var(--transition-fast);
        flex-shrink: 0;
    }
    
    .toast-close:hover {
        background-color: var(--bg-tertiary);
        color: var(--text-primary);
    }
</style>
