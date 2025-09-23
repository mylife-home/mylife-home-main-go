import React, { FunctionComponent, useMemo } from 'react';
import clsx from 'clsx';
import { useDispatch, useSelector } from 'react-redux';
import { AppState, AppThunkDispatch } from '../store/types';
import { UIControl, makeGetUIControl } from '../store/selectors/control';
import { actionPrimary, actionSecondary } from '../store/actions/actions';
import { useInputActions } from '../behaviors/input-actions';

type ControlProps = {
  windowId: string;
  controlId: string;
};

const Control: FunctionComponent<ControlProps> = ({ windowId, controlId }) => {
  const { control, onActionPrimary, onActionSecondary } = useConnect(windowId, controlId);
  const { state, startPress, endPress, cancelPress } = useInputActions(onActionPrimary, onActionSecondary);
  const active = state !== 'none';

  return (
    <>
      {active && (
        <div className={clsx('mylife-control-overlay', state === 'primary' && 'primary', state === 'secondary' && 'secondary')} />
      )}

      <div
        style={getStyleSizePosition(control)}
        className={clsx(control.active ? 'mylife-control-button' : 'mylife-control-inactive', active && 'active', ...control.style)}
        onTouchStart={startPress}
        onTouchEnd={endPress}
        onTouchCancel={cancelPress}
        onMouseDown={startPress}
        onMouseUp={endPress}
        onMouseLeave={cancelPress}
      >
        {control.displayResource && <img src={`/resources/${control.displayResource}`} />}
        {control.text && <p>{control.text}</p>}
      </div>
    </>
  )
};

export default Control;

function getStyleSizePosition(control: UIControl) {
  const { left, top, height, width } = control;
  return { left, top, height, width };
}

function useConnect(windowId: string, controlId: string) {
  const dispatch = useDispatch<AppThunkDispatch>();
  const getUIControl = useMemo(() => makeGetUIControl(windowId, controlId), [windowId, controlId]);
  return {
    control: useSelector((state: AppState) => getUIControl(state)),
    ...useMemo(() => ({
      onActionPrimary: () => dispatch(actionPrimary(windowId, controlId)),
      onActionSecondary: () => dispatch(actionSecondary(windowId, controlId))
    }), [dispatch, windowId, controlId])
  };
};
