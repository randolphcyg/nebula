import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import ts from 'typescript-eslint';

export default ts.config(
    js.configs.recommended,
    ...ts.configs.recommended,
    ...svelte.configs['flat/recommended'],
    {
        languageOptions: {
            globals: {
                console: 'readonly',
                window: 'readonly',
                document: 'readonly',
                navigator: 'readonly',
                localStorage: 'readonly',
                fetch: 'readonly',
                CustomEvent: 'readonly',
                Event: 'readonly',
                setTimeout: 'readonly',
                clearTimeout: 'readonly',
                setInterval: 'readonly',
                clearInterval: 'readonly',
                atob: 'readonly',
                btoa: 'readonly',
                navigator: 'readonly',
                Alert: 'readonly',
                Confirm: 'readonly',
                OpenURL: 'readonly',
                SelectFile: 'readonly',
                SelectFiles: 'readonly',
                SelectDirectory: 'readonly',
                SaveFile: 'readonly',
            },
        },
        rules: {
            '@typescript-eslint/no-explicit-any': 'warn',
            '@typescript-eslint/no-unused-vars': ['warn', { 
                'argsIgnorePattern': '^_',
                'varsIgnorePattern': '^_'
            }],
            '@typescript-eslint/explicit-function-return-type': 'off',
            '@typescript-eslint/no-non-null-assertion': 'warn',
            'no-console': 'off', // 允许 console，但建议使用 logger
            'eqeqeq': ['error', 'always'],
            'curly': ['error', 'all'],
            'prefer-const': 'warn',
            'no-var': 'error',
        },
    },
    {
        files: ['**/*.svelte'],
        languageOptions: {
            parserOptions: {
                parser: ts.parser,
            },
        },
        rules: {
            'no-undef': 'off',
            '@typescript-eslint/no-unused-vars': 'off',
        },
    },
    {
        ignores: [
            '**/node_modules/',
            '**/dist/',
            '**/build/',
            '**/wailsjs/',
            '**/*.d.ts',
        ],
    }
);
