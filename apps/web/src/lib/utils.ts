// Basic utility types for Skeleton UI
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
