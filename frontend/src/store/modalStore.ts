import { create } from 'zustand';

export enum MODAL_CONTENT {
  LOGIN = 'LOGIN',
}

interface ModalState {
  content: MODAL_CONTENT | null;
  isOpen: boolean;
  openModal: (content?: MODAL_CONTENT) => void;
  handleClose: (() => void) | null;
  setHandleClose: (handleClose: (() => void) | null) => void;
  closeModal: () => void;
}

const useModalStore = create<ModalState>((set, get) => ({
  isOpen: false,
  content: null,
  handleClose: null,
  setHandleClose: (handleClose) => set({ handleClose }),
  openModal: (content?: MODAL_CONTENT) =>
    set({ isOpen: true, ...(content ? { content } : {}) }),
  closeModal: () => {
    const { handleClose } = get();

    handleClose?.();

    set({
      handleClose: null,
      isOpen: false,
      content: null,
    });
  },
}));

export default useModalStore;
