import '@vue/runtime-core'

declare module '@vue/runtime-core' {
    interface ComponentCustomProperties {
        $utils: typeof pkg.utils;
    }
}

export {}  // Important! See note.
