import { useEffect } from "react";

let lockCount = 0;
let savedOverflow = "";
let savedPaddingRight = "";

const lockBodyScroll = () => {
  const { body, documentElement } = document;

  if (lockCount === 0) {
    savedOverflow = body.style.overflow;
    savedPaddingRight = body.style.paddingRight;

    const scrollbarWidth = window.innerWidth - documentElement.clientWidth;
    body.style.overflow = "hidden";
    if (scrollbarWidth > 0) {
      body.style.paddingRight = `${scrollbarWidth}px`;
    }
  }

  lockCount += 1;
};

const unlockBodyScroll = () => {
  lockCount = Math.max(0, lockCount - 1);
  if (lockCount === 0) {
    document.body.style.overflow = savedOverflow;
    document.body.style.paddingRight = savedPaddingRight;
  }
};

export const useBodyScrollLock = (locked: boolean) => {
  useEffect(() => {
    if (!locked) return;

    lockBodyScroll();
    return () => unlockBodyScroll();
  }, [locked]);
};
