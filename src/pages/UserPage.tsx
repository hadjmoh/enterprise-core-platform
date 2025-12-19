import { useState } from 'react';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import {
    Search,
    Plus,
    Filter,
    MoreVertical,
    Shield,
    Mail,
    Calendar
} from 'lucide-react';

// Mock Data
const MOCK_USERS = [
    { id: 'u-1', name: 'Alina Starkov', email: 'alina@enterprise.core', role: 'Admin', status: 'Active', joined: '2024-01-12' },
    { id: 'u-2', name: 'Kaz Brekker', email: 'kaz@enterprise.core', role: 'User', status: 'Active', joined: '2024-02-05' },
    { id: 'u-3', name: 'Jesper Fahey', email: 'jesper@enterprise.core', role: 'User', status: 'Inactive', joined: '2024-02-08' },
    { id: 'u-4', name: 'Nina Zenik', email: 'nina@enterprise.core', role: 'Manager', status: 'Active', joined: '2024-03-10' },
    { id: 'u-5', name: 'Matthias Helvar', email: 'matthias@enterprise.core', role: 'User', status: 'Active', joined: '2024-03-15' },
    { id: 'u-6', name: 'Inej Ghafa', email: 'inej@enterprise.core', role: 'Admin', status: 'Active', joined: '2024-01-20' },
    { id: 'u-7', name: 'Wylan Van Eck', email: 'wylan@enterprise.core', role: 'Editor', status: 'Pending', joined: '2024-04-01' },
];

export function UserPage() {
    const [searchTerm, setSearchTerm] = useState('');
    const [roleFilter, setRoleFilter] = useState<string | null>(null);

    // Filter Logic
    const filteredUsers = MOCK_USERS.filter(user => {
        const matchesSearch = user.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            user.email.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesRole = roleFilter ? user.role === roleFilter : true;
        return matchesSearch && matchesRole;
    });

    return (
        <div className="flex flex-col gap-6">
            {/* Header Controls */}
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-white tracking-tight">User Directory</h1>
                    <p className="text-slate-400">Manage user access and permissions across the platform.</p>
                </div>
                <Button variant="primary" size="md">
                    <Plus className="h-4 w-4 mr-2" />
                    Add User
                </Button>
            </div>

            {/* Filters Bar */}
            <Card className="flex flex-col md:flex-row gap-4 p-4">
                <div className="flex-1">
                    <Input
                        placeholder="Search users by name or email..."
                        icon={<Search className="h-4 w-4" />}
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
                <div className="flex gap-3 overflow-x-auto pb-2 md:pb-0">
                    <Button
                        variant={roleFilter === null ? 'primary' : 'outline'}
                        size="sm"
                        onClick={() => setRoleFilter(null)}
                    >
                        All
                    </Button>
                    <Button
                        variant={roleFilter === 'Admin' ? 'primary' : 'outline'}
                        size="sm"
                        onClick={() => setRoleFilter(roleFilter === 'Admin' ? null : 'Admin')}
                    >
                        Admins
                    </Button>
                    <Button
                        variant={roleFilter === 'User' ? 'primary' : 'outline'}
                        size="sm"
                        onClick={() => setRoleFilter(roleFilter === 'User' ? null : 'User')}
                    >
                        Users
                    </Button>
                    <Button variant="outline" size="sm">
                        <Filter className="h-4 w-4 mr-2" />
                        More Filters
                    </Button>
                </div>
            </Card>

            {/* Data Table */}
            <Card className="overflow-hidden p-0">
                <div className="overflow-x-auto">
                    <table className="w-full text-left">
                        <thead>
                            <tr className="bg-slate-900 border-b border-slate-800 text-slate-400 text-sm">
                                <th className="p-4 font-medium">User Profile</th>
                                <th className="p-4 font-medium">Role</th>
                                <th className="p-4 font-medium">Status</th>
                                <th className="p-4 font-medium">Date Joined</th>
                                <th className="p-4 font-medium text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800/50">
                            {filteredUsers.length > 0 ? (
                                filteredUsers.map((user) => (
                                    <tr key={user.id} className="group hover:bg-slate-800/30 transition-colors">
                                        <td className="p-4">
                                            <div className="flex items-center gap-3">
                                                <div className="h-10 w-10 rounded-full bg-gradient-to-tr from-slate-700 to-slate-600 flex items-center justify-center text-white font-bold">
                                                    {user.name.charAt(0)}
                                                </div>
                                                <div>
                                                    <div className="font-medium text-white">{user.name}</div>
                                                    <div className="text-sm text-slate-500 flex items-center gap-1">
                                                        <Mail className="h-3 w-3" /> {user.email}
                                                    </div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="p-4">
                                            <div className="flex items-center gap-2 text-slate-300">
                                                <Shield className="h-3 w-3 text-brand-400" />
                                                <span>{user.role}</span>
                                            </div>
                                        </td>
                                        <td className="p-4">
                                            <Badge
                                                variant={
                                                    user.status === 'Active' ? 'success' :
                                                        user.status === 'Inactive' ? 'error' : 'warning'
                                                }
                                            >
                                                {user.status}
                                            </Badge>
                                        </td>
                                        <td className="p-4 text-slate-400 text-sm">
                                            <div className="flex items-center gap-2">
                                                <Calendar className="h-3 w-3" />
                                                {user.joined}
                                            </div>
                                        </td>
                                        <td className="p-4 text-right">
                                            <Button variant="ghost" size="sm">
                                                <MoreVertical className="h-4 w-4" />
                                            </Button>
                                        </td>
                                    </tr>
                                ))
                            ) : (
                                <tr>
                                    <td colSpan={5} className="p-8 text-center text-slate-500">
                                        No users found matching your search.
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
                <div className="p-4 border-t border-slate-800 bg-slate-900/50 flex justify-between items-center text-sm text-slate-400">
                    <span>Showing {filteredUsers.length} users</span>
                    <div className="flex gap-2">
                        <Button variant="outline" size="sm" disabled>Previous</Button>
                        <Button variant="outline" size="sm" disabled>Next</Button>
                    </div>
                </div>
            </Card>
        </div>
    );
}
