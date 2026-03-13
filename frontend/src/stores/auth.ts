import { writable } from 'svelte/store';
import type { User } from '../types';

function createAuthStore() {
    const storedToken = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
    const storedUser = typeof localStorage !== 'undefined' ? localStorage.getItem('user') : null;

    const { subscribe, set, update } = writable({
        isAuthenticated: !!storedToken,
        user: storedUser ? JSON.parse(storedUser) : null,
        token: storedToken
    });

    return {
        subscribe,
        
        login: (token: string, user: User) => {
            if (typeof localStorage !== 'undefined') {
                localStorage.setItem('token', token);
                localStorage.setItem('user', JSON.stringify(user));
            }
            set({
                isAuthenticated: true,
                user,
                token
            });
        },
        
        logout: () => {
            if (typeof localStorage !== 'undefined') {
                localStorage.removeItem('token');
                localStorage.removeItem('user');
            }
            set({
                isAuthenticated: false,
                user: null,
                token: null
            });
        },
        
        updateUser: (user: User) => {
            update(state => {
                if (typeof localStorage !== 'undefined') {
                    localStorage.setItem('user', JSON.stringify(user));
                }
                return { ...state, user };
            });
        },
        
        check: (): boolean => {
            let isAuthenticated = false;
            subscribe(state => {
                isAuthenticated = state.isAuthenticated;
            })();
            return isAuthenticated;
        },
        
        getCurrentUser: (): User | null => {
            let currentUser: User | null = null;
            subscribe(state => {
                currentUser = state.user;
            })();
            return currentUser;
        }
    };
}

export const auth = createAuthStore();
