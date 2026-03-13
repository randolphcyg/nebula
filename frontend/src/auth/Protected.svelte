<script lang="ts">
    import { auth } from '../../stores/auth';
    import { onMount } from 'svelte';
    
    export let showLogin: boolean = false;
    
    let isAuthenticated = false;
    
    onMount(() => {
        const unsubscribe = auth.subscribe(state => {
            isAuthenticated = state.isAuthenticated;
        });
        
        return unsubscribe;
    });
</script>

{#if !isAuthenticated}
    {#if showLogin}
        <slot></slot>
    {:else}
        <div class="protected-fallback">
            <p>请先登录</p>
        </div>
    {/if}
{:else}
    <slot></slot>
{/if}

<style>
    .protected-fallback {
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        color: var(--text-secondary);
        font-size: var(--font-lg);
    }
</style>
