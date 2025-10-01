import { useRef, useCallback, useState } from "react";

export function useClickActions(
  onSingleClick: () => void,
  onDoubleClick: () => void,
  delay: number = 300 // ms threshold between clicks
) {
  const timerRef = useRef<number | null>(null);
  const clickCountRef = useRef(0);

  const handleClick = useCallback(
    (e: React.SyntheticEvent) => {
      e.preventDefault();

      clickCountRef.current += 1;

      if (clickCountRef.current === 1) {
        // Start timer for single click
        timerRef.current = window.setTimeout(() => {
          if (clickCountRef.current === 1) {
            onSingleClick();
          }
          clickCountRef.current = 0;
          timerRef.current = null;
        }, delay);
      } else if (clickCountRef.current === 2) {
        // Double click detected → cancel timer
        if (timerRef.current) {
          clearTimeout(timerRef.current);
          timerRef.current = null;
        }
        onDoubleClick();
        clickCountRef.current = 0;
      }
    },
    [onSingleClick, onDoubleClick, delay]
  );

  return { handleClick };
}

export function useControlActive() {
  const [active, setActive] = useState(false);

  const activate = useCallback(() => {
    setActive(true);
  }, []);

  const deactivate = useCallback(() => {
    setActive(false);
  }, []);

  return {
    active,
    activate,
    deactivate,
  };
}
