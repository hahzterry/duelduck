import { useMediaQueryStore } from '~store/mediaStore';
import useModalStore, { MODAL_CONTENT } from '~store/modalStore';
import { useSidebarStore } from '~store/sidebarStore';
import { SIDEBAR_CONTENT } from '~types/general';

export const openLoginSelect = () => {
  const isSmallTablet = useMediaQueryStore.getState().isSmallTablet;

  if (isSmallTablet) {
    useModalStore.getState().openModal(MODAL_CONTENT.LOGIN);

    return;
  }

  useSidebarStore.getState().openSidebar(SIDEBAR_CONTENT.SELECT_TYPE_SIGN_IN);
};
