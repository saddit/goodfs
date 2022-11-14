import '@vue/runtime-core'

declare module '@vue/runtime-core' {
    interface ComponentCustomProperties {
        $utils: typeof pkg.utils;
        $cst: typeof pkg.cst
    }
}

export {}  // Important! See note.
