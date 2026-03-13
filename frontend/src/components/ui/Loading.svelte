<script lang="ts">
    export let size: 'sm' | 'md' | 'lg' = 'md';
    export let text: string = '加载中...';
    export let fullscreen: boolean = false;
    
    function getSizeClass() {
        switch (size) {
            case 'sm': return 'w-8 h-8';
            case 'lg': return 'w-16 h-16';
            default: return 'w-12 h-12';
        }
    }
</script>

<div class="loading-container" class:fullscreen>
    <div class="spinner {getSizeClass()}" role="status" aria-label={text}>
        <span class="spinner-ring"></span>
    </div>
    
    {#if text}
        <p class="loading-text">{text}</p>
    {/if}
</div>

<style>
    .loading-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: var(--spacing-md);
        padding: var(--spacing-2xl);
    }
    
    .loading-container.fullscreen {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(15, 23, 42, 0.9);
        z-index: var(--z-modal);
    }
    
    .spinner {
        display: inline-block;
        position: relative;
    }
    
    .spinner-ring {
        display: block;
        width: 100%;
        height: 100%;
        border-radius: 50%;
        border: 3px solid var(--color-gray-700);
        border-top-color: var(--color-primary);
        animation: spin 1s linear infinite;
    }
    
    .loading-text {
        color: var(--text-secondary);
        font-size: var(--font-sm);
        animation: pulse 2s ease-in-out infinite;
    }
    
    @keyframes spin {
        to { transform: rotate(360deg); }
    }
    
    @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
    }
</style>
