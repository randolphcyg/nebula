<script lang="ts">
    export let variant: 'primary' | 'success' | 'danger' | 'outline' = 'primary';
    export let size: 'sm' | 'md' | 'lg' = 'md';
    export let disabled: boolean = false;
    export let loading: boolean = false;
    export let type: 'button' | 'submit' | 'reset' = 'button';
    
    function getVariantClass() {
        switch (variant) {
            case 'success': return 'btn-success';
            case 'danger': return 'btn-danger';
            case 'outline': return 'btn-outline';
            default: return 'btn-primary';
        }
    }
    
    function getSizeClass() {
        switch (size) {
            case 'sm': return 'text-xs px-3 py-1.5';
            case 'lg': return 'text-lg px-6 py-3';
            default: return 'text-sm px-4 py-2';
        }
    }
</script>

<button 
    type={type}
    class="btn {getVariantClass()} {getSizeClass()}"
    disabled={disabled || loading}
    on:click
>
    {#if loading}
        <span class="animate-spin mr-2">⟳</span>
    {/if}
    
    <slot></slot>
</button>

<style>
    .btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        font-weight: 500;
        border: none;
        border-radius: var(--radius-md);
        cursor: pointer;
        transition: var(--transition-base);
        white-space: nowrap;
    }
    
    .btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
    
    .btn-primary {
        background-color: var(--color-primary);
        color: var(--color-white);
    }
    
    .btn-primary:hover:not(:disabled) {
        background-color: var(--color-primary-hover);
    }
    
    .btn-success {
        background-color: var(--color-success);
        color: var(--color-white);
    }
    
    .btn-success:hover:not(:disabled) {
        background-color: var(--color-success-hover);
    }
    
    .btn-danger {
        background-color: var(--color-danger);
        color: var(--color-white);
    }
    
    .btn-danger:hover:not(:disabled) {
        background-color: var(--color-danger-hover);
    }
    
    .btn-outline {
        background-color: transparent;
        border: 1px solid var(--border-color-light);
        color: var(--text-primary);
    }
    
    .btn-outline:hover:not(:disabled) {
        background-color: var(--bg-tertiary);
    }
</style>
