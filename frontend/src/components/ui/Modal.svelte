<script lang="ts">
    export let title: string = '';
    export let showClose: boolean = true;
    export let showFooter: boolean = true;
    export let width: string = '500px';
    
    let visible = true;
    
    function handleClose() {
        visible = false;
        setTimeout(() => {
            dispatch('close');
        }, 300);
    }
    
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();
</script>

{#if visible}
    <div class="modal-backdrop animate-fade-in" on:click={showClose ? handleClose : null}>
        <div 
            class="modal-content" 
            style="width: {width}"
            on:click|stopPropagation
            role="dialog"
            aria-modal="true"
            aria-labelledby={title ? 'modal-title' : undefined}
        >
            {#if title || showClose}
                <div class="modal-header">
                    {#if title}
                        <h3 id="modal-title">{title}</h3>
                    {/if}
                    
                    {#if showClose}
                        <button class="close-btn" on:click={handleClose} aria-label="关闭">
                            ✕
                        </button>
                    {/if}
                </div>
            {/if}
            
            <div class="modal-body">
                <slot></slot>
            </div>
            
            {#if showFooter}
                <div class="modal-footer">
                    <slot name="footer"></slot>
                </div>
            {/if}
        </div>
    </div>
{/if}

<style>
    .modal-backdrop {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: rgba(0, 0, 0, 0.7);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: var(--z-modal-backdrop);
    }
    
    .modal-content {
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: var(--radius-lg);
        box-shadow: var(--shadow-xl);
        max-height: 90vh;
        overflow: hidden;
        display: flex;
        flex-direction: column;
        z-index: var(--z-modal);
    }
    
    .modal-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--spacing-lg);
        border-bottom: 1px solid var(--border-color);
    }
    
    .modal-header h3 {
        margin: 0;
        font-size: var(--font-lg);
        color: var(--text-primary);
    }
    
    .close-btn {
        background: transparent;
        border: none;
        color: var(--text-secondary);
        font-size: var(--font-xl);
        cursor: pointer;
        padding: var(--spacing-sm);
        border-radius: var(--radius-md);
        transition: var(--transition-fast);
    }
    
    .close-btn:hover {
        background-color: var(--bg-tertiary);
        color: var(--text-primary);
    }
    
    .modal-body {
        padding: var(--spacing-lg);
        overflow-y: auto;
        flex: 1;
    }
    
    .modal-footer {
        display: flex;
        justify-content: flex-end;
        gap: var(--spacing-md);
        padding: var(--spacing-lg);
        border-top: 1px solid var(--border-color);
        background-color: var(--bg-secondary);
    }
</style>
