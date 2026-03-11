<script lang="ts">
    export let data: any;
    export let keyName: string = "";
    export let isLast: boolean = true;
    export let defaultExpanded: boolean = true;

    // 当前节点的折叠状态
    let expanded = defaultExpanded;

    // 类型推断
    $: isNull = data === null;
    $: type = isNull ? 'null' : typeof data;
    $: isArray = Array.isArray(data);
    $: isObject = type === 'object' && !isNull;
    $: isComplex = isObject || isArray;
    $: children = isComplex ? Object.keys(data) : [];

    function toggle() {
        if (isComplex) expanded = !expanded;
    }
</script>

<div class="json-node">
    <div class="line" class:clickable={isComplex} on:click={toggle} role="button">
        {#if isComplex}
            <span class="toggle">{expanded ? '▼' : '▶'}</span>
        {:else}
            <span class="toggle indent"></span>
        {/if}

        {#if keyName !== ""}
            <span class="key">"{keyName}"</span><span class="colon">: </span>
        {/if}

        {#if isComplex}
            <span class="bracket">{isArray ? '[' : '{'}</span>
            {#if !expanded}
                <span class="summary"> {isArray ? `... ${children.length} items ...` : `... ${children.length} keys ...`} </span>
                <span class="bracket">{isArray ? ']' : '}'}
                    {#if !isLast}<span class="comma">,</span>{/if}</span>
            {/if}
        {:else}
            <span class="value {type}">
                {#if type === 'string'}
                    "{data}"
                {:else}
                    {String(data)}
                {/if}
            </span>
            {#if !isLast}<span class="comma">,</span>{/if}
        {/if}
    </div>

    {#if isComplex && expanded}
        <div class="children">
            {#each children as childKey, i}
                <svelte:self
                        data={data[childKey]}
                        keyName={isArray ? "" : childKey}
                        isLast={i === children.length - 1}
                        defaultExpanded={false}
                />
            {/each}
        </div>
        <div class="line indent-bracket">
            <span class="bracket">{isArray ? ']' : '}'}</span>
            {#if !isLast}<span class="comma">,</span>{/if}
        </div>
    {/if}
</div>

<style>
    .json-node {
        font-family: 'Fira Code', Consolas, monospace;
        font-size: 0.85rem;
        line-height: 1.6;
    }

    .line {
        display: flex;
        align-items: flex-start;
        border-radius: 4px;
        transition: background 0.1s;
    }

    .line.clickable {
        cursor: pointer;
        user-select: none;
    }

    .line.clickable:hover {
        background: #1e293b;
    }

    .toggle {
        display: inline-block;
        width: 18px;
        color: #64748b;
        font-size: 0.65rem;
        text-align: center;
        margin-top: 3px;
    }

    .indent {
        width: 18px;
    }

    /* 缩进与虚线辅助线 */
    .children {
        margin-left: 8px;
        border-left: 1px dashed #334155;
        padding-left: 8px;
    }

    .indent-bracket {
        margin-left: 18px;
    }

    /* 语法高亮配色 (完美适配暗黑主题) */
    .key {
        color: #38bdf8;
    }

    /* 键名：天蓝 */
    .colon, .comma, .bracket {
        color: #94a3b8;
    }

    .summary {
        color: #475569;
        font-style: italic;
        margin: 0 4px;
    }

    .string {
        color: #10b981;
    }

    /* 字符串：翠绿 */
    .number {
        color: #f59e0b;
    }

    /* 数字：橙黄 */
    .boolean {
        color: #a855f7;
        font-weight: bold;
    }

    /* 布尔：紫色 */
    .null {
        color: #ef4444;
        font-style: italic;
    }

    /* Null：红色 */
</style>