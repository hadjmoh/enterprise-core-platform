import React, { useState, useEffect } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { cn } from '../../lib/utils';
import { AlertSidebar } from '../alerts/AlertSidebar';
import { AlertToastContainer } from '../alerts/AlertToast';

export function DashboardLayout({ children }: { children: React.ReactNode }) {
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const [isMobile, setIsMobile] = useState(false);
    const [alertSidebarOpen, setAlertSidebarOpen] = useState(false);

    useEffect(() => {
        const checkMobile = () => {
            setIsMobile(window.innerWidth < 1024);
            if (window.innerWidth < 1024) {
                setSidebarOpen(false);
            } else {
                setSidebarOpen(true);
            }
        };
        checkMobile();
        window.addEventListener('resize', checkMobile);
        return () => window.removeEventListener('resize', checkMobile);
    }, []);

    return (
        <div className="min-h-screen bg-slate-950 text-slate-200">
            <Sidebar
                isOpen={sidebarOpen}
                setIsOpen={setSidebarOpen}
                isMobile={isMobile}
            />

            <div
                className={cn(
                    "transition-all duration-300 flex flex-col min-h-screen",
                    sidebarOpen && !isMobile ? "ml-64" : isMobile ? "ml-0" : "ml-20"
                )}
            >
                <Header
                    toggleSidebar={() => setSidebarOpen(!sidebarOpen)}
                    openAlerts={() => setAlertSidebarOpen(true)}
                />

                <main className="flex-1 p-6 overflow-y-auto">
                    <div className="mx-auto max-w-7xl animate-in fade-in duration-500">
                        {children}
                    </div>
                </main>
            </div>

            <AlertSidebar isOpen={alertSidebarOpen} onClose={() => setAlertSidebarOpen(false)} />
            <AlertToastContainer />

            {/* Mobile Overlay */}
            {isMobile && sidebarOpen && (
                <div
                    className="fixed inset-0 z-30 bg-slate-950/80 backdrop-blur-sm"
                    onClick={() => setSidebarOpen(false)}
                    role="presentation"
                    aria-hidden="true"
                />
            )}
        </div>
    );
}
