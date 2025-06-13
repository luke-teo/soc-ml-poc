# SOC ML - Log Correlation and Analysis System

A comprehensive system for correlating logs across multiple security systems and performing automated analysis when alerts are triggered, featuring a real-time web dashboard.

## 🎯 What This System Solves

In modern cybersecurity, organizations use multiple security tools that generate logs independently:
- **Cloud WAF logs** contain user emails but limited IP information
- **System logs** contain IP addresses but no user identification
- **Security alerts** trigger from different sources at different times

**The Challenge**: When a security alert occurs, analysts need to quickly understand:
- Which user was involved?
- What other systems were affected?
- Are there related activities across different security tools?

**Our Solution**: Automatically correlate logs across different sources to build a complete picture of security events, connecting user identities with network activities through intelligent time-based analysis.

## 🏗️ System Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Alert Source  │───▶│  Alert Processor │───▶│ Log Correlation │
│ (WAF, Security) │    │   (HTTP Server)  │    │    Engine       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│     Loki        │◀───│   Log Fetcher    │◀───│   Normalizer    │
│   (Log Store)   │    │  (LogQL Client)  │    │    Engine       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   PostgreSQL    │◀───│ Analysis Results │◀───│  Enrichment     │
│   (Results)     │    │     Storage      │    │    Engine       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌──────────────────┐
│  React Dashboard│◀───│   REST API       │
│   (Frontend)    │    │   (Chi Router)   │
└─────────────────┘    └──────────────────┘
```

### Core Components

1. **Alert Processor** - Receives security alerts via HTTP API
2. **Log Fetcher** - Queries Grafana Loki for logs within ±15 minute windows
3. **Log Normalizer** - Extracts common fields from different log formats
4. **Correlation Engine** - Builds user-to-IP relationships using multiple methods
5. **Enrichment Engine** - Adds context and statistics to analysis results
6. **React Dashboard** - Real-time visualization of alerts and correlations

## 🧠 How Correlation Works (Explained Simply)

### The Problem: Connecting the Dots

Imagine you're a detective investigating a case:
- **Witness A** saw someone with a red car at 2:00 PM
- **Witness B** saw "John Smith" near the scene at 2:05 PM
- **Question**: Was John Smith driving the red car?

This is exactly what our system does with security logs, but instead of witnesses, we have different security systems, and instead of cars and names, we have IP addresses and email addresses.

### Our Correlation Methods

#### 1. 🎯 Direct Correlation (Confidence: 90%)
**What it is**: The same log entry contains both a user email and an IP address.

**Example**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "user_email": "john.doe@company.com",
  "client_ip": "192.168.1.100",
  "action": "login_attempt"
}
```

**Why it's reliable**: When one system captures both pieces of information simultaneously, we can be very confident they're related.

#### 2. ⏰ Time Proximity Correlation (Confidence: 50-80%)
**What it is**: We find logs with emails and logs with IP addresses that occur close together in time.

**Example**:
- **2:00 PM**: AWS WAF log shows `john.doe@company.com` accessing `/login`
- **2:02 PM**: Deep Security log shows suspicious activity from `192.168.1.100`

**The Logic**: If these events happen within a few minutes of each other, there's a good chance the same person is involved.

**Confidence Factors**:
- **1 minute apart**: 80% confidence
- **5 minutes apart**: 60% confidence  
- **15 minutes apart**: 50% confidence

#### 3. 📊 Historical Correlation (Confidence: Variable)
**What it is**: We remember previous correlations and use them to strengthen new ones.

**Example**: If we've seen `john.doe@company.com` and `192.168.1.100` together multiple times before, we're more confident when we see them again.

### Confidence Scoring Algorithm

Our system calculates a confidence score (0-100%) for each correlation:

```
Base Score = 50%

+ Time Proximity Bonus:
  - Same minute: +30%
  - Within 5 minutes: +20%
  - Within 15 minutes: +10%

+ Context Bonus:
  - Same company/host: +20%
  - Same geographic location: +10%

+ Historical Bonus:
  - Previously seen together: +15%
  - Multiple source confirmation: +10%

Final Score = min(100%, Base + All Bonuses)
```

### Real-World Example

**Scenario**: A SQL injection alert triggers at 2:15 PM

**Step 1 - Log Collection**: System queries all logs from 2:00 PM to 2:30 PM
```
2:10 PM - Azure WAF: john.doe@company.com accessed /admin
2:12 PM - System Log: 192.168.1.100 attempted file access
2:15 PM - AWS WAF: SQL injection blocked from 192.168.1.100
2:18 PM - Deep Security: Suspicious process on server-01
```

**Step 2 - Correlation Analysis**:
- **Found**: `john.doe@company.com` (2:10 PM) and `192.168.1.100` (2:12 PM)
- **Time Gap**: 2 minutes → High confidence
- **Pattern**: User login followed by suspicious activity → Likely related

**Step 3 - Confidence Calculation**:
```
Base Score: 50%
+ Time proximity (2 min): +25%
+ Same session pattern: +15%
= 90% confidence
```

**Step 4 - Result**: The system determines with 90% confidence that `john.doe@company.com` was using `192.168.1.100` when the SQL injection occurred.

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