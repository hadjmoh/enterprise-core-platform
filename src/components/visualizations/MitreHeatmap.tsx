import React from 'react';
import { Card } from '../ui/Card';

interface Tactic {
    id: string;
    name: string;
    techniques: string[];
}

const MITRE_TACTICS: Tactic[] = [
    { id: 'TA0001', name: 'Initial Access', techniques: ['Phishing', 'Drive-by', 'Exploit Public App'] },
    { id: 'TA0002', name: 'Execution', techniques: ['PowerShell', 'Cmd', 'Scheduled Task'] },
    { id: 'TA0003', name: 'Persistence', techniques: ['Registry Run Keys', 'New Service', 'Account Manipulation'] },
    { id: 'TA0004', name: 'Privilege Esc', techniques: ['Sudo', 'Token Impersonation', 'Process Injection'] },
    { id: 'TA0005', name: 'Defense Evasion', techniques: ['Masquerading', 'Obfuscated Files', 'Disable Tools'] },
    { id: 'TA0006', name: 'Credential Access', techniques: ['Brute Force', 'OS Dump', 'Keylogging'] },
    { id: 'TA0007', name: 'Discovery', techniques: ['Net Scanning', 'User Discovery', 'File Discovery'] },
    { id: 'TA0008', name: 'Lateral Move', techniques: ['SMB/Windows Admin', 'RDP', 'SSH Hijacking'] },
    { id: 'TA0011', name: 'C2', techniques: ['Web Protocols', 'DNS Tunneling', 'Non-Standard Port'] },
    { id: 'TA0010', name: 'Exfiltration', techniques: ['Auto-exfil', 'Data Transfer', 'Cloud Sync'] }
];

interface MitreHeatmapProps {
    coverage: Record<string, number>; // Technique Name -> Rule Count
}

export const MitreHeatmap: React.FC<MitreHeatmapProps> = ({ coverage }) => {
    return (
        <Card className="p-0 overflow-hidden bg-slate-900 border-slate-700">
            <div className="p-4 border-b border-slate-700 bg-slate-800/50">
                <h3 className="font-bold text-white">MITRE ATT&CK Coverage Map</h3>
                <p className="text-xs text-slate-400">Heatmap of active detection rules per tactic.</p>
            </div>
            <div className="grid grid-cols-5 lg:grid-cols-10 divide-x divide-slate-800 text-xs">
                {MITRE_TACTICS.map(tactic => (
                    <div key={tactic.id} className="flex flex-col min-w-[100px]">
                        <div className="p-2 bg-slate-950 font-bold text-slate-300 text-center border-b border-slate-800 truncate" title={tactic.name}>
                            {tactic.name}
                        </div>
                        <div className="divide-y divide-slate-800">
                            {tactic.techniques.map(tech => {
                                const count = coverage[tech] || 0;
                                let bgClass = 'bg-transparent';
                                if (count > 0) bgClass = 'bg-emerald-900/40 text-emerald-200';
                                if (count > 2) bgClass = 'bg-emerald-600/40 text-emerald-100 font-bold';

                                return (
                                    <div key={tech} className={`p-2 transition-colors hover:bg-slate-800 cursor-help ${bgClass}`} title={`${count} Rules Covering ${tech}`}>
                                        {tech}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                ))}
            </div>
        </Card>
    );
};
