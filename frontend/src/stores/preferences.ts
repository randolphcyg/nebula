import { writable } from 'svelte/store';

export interface UserPreferences {
    pageSize: number;
    autoRefresh: boolean;
    refreshInterval: number;
    showNotifications: boolean;
    compactMode: boolean;
}

const defaultPreferences: UserPreferences = {
    pageSize: 10,
    autoRefresh: false,
    refreshInterval: 30,
    showNotifications: true,
    compactMode: false
};

function loadPreferences(): UserPreferences {
    if (typeof window === 'undefined') return defaultPreferences;
    
    const saved = localStorage.getItem('preferences');
    if (saved) {
        try {
            return { ...defaultPreferences, ...JSON.parse(saved) };
        } catch (e) {
            return defaultPreferences;
        }
    }
    
    return defaultPreferences;
}

function savePreferences(prefs: UserPreferences) {
    if (typeof window === 'undefined') return;
    localStorage.setItem('preferences', JSON.stringify(prefs));
}

function createPreferencesStore() {
    const preferences = loadPreferences();
    
    const { subscribe, update } = writable<UserPreferences>(preferences);
    
    return {
        subscribe,
        
        setPageSize(size: number) {
            update(current => {
                const newPrefs = { ...current, pageSize: size };
                savePreferences(newPrefs);
                return newPrefs;
            });
        },
        
        toggleAutoRefresh() {
            update(current => {
                const newPrefs = { ...current, autoRefresh: !current.autoRefresh };
                savePreferences(newPrefs);
                return newPrefs;
            });
        },
        
        setRefreshInterval(interval: number) {
            update(current => {
                const newPrefs = { ...current, refreshInterval: interval };
                savePreferences(newPrefs);
                return newPrefs;
            });
        },
        
        toggleNotifications() {
            update(current => {
                const newPrefs = { ...current, showNotifications: !current.showNotifications };
                savePreferences(newPrefs);
                return newPrefs;
            });
        },
        
        toggleCompactMode() {
            update(current => {
                const newPrefs = { ...current, compactMode: !current.compactMode };
                savePreferences(newPrefs);
                return newPrefs;
            });
        },
        
        reset() {
            update(current => {
                savePreferences(defaultPreferences);
                return defaultPreferences;
            });
        }
    };
}

export const preferencesStore = createPreferencesStore();
