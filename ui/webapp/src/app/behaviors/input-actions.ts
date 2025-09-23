import { useState, useRef, useCallback } from "react";

type State = 'primary' | 'secondary' | 'none';

export function useInputActions(
  onActionPrimary: () => void,
  onActionSecondary: () => void,
  threshold: number = 300,
) {
  const [state, setState] = useState<State>('none');
  const timerRef = useRef<number | null>(null);

  const startPress = useCallback(
    (e: React.SyntheticEvent) => {
      e.preventDefault();

      setState('primary');

      timerRef.current = window.setTimeout(() => {
        setState('secondary');
      }, threshold);
    },
    [onActionSecondary, threshold]
  );

  const endPress = useCallback(
    (e: React.SyntheticEvent) => {
      e.preventDefault();

      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }

      switch (state) {
      case 'primary':
        onActionPrimary();
        break;
      case 'secondary':
        onActionSecondary();
        break;
      }

      setState('none');
    },
    [state, onActionPrimary, onActionSecondary]
  );

  const cancelPress = useCallback((e: React.SyntheticEvent) => {
    e.preventDefault();

    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }

    setState('none');
  }, []);

  return {
    state,
    startPress,
    endPress,
    cancelPress,
  };
}
