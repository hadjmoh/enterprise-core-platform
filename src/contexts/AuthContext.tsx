import React, { useState, useEffect } from 'react';
import type { User } from '../types/auth';
import { AuthContext } from './AuthContextInstance';

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        // Simulate checking session on mount
        const timer = setTimeout(() => {
            setIsLoading(false);
        }, 1000);
        return () => clearTimeout(timer);
    }, []);

    const login = async () => {
        setIsLoading(true);
        // Simulate API call
        return new Promise<void>((resolve) => {
            setTimeout(() => {
                setUser({
                    id: 'u-123',
                    name: 'Admin User',
                    email: 'admin@enterprise.core',
                    role: 'admin'
                });
                setIsLoading(false);
                resolve();
            }, 1500);
        });
    };

    const logout = () => {
        setUser(null);
    };

    return (
        <AuthContext.Provider value={{
            user,
            isLoading,
            login,
            logout,
            isAuthenticated: !!user
        }}>
            {children}
        </AuthContext.Provider>
    );
}
