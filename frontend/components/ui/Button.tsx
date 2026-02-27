'use client';

import { HTMLMotionProps, LazyMotion, domAnimation, m } from 'framer-motion';
import { RefreshCcw } from 'lucide-react';
import React, { forwardRef } from 'react';

interface ButtonProps extends Omit<HTMLMotionProps<'button'>, 'children'> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  icon?: React.ReactNode;
  iconPosition?: 'left' | 'right';
  success?: boolean;
  children?: React.ReactNode;
}

const CheckIcon = ({ animate }: { animate: boolean }) => (
  <LazyMotion features={domAnimation}>
    <m.svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <m.path
        d="M5 12l5 5L19 7"
        initial={{ pathLength: 0 }}
        animate={{ pathLength: animate ? 1 : 0 }}
        transition={{ duration: 0.3, ease: 'easeOut' }}
      />
    </m.svg>
  </LazyMotion>
);

const variantStyles = {
  primary:
    'bg-gradient-to-r from-accent-orange to-orange-500 text-white hover:shadow-lg hover:shadow-accent-orange/25',
  secondary:
    'bg-white/5 border border-white/10 text-white hover:bg-white/10',
  ghost: 'bg-transparent text-white hover:bg-white/5',
  danger:
    'bg-red-500/10 border border-red-500/30 text-red-400 hover:bg-red-500/20',
};

const sizeStyles = {
  sm: 'px-3 py-1.5 text-[10px] gap-1.5',
  md: 'px-4 py-2 text-xs gap-2',
  lg: 'px-6 py-3 text-sm gap-2.5',
};

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = 'primary',
      size = 'md',
      loading = false,
      icon,
      iconPosition = 'left',
      success = false,
      disabled,
      className = '',
      children,
      ...props
    },
    ref
  ) => {
    const isDisabled = disabled || loading;

    const baseStyles = `
      inline-flex items-center justify-center
      font-bold uppercase tracking-wider
      rounded-lg
      transition-[transform,box-shadow,background-color,border-color] duration-200
      focus:outline-none focus-visible:ring-2 focus-visible:ring-accent-orange focus-visible:ring-offset-2 focus-visible:ring-offset-black
      ${isDisabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
    `;

    const renderIcon = () => {
      if (loading) {
        return <RefreshCcw className="w-4 h-4 animate-spin" />;
      }
      if (success) {
        return <CheckIcon animate={success} />;
      }
      return icon;
    };

    const iconElement = renderIcon();

    return (
      <LazyMotion features={domAnimation}>
        <m.button
          ref={ref}
          disabled={isDisabled}
          className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
          whileHover={isDisabled ? undefined : { scale: 1.02 }}
          whileTap={isDisabled ? undefined : { scale: 0.97 }}
          transition={{ duration: 0.2 }}
          {...props}
        >
          {iconElement && iconPosition === 'left' && iconElement}
          {children}
          {iconElement && iconPosition === 'right' && iconElement}
        </m.button>
      </LazyMotion>
    );
  }
);

Button.displayName = 'Button';

export { Button, type ButtonProps };
