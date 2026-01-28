'use client';

import { NotificationBell } from '@/components/NotificationBell';

export function HeaderWithNotifications() {
  return (
    <header className="border-b border-slate-200/70 bg-white/70 backdrop-blur">
      <div className="app-container flex items-center justify-end py-3">
        <NotificationBell />
      </div>
    </header>
  );
}
