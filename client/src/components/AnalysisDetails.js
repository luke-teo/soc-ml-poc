import React from 'react';
import { Users, Network, Clock, TrendingUp, Shield } from 'lucide-react';

const AnalysisDetails = ({ analysis, loading }) => {
  if (loading) {
    return (
      <div className="card">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="space-y-3">
            <div className="h-3 bg-gray-200 rounded"></div>
            <div className="h-3 bg-gray-200 rounded w-5/6"></div>
            <div className="h-3 bg-gray-200 rounded w-4/6"></div>
          </div>
        </div>
      </div>
    );
  }

  const getConfidenceColor = (score) => {
    if (score >= 0.8) return 'text-green-600 bg-green-50';
    if (score >= 0.6) return 'text-yellow-600 bg-yellow-50';
    return 'text-red-600 bg-red-50';
  };

  const getCorrelationTypeIcon = (type) => {
    switch (type) {
      case 'direct':
        return '🎯';
      case 'time_proximity':
        return '⏰';
      case 'historical':
        return '📊';
      default:
        return '🔗';
    }
  };

  return (
    <div className="space-y-6">
      {/* Analysis Summary */}
      <div className="card">
        <div className="flex items-center mb-4">
          <TrendingUp className="h-5 w-5 text-blue-500 mr-2" />
          <h3 className="text-lg font-semibold text-gray-900">Analysis Summary</h3>
        </div>
        
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-gray-500">Logs Analyzed</p>
            <p className="text-xl font-bold text-gray-900">
              {analysis.enrichment_data.correlation_stats.total_logs_analyzed}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Processing Time</p>
            <p className="text-xl font-bold text-gray-900">
              {analysis.processing_time_ms}ms
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Correlation Score</p>
            <p className="text-xl font-bold text-gray-900">
              {(analysis.enrichment_data.correlation_stats.correlation_score * 100).toFixed(1)}%
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Correlations Found</p>
            <p className="text-xl font-bold text-gray-900">
              {analysis.user_correlations.length}
            </p>
          </div>
        </div>
      </div>

      {/* User Correlations */}
      <div className="card">
        <div className="flex items-center mb-4">
          <Users className="h-5 w-5 text-blue-500 mr-2" />
          <h3 className="text-lg font-semibold text-gray-900">User Correlations</h3>
        </div>
        
        <div className="space-y-3">
          {analysis.user_correlations.map((correlation, index) => (
            <div key={`${correlation.user_identifier}-${correlation.ip_address}`} className="border border-gray-200 rounded-lg p-4">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center mb-2">
                    <span className="text-lg mr-2">
                      {getCorrelationTypeIcon(correlation.correlation_type)}
                    </span>
                    <span className="font-medium text-gray-900">
                      {correlation.user_identifier}
                    </span>
                    <span className="mx-2 text-gray-400">↔</span>
                    <span className="font-mono text-sm text-gray-600">
                      {correlation.ip_address}
                    </span>
                  </div>
                  
                  <div className="flex items-center space-x-4 text-sm text-gray-500">
                    <span>Type: {correlation.correlation_type.replace('_', ' ')}</span>
                    <span>Sources: {correlation.source_systems.join(', ')}</span>
                  </div>
                </div>
                
                <div className={`px-3 py-1 rounded-full text-sm font-medium ${getConfidenceColor(correlation.confidence_score)}`}>
                  {(correlation.confidence_score * 100).toFixed(0)}% confidence
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Involved Entities */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="card">
          <div className="flex items-center mb-4">
            <Users className="h-5 w-5 text-blue-500 mr-2" />
            <h3 className="text-lg font-semibold text-gray-900">Involved Users</h3>
          </div>
          <div className="space-y-2">
            {analysis.enrichment_data.involved_users.map((user) => (
              <div key={user} className="flex items-center p-2 bg-gray-50 rounded">
                <span className="text-sm font-mono">{user}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="card">
          <div className="flex items-center mb-4">
            <Network className="h-5 w-5 text-blue-500 mr-2" />
            <h3 className="text-lg font-semibold text-gray-900">Involved IPs</h3>
          </div>
          <div className="space-y-2">
            {analysis.enrichment_data.involved_ips.map((ip) => (
              <div key={ip} className="flex items-center p-2 bg-gray-50 rounded">
                <span className="text-sm font-mono">{ip}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default AnalysisDetails; 