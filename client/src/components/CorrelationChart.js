import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

const CorrelationChart = ({ correlations, sourceBreakdown }) => {
  // Prepare data for confidence score distribution
  const confidenceData = correlations.reduce((acc, correlation) => {
    const range = correlation.confidence_score >= 0.8 ? 'High (80%+)' :
                  correlation.confidence_score >= 0.6 ? 'Medium (60-80%)' : 'Low (<60%)';
    acc[range] = (acc[range] || 0) + 1;
    return acc;
  }, {});

  const confidenceChartData = Object.entries(confidenceData).map(([range, count]) => ({
    range,
    count
  }));

  // Prepare data for source breakdown
  const sourceChartData = Object.entries(sourceBreakdown).map(([source, count]) => ({
    source: source.replace('_', ' ').toUpperCase(),
    count
  }));

  const COLORS = ['#3b82f6', '#ef4444', '#f59e0b', '#22c55e'];

  return (
    <div className="space-y-6">
      {/* Confidence Score Distribution */}
      <div className="card">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Confidence Score Distribution</h3>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={confidenceChartData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="range" />
            <YAxis />
            <Tooltip />
            <Bar dataKey="count" fill="#3b82f6" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Source System Breakdown */}
      <div className="card">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Log Sources Breakdown</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <ResponsiveContainer width="100%" height={200}>
            <PieChart>
              <Pie
                data={sourceChartData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ source, percent }) => `${source} ${(percent * 100).toFixed(0)}%`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="count"
              >
                {sourceChartData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>

          <div className="space-y-3">
            {sourceChartData.map((item, index) => (
              <div key={item.source} className="flex items-center justify-between">
                <div className="flex items-center">
                  <div 
                    className="w-4 h-4 rounded mr-3"
                    style={{ backgroundColor: COLORS[index % COLORS.length] }}
                  />
                  <span className="text-sm font-medium">{item.source}</span>
                </div>
                <span className="text-sm text-gray-500">{item.count} logs</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Correlation Types */}
      <div className="card">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Correlation Types</h3>
        <div className="space-y-3">
          {correlations.map((correlation, index) => (
            <div key={`correlation-${index}`} className="flex items-center justify-between p-3 bg-gray-50 rounded">
              <div className="flex items-center">
                <span className="text-lg mr-3">
                  {correlation.correlation_type === 'direct' ? '🎯' :
                   correlation.correlation_type === 'time_proximity' ? '⏰' : '📊'}
                </span>
                <div>
                  <p className="font-medium text-sm">{correlation.correlation_type.replace('_', ' ')}</p>
                  <p className="text-xs text-gray-500">
                    {correlation.user_identifier} ↔ {correlation.ip_address}
                  </p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-sm font-medium">
                  {(correlation.confidence_score * 100).toFixed(0)}%
                </p>
                <p className="text-xs text-gray-500">confidence</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default CorrelationChart; 