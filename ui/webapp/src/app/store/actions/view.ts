import { createAction } from '@reduxjs/toolkit';
import { AppThunkAction } from '../types';
import { getView, hasView, isViewPopup } from '../selectors/view';
import { hasWindow, getDefaultWindowId } from '../selectors/model';
import { isMobile } from '../../utils/detect-browser';
import { navigate } from './navigation';

export const internalViewClose = createAction('view/close');
export const viewPopup = createAction<string>('view/popup');
export const internalViewChange = createAction<string>('view/change');

const navigateToDefault = (): AppThunkAction => (dispatch, getState) => {
  const state = getState();

  const defaultWindowId = getDefaultWindowId(state, isMobile ? 'mobile' : 'desktop');
  console.log(`using default window: ${defaultWindowId}`); // eslint-disable-line no-console

  dispatch(viewChange(defaultWindowId));
};

export const viewNavigationChange = (windowId: string): AppThunkAction => (dispatch, getState) => {
  const state = getState();

  // ensure that the window exists
  if (!hasWindow(state, windowId)) {
    dispatch(navigateToDefault());
    return;
  }

  console.log(`showing window: ${windowId}`); // eslint-disable-line no-console
  dispatch(internalViewChange(windowId));
};

export const viewInit = (): AppThunkAction => (dispatch, getState) => {
  const state = getState();

  // ensure that the window exists
  const windowId = getView(state)[0];
  if (!windowId || !hasWindow(state, windowId)) {
    dispatch(navigateToDefault());
  }
};

export const viewChange = (windowId: string) => {
  return navigate(windowId);
};

export const viewClose = (): AppThunkAction => (dispatch, getState) => {
  const state = getState();
  if (!isViewPopup(state)) {
    console.error('Cannot close root window!'); // eslint-disable-line no-console
    return;
  }

  dispatch(internalViewClose());
};
