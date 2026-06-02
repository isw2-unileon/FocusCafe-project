import { ReactNode } from 'react';
import { Card } from './ui/card';

interface StatItem {
  icon: ReactNode;
  label: string;
  current: number;
  max?: number;
  barColor: string;
}

interface StatCardProps {
  title: string;
  stats: StatItem[];
  color?: string;
}

export const StatCard = ({ title, stats, color = "bg-white" }: StatCardProps) => {
  return (
    <Card className={`${color} p-7 w-full h-full flex flex-col`}>
      
      <div className="pt-4"> 
        <h2 className="text-lg font-bold text-stone-800 uppercase tracking-widest mb-10">
          {title}
        </h2>
      </div>
      
      <div className="flex flex-col gap-6">
        {stats.map((stat, index) => {
          const maxTarget = stat.max && stat.max > 0 ? stat.max : 100;
          const percentage = Math.min((stat.current / maxTarget) * 100, 100);

          return (
            <div key={index} className="flex items-center gap-4">
              {/* Icon with white background */}
              <div className="p-3 bg-gray-50 rounded-xl shadow-sm border border-black/5">
                {stat.icon}
              </div>
              
              {/* Info and Line */}
              <div className="flex-1">
                <div className="flex justify-between items-end mb-1">
                  <p className="text-[11px] font-bold uppercase text-gray-500">{stat.label}</p>
                  <p className="text-sm font-black text-gray-800">
                    {stat.current} 
                    {stat.max !== undefined && (
                      <span className="text-gray-400 font-medium">/ {stat.max}</span>
                    )}
                  </p>
                </div>
                
                {/* Progress bar*/}
                <div className="w-full h-2.5 bg-gray-100 rounded-full overflow-hidden">
                  <div 
                    className={`h-full ${stat.barColor} transition-all duration-500 rounded-full`} 
                    style={{ width: `${percentage}%` }}
                  />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
};