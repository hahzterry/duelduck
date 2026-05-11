import { CSSProperties, ReactNode } from 'react';
import { create } from 'zustand';

export interface ErrorToastTextPart {
  text: string;
  color?: string;
  fontSize?: string;
  href?: string;
  hoverColor?: string;
  onClick?: () => void;
}

export interface ErrorToastData {
  titleParts?: ErrorToastTextPart[];
  textParts?: ErrorToastTextPart[];
  backgroundColor?: string;
  textColor?: string;
  stylesContainer?: CSSProperties;
  stylesToast?: CSSProperties;
  accentTextColor?: string;
  additionalContent?: ReactNode;
  progressType?: 'line' | 'cross-circle';
  positionClose?: 'outside-bottom' | 'inside-right';
  progressColor?: string;
  timeClose?: number;
}

interface State {
  dataToast: ErrorToastData;
  isShow: boolean;
}

interface Actions {
  open: (dataToast?: ErrorToastData) => void;
  close: () => void;
  resetMsg: () => void;
}

const initialState: State = {
  dataToast: {},
  isShow: false,
};

export const useErrorToastStore = create<State & Actions>((set) => ({
  ...initialState,
  close: () => {
    set({
      isShow: false,
    });
  },
  resetMsg: () => {
    set({
      dataToast: initialState.dataToast,
    });
  },
  open: (dataToast = initialState.dataToast) =>
    set({ dataToast, isShow: true }),
}));
