import { ReactNode } from 'react';
import { Card } from './ui/card';

interface StatCardProps {
  icon: ReactNode;
  label: string;
  value?: string;
  color: string;
}

export const StatCard = ({icon, label, value, color }: StatCardProps) => {
    return (
        <Card className={`${color} p-4 rounded-2xl flex flex-col items-center justify-center gap-2 shadow-sm border border-black/5`}>
            <div className="p-3 bg-white rounded-xl shadow-inner">
                {icon}
            </div>
            <div>
                <p className="text-[12px] font-bold uppercase text-gray-500 tracking-tighter">{label}</p>
                <p className="text-xl font-black text-gray-800">{value}</p>
            </div>
        </Card>
    );
};