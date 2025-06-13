import React from 'react';
import { AlertTriangle, Users, Network, Clock } from 'lucide-react';

const StatsCards = ({ stats }) => {
  const cards = [
    {
      title: 'Total Alerts',
      value: stats.totalAlerts,
      icon: AlertTriangle,
      color: 'text-blue-500',
      bgColor: 'bg-blue-50'
    },
    {
      title: 'High Severity',
      value: stats.highSeverity,
      icon: AlertTriangle,
      color: 'text-red-500',
      bgColor: 'bg-red-50'
    },
    {
      title: 'Correlations Found',
      value: stats.correlationsFound,
      icon: Network,
      color: 'text-green-500',
      bgColor: 'bg-green-50'
    },
    {
      title: 'Avg Processing Time',
      value: `${stats.avgProcessingTime}ms`,
      icon: Clock,
      color: 'text-yellow-500',
      bgColor: 'bg-yellow-50'
    }
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {cards.map((card, index) => {
        const Icon = card.icon;
        return (
          <div key={index} className="card">
            <div className="flex items-center">
              <div className={`p-3 rounded-lg ${card.bgColor}`}>
                <Icon className={`h-6 w-6 ${card.color}`} />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">{card.title}</p>
                <p className="text-2xl font-bold text-gray-900">{card.value}</p>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default StatsCards; 