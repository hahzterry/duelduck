import { create } from 'zustand';

export enum BOTTOM_SHEET_STATE {}

interface State {
  content: BOTTOM_SHEET_STATE | null;
}

interface Actions {
  setContent: (content: BOTTOM_SHEET_STATE | null) => void;
  close: () => void;
}

const initialState: State = {
  content: null,
};

export const useBottomSheetStore = create<State & Actions>((set) => ({
  ...initialState,
  setContent: (content) =>
    set({
      content,
      ...(content === null
        ? {
            leaderboardCalendarDate: null,
            claimRewardData: null,
            dataReportUser: null,
            reportsData: null,
          }
        : {}),
    }),
  close: () => set(initialState),
}));
