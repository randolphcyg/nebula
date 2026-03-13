import { writable } from 'svelte/store';

interface AppState {
    loading: boolean;
    error: string | null;
    sidebarOpen: boolean;
    activeTab: string;
}

function createAppStore() {
    const { subscribe, set, update } = writable<AppState>({
        loading: false,
        error: null,
        sidebarOpen: true,
        activeTab: 'home'
    });

    return {
        subscribe,
        
        setLoading: (loading: boolean) => {
            update(state => ({ ...state, loading }));
        },
        
        setError: (error: string | null) => {
            update(state => ({ ...state, error }));
        },
        
        toggleSidebar: () => {
            update(state => ({ ...state, sidebarOpen: !state.sidebarOpen }));
        },
        
        setActiveTab: (tab: string) => {
            update(state => ({ ...state, activeTab: tab }));
        },
        
        reset: () => {
            set({
                loading: false,
                error: null,
                sidebarOpen: true,
                activeTab: 'home'
            });
        }
    };
}

export const app = createAppStore();
