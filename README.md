# SOC ML - Log Correlation and Analysis System

A comprehensive system for correlating logs across multiple security systems and performing automated analysis when alerts are triggered, featuring a real-time web dashboard.

## 🏗️ Project Structure

```
soc-ml/
├── server/                 # Go backend API
│   ├── *.go               # Go source files
│   ├── go.mod             # Go dependencies
│   ├── docker-compose.yml # Infrastructure services
│   └── loki-config.yaml   # Loki configuration
├── client/                # React web dashboard
│   ├── src/               # React source files
│   ├── public/            # Static assets
│   └── package.json       # Node.js dependencies
└── README.md              # This file
```

## 🚀 Quick Start

### 1. Start Infrastructure Services

```bash
cd server
docker-compose up -d
```

This starts:
- **Loki** (port 3100) - Log storage and querying
- **PostgreSQL** (port 5432) - Analysis results storage
- **Redis** (port 6379) - Task queue
- **Grafana** (port 3000) - Optional log visualization

### 2. Start the Backend Server

```bash
cd server
go mod tidy
go run *.go
```

The server will:
- Start HTTP API on port 8080
- Connect to PostgreSQL and create tables
- Begin generating mock alerts every 30 seconds
- Process alerts through the correlation engine

### 3. Start the Web Dashboard

```bash
cd client
npm install
npm start
```

The React dashboard will:
- Start on port 3000
- Connect to the backend API
- Display real-time alerts and analysis results
- Show correlation visualizations and charts

### 4. Access the Dashboard

Open your browser to: **http://localhost:3000**

## 📊 Dashboard Features

### Real-Time Alert Monitoring
- **Live Alert Feed**: New alerts appear every 10 seconds
- **Status Tracking**: Watch alerts transition from "analyzing" to "completed"
- **Source Identification**: Visual icons for different security systems
- **Severity Badges**: Color-coded severity levels (high, medium, low)

### Analysis Visualization
- **Correlation Details**: User-to-IP mappings with confidence scores
- **Processing Metrics**: Analysis time and correlation statistics
- **Interactive Charts**: Bar charts and pie charts for data visualization
- **Entity Tracking**: Lists of involved users and IP addresses

### Key Metrics Dashboard
- **Total Alerts**: Running count of processed alerts
- **High Severity**: Count of critical security events
- **Correlations Found**: Number of user-IP correlations discovered
- **Average Processing Time**: Performance metrics

## 🔧 API Endpoints

### Submit Alert for Analysis
```bash
POST http://localhost:8080/alerts
Content-Type: application/json

{
  "source": "aws_waf",
  "severity": "high", 
  "message": "Suspicious activity detected",
  "project_id": "demo-project-1",
  "raw_data": {
    "clientIP": "192.168.1.100",
    "uri": "/api/login"
  }
}
```

### Get Analysis Results
```bash
GET http://localhost:8080/analysis/{alert_id}
```

### Health Check
```bash
GET http://localhost:8080/health
```

## 🧠 Correlation Intelligence

### User-to-IP Correlation Methods

1. **Direct Correlation** (🎯): Same log contains both email and IP
   - Confidence: 90%
   - Most reliable correlation type

2. **Time Proximity** (⏰): Logs with emails and IPs within 5-minute windows
   - Confidence: 50-80% based on time distance
   - Accounts for user activity patterns

3. **Historical Correlation** (📊): Previously established user-IP relationships
   - Stored in database for future reference
   - Builds knowledge over time

### Confidence Scoring Factors

- **Time Proximity**: Closer timestamps = higher confidence
- **Same Company/Host**: Matching company codes boost confidence  
- **Source Diversity**: Multiple source systems increase correlation score
- **Historical Validation**: Previous correlations strengthen confidence

## 🎯 Supported Log Sources

- **🛡️ AWS WAF**: Amazon Web Application Firewall logs
- **🔷 Azure WAF**: Microsoft Azure Web Application Firewall logs  
- **☁️ Akamai WAF**: Akamai Web Application Firewall logs
- **🔒 Deep Security**: Trend Micro Deep Security system logs
- **🛡️ AWS GuardDuty**: Amazon GuardDuty threat detection alerts

## 📈 Sample Analysis Result

```json
{
  "alert_id": "1705312215123456789",
  "project_id": "demo-project-1",
  "user_correlations": [
    {
      "user_identifier": "user@example.com",
      "ip_address": "192.168.1.100", 
      "confidence_score": 0.85,
      "correlation_type": "time_proximity",
      "source_systems": ["aws_waf", "deep_security"]
    }
  ],
  "enrichment_data": {
    "correlation_stats": {
      "total_logs_analyzed": 15,
      "user_correlations_found": 3,
      "correlation_score": 0.72
    },
    "involved_users": ["user@example.com", "admin@company.com"],
    "involved_ips": ["192.168.1.100", "10.0.0.50"]
  },
  "processing_time_ms": 45
}
```

## 🛠️ Development

### Backend Development
```bash
cd server
go run *.go
```

### Frontend Development
```bash
cd client
npm start
```

### Adding New Log Sources

1. Update `server/normalizer.go` with new source detection logic
2. Add field mappings for the new log format
3. Test with sample logs from the new source
4. Update the client dashboard icons and labels

## 🔍 Troubleshooting

### Common Issues

1. **Database connection failed**: Ensure PostgreSQL is running via Docker
2. **Redis connection failed**: Check Redis service status
3. **Frontend can't connect**: Verify backend is running on port 8080
4. **No correlations found**: Check that logs contain both user identifiers and IP addresses

### Logs and Debugging

- **Backend logs**: Check terminal running `go run *.go`
- **Frontend logs**: Check browser developer console
- **Database queries**: Monitor PostgreSQL logs via Docker
- **Redis tasks**: Check Redis CLI for task queue status

## 🚀 Production Deployment

### Backend Scaling
- Deploy multiple Go server instances behind a load balancer
- Use managed PostgreSQL (RDS, Cloud SQL)
- Deploy Redis cluster for high availability
- Add monitoring with Prometheus and Grafana

### Frontend Deployment
- Build optimized React bundle: `npm run build`
- Deploy to CDN (Cloudflare, AWS CloudFront)
- Configure environment variables for API endpoints
- Enable HTTPS and security headers

### Security Considerations
- Implement proper API authentication (JWT, OAuth)
- Add request rate limiting and throttling
- Enable CORS for specific domains only
- Implement proper input validation and sanitization

## 📋 Next Steps

1. **🔐 Authentication**: Implement user authentication and authorization
2. **📊 Advanced ML**: Add machine learning models for anomaly detection
3. **🔔 Alerting**: Integrate with notification systems (Slack, PagerDuty)
4. **📱 Mobile**: Create mobile-responsive dashboard
5. **🔍 Search**: Add advanced search and filtering capabilities
6. **📈 Reporting**: Generate automated security reports
7. **🌐 Multi-tenancy**: Enhanced project isolation and access controls

## 📄 License

This project is a proof-of-concept for log correlation and analysis systems. 