import React from 'react';
import { Eye, Clock, AlertTriangle, CheckCircle, Loader } from 'lucide-react';

const AlertsList = ({ alerts, onViewAnalysis, selectedAlert }) => {
  const getSeverityBadge = (severity) => {
    const badges = {
      high: 'badge badge-high',
      medium: 'badge badge-medium',
      low: 'badge badge-low'
    };
    return badges[severity] || 'badge';
  };

  const getSourceIcon = (source) => {
    const icons = {
      aws_waf: '🛡️',
      azure_waf: '🔷',
      deep_security: '🔒',
      akamai_waf: '☁️'
    };
    return icons[source] || '🔍';
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'analyzing':
        return <Loader className="h-4 w-4 text-warning-500 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-success-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-400" />;
    }
  };

  const formatTime = (timestamp) => {
    return new Date(timestamp).toLocaleTimeString();
  };

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-lg font-semibold text-gray-900">Recent Alerts</h2>
        <div className="flex items-center text-sm text-gray-500">
          <div className="w-2 h-2 bg-green-400 rounded-full mr-2 animate-pulse"></div>
          Live
        </div>
      </div>

      <div className="space-y-4 max-h-96 overflow-y-auto">
        {alerts.length === 0 ? (
          <div className="text-center py-8">
            <AlertTriangle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-500">No alerts yet. Waiting for incoming alerts...</p>
          </div>
        ) : (
          alerts.map((alert) => (
            <div
              key={alert.id}
                               className={`border rounded-lg p-4 transition-all duration-200 ${
                   selectedAlert === alert.id
                     ? 'border-blue-500 bg-blue-50'
                     : 'border-gray-200 hover:border-gray-300'
                 }`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center mb-2">
                    <span className="text-lg mr-2">{getSourceIcon(alert.source)}</span>
                    <span className="font-medium text-gray-900 capitalize">
                      {alert.source.replace('_', ' ')}
                    </span>
                    <span className={getSeverityBadge(alert.severity)} style={{ marginLeft: '8px' }}>
                      {alert.severity}
                    </span>
                  </div>
                  
                  <p className="text-sm text-gray-600 mb-2">{alert.message}</p>
                  
                  <div className="flex items-center text-xs text-gray-500">
                    <Clock className="h-3 w-3 mr-1" />
                    {formatTime(alert.timestamp)}
                    <span className="mx-2">•</span>
                    {getStatusIcon(alert.status)}
                    <span className="ml-1 capitalize">{alert.status}</span>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={() => onViewAnalysis(alert.id)}
                  disabled={alert.status === 'analyzing'}
                                     className={`ml-4 btn btn-primary flex items-center text-sm ${
                     alert.status === 'analyzing' 
                       ? 'opacity-50 cursor-not-allowed' 
                       : 'hover:bg-blue-600'
                   }`}
                >
                  <Eye className="h-4 w-4 mr-1" />
                  View Analysis
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default AlertsList; 