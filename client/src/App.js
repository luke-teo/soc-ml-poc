import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { 
  Shield, 
  AlertTriangle, 
  Users, 
  Network, 
  Clock, 
  TrendingUp,
  RefreshCw,
  Eye
} from 'lucide-react';
import AlertsList from './components/AlertsList';
import AnalysisDetails from './components/AnalysisDetails';
import StatsCards from './components/StatsCards';
import CorrelationChart from './components/CorrelationChart';

function App() {
  const [alerts, setAlerts] = useState([]);
  const [selectedAlert, setSelectedAlert] = useState(null);
  const [analysisResult, setAnalysisResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({
    totalAlerts: 0,
    highSeverity: 0,
    correlationsFound: 0,
    avgProcessingTime: 0
  });

  // Simulate real-time alerts (in production, this would be WebSocket or SSE)
  useEffect(() => {
    const interval = setInterval(() => {
      // Generate mock alerts for demo
      const mockAlert = {
        id: Date.now().toString(),
        timestamp: new Date().toISOString(),
        source: ['aws_waf', 'azure_waf', 'deep_security', 'akamai_waf'][Math.floor(Math.random() * 4)],
        severity: ['high', 'medium', 'low'][Math.floor(Math.random() * 3)],
        message: [
          'Suspicious SQL injection attempt detected',
          'Cross-site scripting attempt blocked',
          'File modification detected in system directory',
          'Multiple failed login attempts'
        ][Math.floor(Math.random() * 4)],
        project_id: 'demo-project-1',
        status: 'analyzing'
      };

      setAlerts(prev => [mockAlert, ...prev.slice(0, 19)]); // Keep last 20 alerts
      
      // Update stats
      setStats(prev => ({
        ...prev,
        totalAlerts: prev.totalAlerts + 1,
        highSeverity: prev.highSeverity + (mockAlert.severity === 'high' ? 1 : 0)
      }));

      // Simulate analysis completion after 2-5 seconds
      setTimeout(() => {
        setAlerts(prev => 
          prev.map(alert => 
            alert.id === mockAlert.id 
              ? { ...alert, status: 'completed' }
              : alert
          )
        );
      }, Math.random() * 3000 + 2000);

    }, 10000); // New alert every 10 seconds

    return () => clearInterval(interval);
  }, []);

  const handleViewAnalysis = async (alertId) => {
    setLoading(true);
    setSelectedAlert(alertId);
    
    try {
      // In a real implementation, this would fetch from /analysis/{alert_id}
      // For demo, we'll generate mock analysis data
      const mockAnalysis = {
        alert_id: alertId,
        project_id: 'demo-project-1',
        correlated_logs: [
          {
            source: 'aws_waf',
            timestamp: new Date().toISOString(),
            ip_addresses: ['192.168.1.100', '10.0.0.50'],
            user_emails: ['user@example.com', 'admin@company.com'],
            action: 'BLOCK',
            severity: 'high'
          },
          {
            source: 'deep_security',
            timestamp: new Date().toISOString(),
            ip_addresses: ['192.168.1.100'],
            user_emails: [],
            action: 'ALERT',
            severity: 'medium'
          }
        ],
        user_correlations: [
          {
            user_identifier: 'user@example.com',
            ip_address: '192.168.1.100',
            confidence_score: 0.85,
            correlation_type: 'time_proximity',
            source_systems: ['aws_waf', 'deep_security']
          },
          {
            user_identifier: 'admin@company.com',
            ip_address: '10.0.0.50',
            confidence_score: 0.72,
            correlation_type: 'direct',
            source_systems: ['aws_waf']
          }
        ],
        enrichment_data: {
          correlation_stats: {
            total_logs_analyzed: 15,
            user_correlations_found: 2,
            correlation_score: 0.78
          },
          involved_users: ['user@example.com', 'admin@company.com'],
          involved_ips: ['192.168.1.100', '10.0.0.50'],
          source_breakdown: {
            'aws_waf': 8,
            'deep_security': 4,
            'azure_waf': 3
          }
        },
        processing_time_ms: Math.floor(Math.random() * 200) + 50
      };

      setAnalysisResult(mockAnalysis);
      
      // Update stats
      setStats(prev => ({
        ...prev,
        correlationsFound: prev.correlationsFound + mockAnalysis.user_correlations.length,
        avgProcessingTime: Math.round((prev.avgProcessingTime + mockAnalysis.processing_time_ms) / 2)
      }));

    } catch (error) {
      console.error('Failed to fetch analysis:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    window.location.reload();
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
                         <div className="flex items-center">
               <Shield className="h-8 w-8 text-blue-500 mr-3" />
               <div>
                 <h1 className="text-2xl font-bold text-gray-900">SOC ML Dashboard</h1>
                 <p className="text-sm text-gray-500">Log Correlation & Analysis System</p>
               </div>
             </div>
                         <button 
               type="button"
               onClick={handleRefresh}
               className="btn btn-secondary flex items-center"
             >
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh
            </button>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats Cards */}
        <StatsCards stats={stats} />

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mt-8">
          {/* Alerts List */}
          <div className="space-y-6">
            <AlertsList 
              alerts={alerts} 
              onViewAnalysis={handleViewAnalysis}
              selectedAlert={selectedAlert}
            />
          </div>

          {/* Analysis Details */}
          <div className="space-y-6">
            {analysisResult ? (
              <>
                <AnalysisDetails 
                  analysis={analysisResult} 
                  loading={loading}
                />
                <CorrelationChart 
                  correlations={analysisResult.user_correlations}
                  sourceBreakdown={analysisResult.enrichment_data.source_breakdown}
                />
              </>
            ) : (
              <div className="card text-center py-12">
                <Eye className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 mb-2">
                  Select an Alert
                </h3>
                <p className="text-gray-500">
                  Click "View Analysis" on any alert to see correlation results
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App; 