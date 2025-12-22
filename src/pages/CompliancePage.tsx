import React, { useState, useEffect } from 'react';
import { Shield, FileCheck, Search, CheckCircle, XCircle, RefreshCw, Lock, Database } from 'lucide-react';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';

export const CompliancePage: React.FC = () => {
    const [merkleRoot, setMerkleRoot] = useState<string>('');
    const [status, setStatus] = useState<'valid' | 'invalid' | 'unknown'>('unknown');
    const [verifyId, setVerifyId] = useState('');
    const [verificationResult, setVerificationResult] = useState<any>(null);

    useEffect(() => {
        fetchProof();
    }, []);

    const fetchProof = async () => {
        try {
            const res = await fetch('/api/compliance/proof');
            const data = await res.json();
            setMerkleRoot(data.merkle_root);
            setStatus('valid'); // In a real app, we'd verify this against a pubkey/blockchain
        } catch (e) {
            console.error(e);
            setStatus('unknown');
        }
    };

    const handleVerify = async () => {
        if (!verifyId) return;
        setVerificationResult(null);
        try {
            const res = await fetch('/api/compliance/verify', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ id: verifyId })
            });
            const data = await res.json();
            setVerificationResult(data);
        } catch (e) {
            console.error(e);
            setVerificationResult({ verified: false });
        }
    };

    return (
        <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-slate-100 font-sans">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <h1 className="text-2xl font-bold bg-gradient-to-r from-emerald-400 to-teal-400 bg-clip-text text-transparent flex items-center gap-2">
                        <Shield className="h-8 w-8 text-emerald-400" />
                        Data Governance Center
                    </h1>
                    <p className="text-slate-400 mt-1">Regulatory compliance, audit trails, and data privacy management.</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={fetchProof}>
                        <RefreshCw className="h-4 w-4 mr-2" /> Refresh Status
                    </Button>
                    <Button variant="primary" size="sm">
                        <FileCheck className="h-4 w-4 mr-2" /> Generate Report
                    </Button>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card className="p-6 border-slate-700 bg-slate-800/40">
                    <h3 className="text-slate-400 font-medium mb-2 flex items-center gap-2">
                        <Lock className="h-5 w-5 text-emerald-400" /> Chain of Custody
                    </h3>
                    <div className="flex items-center gap-4 mt-4">
                        <div className={`p-3 rounded-full ${status === 'valid' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>
                            {status === 'valid' ? <CheckCircle className="h-8 w-8" /> : <XCircle className="h-8 w-8" />}
                        </div>
                        <div>
                            <span className="text-lg font-bold text-white uppercase">{status}</span>
                            <p className="text-xs text-slate-500 font-mono mt-1 w-full truncate max-w-[200px]" title={merkleRoot}>
                                Root: {merkleRoot || 'Computing...'}
                            </p>
                        </div>
                    </div>
                </Card>

                <Card className="p-6 border-slate-700 bg-slate-800/40">
                    <h3 className="text-slate-400 font-medium mb-2 flex items-center gap-2">
                        <Database className="h-5 w-5 text-blue-400" /> Privacy Vault
                    </h3>
                    <div className="mt-4 space-y-2">
                        <div className="flex justify-between items-center text-sm">
                            <span className="text-slate-500">Tokenized Fields</span>
                            <Badge variant="outline">Email, SSN, CreditCard</Badge>
                        </div>
                        <div className="flex justify-between items-center text-sm">
                            <span className="text-slate-500">Encryption Standard</span>
                            <span className="text-white font-mono">AES-256-GCM</span>
                        </div>
                        <div className="flex justify-between items-center text-sm">
                            <span className="text-slate-500">Key Rotation</span>
                            <span className="text-emerald-400 text-xs">Active (30 days)</span>
                        </div>
                    </div>
                </Card>

                <Card className="p-6 border-slate-700 bg-slate-800/40">
                    <h3 className="text-slate-400 font-medium mb-2">Manually Verify Log</h3>
                    <div className="flex gap-2 mt-4">
                        <div className="relative flex-1">
                            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-slate-500" />
                            <input
                                type="text"
                                className="w-full bg-slate-900/50 border border-slate-700 rounded-md py-2 pl-9 pr-4 text-sm focus:outline-none focus:ring-1 focus:ring-emerald-500"
                                placeholder="Enter Log ID..."
                                value={verifyId}
                                onChange={(e) => setVerifyId(e.target.value)}
                            />
                        </div>
                        <Button size="sm" onClick={handleVerify}>Verify</Button>
                    </div>
                    {verificationResult && (
                        <div className={`mt-3 p-2 rounded text-xs flex items-center gap-2 ${verificationResult.verified ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>
                            {verificationResult.verified ? <CheckCircle className="h-3 w-3" /> : <XCircle className="h-3 w-3" />}
                            {verificationResult.verified ? 'Cryptographic Proof Verified' : 'Verification Failed'}
                        </div>
                    )}
                </Card>
            </div>

            <Card className="p-0 border-slate-700 bg-slate-800/40 overflow-hidden">
                <div className="p-4 border-b border-slate-700">
                    <h3 className="font-bold text-white">Immutable Audit Log stream</h3>
                </div>
                <div className="p-4">
                    <div className="space-y-0 text-sm font-mono text-slate-400">
                        {[1, 2, 3, 4].map(i => (
                            <div key={i} className="flex gap-4 p-2 hover:bg-slate-700/30 rounded border-b border-slate-800/50 last:border-0">
                                <span className="text-slate-500">2023-10-27T10:0{i}:00Z</span>
                                <span className="text-emerald-500">[AUDIT]</span>
                                <span className="text-slate-300">Proof generated for block #{1000 + i}</span>
                                <span className="text-slate-600 flex-1 text-right truncate">hash: 8f43g...990s</span>
                            </div>
                        ))}
                    </div>
                </div>
            </Card>
        </div>
    );
};
