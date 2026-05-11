import { create } from 'zustand';

import { SIDEBAR_CONTENT } from '~types/general';

interface SidebarState {
  isSidebarOpen: boolean;
  content: SIDEBAR_CONTENT | null;
  prevContent: (SIDEBAR_CONTENT | null)[];
  openSidebar: (newContent: SIDEBAR_CONTENT) => void;
  back: () => void;
  closeSidebar: () => void;
  onConfirmEmailCode?: () => void;
  email: string | null;
  setEmail: (email: string | null, onConfirmEmailCode?: () => void) => void;
}

export const useSidebarStore = create<SidebarState>((set) => ({
  isSidebarOpen: false,
  content: null,
  prevContent: [],
  onConfirmEmailCode: undefined,
  openSidebar: (content) =>
    set((prev) => ({
      content,
      isSidebarOpen: true,
      prevContent: [...prev.prevContent, prev.content],
    })),
  back: () =>
    set((prev) => ({
      content: prev.prevContent.pop()?.length ? prev.prevContent.pop() : null,
      isSidebarOpen: !!prev.prevContent.pop()?.length,
    })),
  email: null,
  setEmail: (email, onConfirmEmailCode) => set({ email, onConfirmEmailCode }),
  closeSidebar: () => {
    set({
      isSidebarOpen: false,
      onConfirmEmailCode: undefined,
      content: null,
      email: null,
    });
  },
}));
