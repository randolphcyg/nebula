<script lang="ts">
    import { subscribe, remove } from '../../stores/toast';
    import Toast from './Toast.svelte';
    
    let toasts: Array<{ id: number; options: any }> = [];
    
    subscribe((value) => {
        toasts = value;
    });
</script>

<div class="toast-container">
    {#each toasts as toast}
        <Toast
            message={toast.options.message}
            type={toast.options.type}
            duration={toast.options.duration}
            on:close={() => remove(toast.id)}
        />
    {/each}
</div>

<style>
    .toast-container {
        position: fixed;
        top: var(--spacing-lg);
        right: var(--spacing-lg);
        display: flex;
        flex-direction: column;
        gap: var(--spacing-md);
        z-index: var(--z-toast);
        pointer-events: none;
    }
    
    .toast-container > * {
        pointer-events: auto;
    }
</style>
