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

## 🚀 How to Run the Project

### Prerequisites
- **Docker & Docker Compose** - For infrastructure services
- **Go 1.21+** - For the backend server
- **Node.js 18+** - For the React dashboard
- **Git** - For version control

### Step 1: Clone and Setup
```bash
git clone <your-repo-url>
cd soc-ml
```

### Step 2: Start Infrastructure Services
```bash
cd server
docker-compose up -d
```

This starts:
- **Loki** (port 3100) - Log storage and querying
- **PostgreSQL** (port 5432) - Analysis results storage  
- **Redis** (port 6379) - Task queue for async processing
- **Grafana** (port 3000) - Optional log visualization

**Wait 30 seconds** for services to fully initialize.

### Step 3: Start the Backend Server
```bash
# In the server directory
go mod tidy
go run *.go
```

You should see:
```
2024/01/15 10:30:00 Starting server on :8080
2024/01/15 10:30:00 Starting mock data generator...
2024/01/15 10:30:30 Generated mock alert: 1705312230123 (Source: aws_waf, Severity: high)
```

### Step 4: Start the Frontend Dashboard
```bash
# In a new terminal
cd client
npm install
npm start
```

The React development server will start on port 3000.

### Step 5: Access the Dashboard
Open your browser to: **http://localhost:3000**

You'll see:
- 📊 **Real-time metrics** at the top
- 🚨 **Live alert feed** on the left (new alerts every 10 seconds)
- 📈 **Analysis results** on the right (click "View Analysis" on any alert)

## 🎮 Testing the System

### Automatic Demo Mode
The system automatically generates realistic mock alerts every 10 seconds, simulating:
- AWS WAF SQL injection attempts
- Azure WAF XSS attacks  
- Deep Security file modifications
- Various severity levels and sources

### Manual Testing
Submit a custom alert:
```bash
curl -X POST http://localhost:8080/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "source": "aws_waf",
    "severity": "high",
    "message": "Suspicious SQL injection detected",
    "project_id": "test-project",
    "raw_data": {
      "clientIP": "192.168.1.100",
      "user_email": "test@company.com",
      "uri": "/api/users"
    }
  }'
```

### What You'll See
1. **Alert appears** in the dashboard with "analyzing" status
2. **Analysis completes** in 2-5 seconds, status changes to "completed"
3. **Click "View Analysis"** to see:
   - User-to-IP correlations with confidence scores
   - Processing time and statistics
   - Interactive charts showing correlation patterns
   - Lists of involved users and IP addresses

## 📊 Understanding the Dashboard

### Stats Cards (Top Row)
- **Total Alerts**: Running count of all processed alerts
- **High Severity**: Count of critical security events
- **Correlations Found**: Number of user-IP relationships discovered
- **Avg Processing Time**: How fast the system analyzes alerts

### Alert Feed (Left Panel)
- **Live indicator**: Green dot shows real-time updates
- **Source icons**: Visual identification of security systems
- **Severity badges**: Color-coded priority levels
- **Status tracking**: Watch alerts progress from analyzing → completed

### Analysis Results (Right Panel)
- **Summary metrics**: Logs analyzed, correlation score, processing time
- **User correlations**: Detailed user-to-IP mappings with confidence
- **Charts**: Visual representation of correlation patterns
- **Entity lists**: All users and IPs involved in the incident

## 🔧 Supported Log Sources

The system can normalize and correlate logs from:

### Cloud WAF Systems
- **🛡️ AWS WAF**: Amazon Web Application Firewall
- **🔷 Azure WAF**: Microsoft Azure Web Application Firewall  
- **☁️ Akamai WAF**: Akamai Edge Security

### Security Systems  
- **🔒 Deep Security**: Trend Micro endpoint protection
- **🛡️ AWS GuardDuty**: Amazon threat detection service

### Log Format Examples
Each system provides different information:
- **WAF logs**: User emails, request details, geographic data
- **System logs**: IP addresses, file access, process information
- **Security alerts**: Threat classifications, risk scores

## 🔍 Troubleshooting

### Common Issues

**"Database connection failed"**
```bash
# Check if PostgreSQL is running
docker-compose ps
# Restart if needed
docker-compose restart postgres
```

**"No correlations found"**
- Ensure logs contain both user identifiers AND IP addresses
- Check that timestamps are within the 15-minute analysis window
- Verify log normalization is extracting fields correctly

**"Frontend won't load"**
```bash
# Check if backend is running
curl http://localhost:8080/health
# Should return: {"status":"healthy"}
```

**"Alerts not appearing"**
- Mock alerts generate every 10 seconds automatically
- Check browser console for JavaScript errors
- Verify the proxy setting in `client/package.json`

### Performance Tuning

For high-volume environments:
- **Increase worker concurrency** in `asynq.Config{Concurrency: 20}`
- **Adjust analysis window** from ±15 minutes to ±5 minutes for faster processing
- **Add database indexes** for frequently queried fields
- **Scale horizontally** with multiple server instances

## 🚀 Production Deployment

### Backend Scaling
```bash
# Build optimized binary
go build -o soc-ml-server *.go

# Deploy with environment variables
export DB_HOST=your-postgres-host
export REDIS_HOST=your-redis-host
./soc-ml-server
```

### Frontend Deployment
```bash
# Build production bundle
npm run build

# Deploy to CDN or static hosting
# Configure API_URL environment variable
```

### Security Considerations
- **Authentication**: Implement JWT or OAuth for API access
- **Rate limiting**: Prevent API abuse with request throttling  
- **HTTPS**: Enable TLS for all communications
- **Input validation**: Sanitize all incoming log data
- **Database security**: Use connection pooling and prepared statements

## 📈 Next Steps

### Immediate Improvements
1. **🔐 Authentication**: Add user login and role-based access
2. **🔔 Alerting**: Integrate with Slack, PagerDuty, or email notifications
3. **📱 Mobile**: Make dashboard responsive for mobile devices
4. **🔍 Search**: Add filtering and search capabilities

### Advanced Features
1. **🤖 Machine Learning**: Implement anomaly detection algorithms
2. **📊 Advanced Analytics**: Add trend analysis and predictive modeling
3. **🌐 Multi-tenancy**: Support multiple organizations with data isolation
4. **📈 Reporting**: Generate automated security reports and dashboards

### Integration Options
1. **SIEM Integration**: Connect with Splunk, QRadar, or Sentinel
2. **Threat Intelligence**: Enrich with external threat feeds
3. **Incident Response**: Integrate with ticketing systems
4. **Compliance**: Add audit trails and compliance reporting

## 📄 License

This project is a proof-of-concept for log correlation and analysis systems. Use and modify according to your organization's needs. 