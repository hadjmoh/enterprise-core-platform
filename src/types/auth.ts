export interface User {
    id: string;
    name: string;
    email: string;
    role: 'admin' | 'user' | 'readonly';
}

export interface AuthContextType {
    user: User | null;
    isLoading: boolean;
    login: () => Promise<void>;
    logout: () => void;
    isAuthenticated: boolean;
}
