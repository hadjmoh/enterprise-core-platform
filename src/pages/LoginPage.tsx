import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { Input } from '../components/ui/Input';
import { Lock, Mail, ArrowRight, ShieldCheck } from 'lucide-react';

export function LoginPage() {
    const { login, isLoading } = useAuth();
    const navigate = useNavigate();
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        await login();
        navigate('/');
    };

    return (
        <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 relative overflow-hidden">
            {/* Background Effects */}
            <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[600px] bg-brand-500/20 rounded-full blur-[120px] opacity-30 animate-pulse" />
            <div className="absolute bottom-0 right-0 w-[800px] h-[600px] bg-accent-500/10 rounded-full blur-[100px] opacity-20" />

            <Card variant="glass" className="w-full max-w-md p-8 border-slate-700/50 relative z-10 backdrop-blur-xl">
                <div className="text-center mb-8">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-brand-500 to-cyan-400 p-[1px] mb-6 shadow-glow">
                        <div className="w-full h-full bg-slate-900 rounded-2xl flex items-center justify-center">
                            <ShieldCheck className="h-8 w-8 text-brand-400" />
                        </div>
                    </div>
                    <div className="flex justify-center mb-4">
                        <Badge variant="outline" className="border-brand-500/30 text-brand-300">
                            Simulated Environment
                        </Badge>
                    </div>
                    <h1 className="text-3xl font-bold text-white mb-2 tracking-tight">Welcome Back</h1>
                    <p className="text-slate-400">Sign in to access your Enterprise Dashboard</p>
                </div>

                <form onSubmit={handleLogin} className="space-y-6">
                    <div className="space-y-4">
                        <div className="space-y-2">
                            <label htmlFor="email" className="text-sm font-medium text-slate-300 ml-1">Email Address</label>
                            <Input
                                id="email"
                                placeholder="admin@enterprise.core"
                                icon={<Mail className="h-4 w-4" />}
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <div className="flex justify-between items-center ml-1">
                                <label htmlFor="password" className="text-sm font-medium text-slate-300">Password</label>
                                <button
                                    type="button"
                                    onClick={() => alert('Password reset flow not implemented in mock.')}
                                    className="text-xs text-brand-400 hover:text-brand-300 transition-colors focus:outline-none focus:underline"
                                >
                                    Forgot password?
                                </button>
                            </div>
                            <Input
                                id="password"
                                type="password"
                                placeholder="••••••••"
                                icon={<Lock className="h-4 w-4" />}
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                required
                            />
                        </div>
                    </div>

                    <Button
                        className="w-full h-11 text-base shadow-lg shadow-brand-500/25"
                        size="lg"
                        type="submit"
                        isLoading={isLoading}
                    >
                        Sign In <ArrowRight className="ml-2 h-4 w-4" />
                    </Button>
                </form>

                <div className="mt-8 text-center text-sm text-slate-500">
                    <p>Protected by Enterprise Grade Security</p>
                    <p className="text-xs mt-1 opacity-60">v2.4.0-RC1build.894</p>
                </div>
            </Card>
        </div>
    );
}
