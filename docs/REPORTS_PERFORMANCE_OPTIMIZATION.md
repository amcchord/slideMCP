# Slide Reports Performance Optimization Guide

## Executive Summary

The current report generation system experiences significant performance issues due to inefficient API usage patterns. This document provides a detailed analysis and actionable recommendations to improve performance by 3-5x.

## Current Performance Issues

### 1. No Date Filtering in API
- **Problem**: The backup and snapshot APIs don't support date range parameters
- **Impact**: Must fetch ALL records and filter client-side
- **Example**: To get backups for one day, might need to fetch thousands of records

### 2. Sequential Pagination
- **Problem**: Each agent requires multiple sequential API calls
- **Impact**: ~10-20 API calls per agent per day
- **Example**: Weekly report with 4 agents = ~420 API calls

### 3. Limited Parallelism
- **Problem**: Only 10 concurrent agent processing threads
- **Impact**: Artificial bottleneck on systems that could handle more

### 4. Redundant Data Fetching
- **Problem**: Weekly/monthly reports fetch the same data multiple times
- **Impact**: 7x redundancy for weekly reports, 30x for monthly

## Performance Metrics

### Current Implementation
- **Weekly Report (4 agents)**: ~15-20 seconds
- **API Calls**: ~420 calls
- **Data Transfer**: High redundancy

### Expected with Optimizations
- **Weekly Report (4 agents)**: ~3-5 seconds
- **API Calls**: ~100 calls (76% reduction)
- **Data Transfer**: Minimal redundancy

## Optimization Strategies

### 1. Short-term (Client-side) Optimizations

#### A. Intelligent Prefetching
```go
// Fetch all data for a date range in one pass
func prefetchBackupsForDateRange(agentID string, startDate, endDate time.Time) error {
    // Implementation that fetches all data once and caches by date
}
```

#### B. In-memory Caching
- Cache backup/snapshot data by agent+date
- Reuse cached data for multi-day reports
- Thread-safe implementation included

#### C. Increased Parallelism
- Increase concurrent agent processing from 10 to 20-30
- Separate prefetch parallelism from report generation

#### D. Batch Processing
- Group related API calls together
- Process entire date ranges instead of individual days

### 2. Medium-term (API) Improvements

#### A. Add Date Range Filtering
```
GET /v1/backup?agent_id=xxx&start_date=2024-01-01&end_date=2024-01-07
```

#### B. Bulk Data Endpoints
```
POST /v1/reports/bulk
{
  "agent_ids": ["a_xxx", "a_yyy"],
  "start_date": "2024-01-01",
  "end_date": "2024-01-07",
  "include": ["backups", "snapshots"]
}
```

#### C. Aggregation Endpoints
```
GET /v1/reports/summary?period=week&date=2024-01-01
```

### 3. Long-term (Architecture) Improvements

#### A. Database Optimizations
- Add database indexes on date columns
- Create materialized views for common report queries
- Implement database-level aggregation

#### B. Caching Layer
- Redis/Memcached for frequently accessed data
- TTL-based cache invalidation
- Pre-generate common reports

#### C. Async Report Generation
- Queue-based report generation
- Progress tracking
- Email/webhook notifications

## Implementation Priority

### Phase 1 (Immediate)
1. Implement client-side prefetching ✓
2. Add in-memory caching ✓
3. Increase parallelism limits
4. Deploy and monitor

### Phase 2 (1-2 weeks)
1. Add date filtering to backup/snapshot APIs
2. Implement bulk data endpoints
3. Add server-side caching

### Phase 3 (1 month)
1. Database query optimization
2. Materialized views
3. Async report system

## Code Examples

### Optimized Weekly Report
```go
func generateWeeklyReportOptimized(args map[string]interface{}) (string, error) {
    // 1. Get date range
    startOfWeek, endOfWeek := getWeekBounds(args["date"])
    
    // 2. Get all agents
    agents := getAgentsForScope(args)
    
    // 3. Prefetch all data in parallel
    prefetchAllAgentData(agents, startOfWeek, endOfWeek)
    
    // 4. Generate reports from cache
    return generateReportsFromCache(agents, startOfWeek, endOfWeek)
}
```

### API Enhancement Example
```go
// New API endpoint for bulk reporting data
func handleBulkReportData(w http.ResponseWriter, r *http.Request) {
    var req BulkReportRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Fetch all data in one optimized query
    data := db.Query(`
        SELECT agent_id, date, 
               COUNT(*) as total_backups,
               SUM(CASE WHEN status='succeeded' THEN 1 ELSE 0 END) as successful
        FROM backups
        WHERE agent_id IN (?) 
          AND started_at BETWEEN ? AND ?
        GROUP BY agent_id, DATE(started_at)
    `, req.AgentIDs, req.StartDate, req.EndDate)
    
    return json.NewEncoder(w).Encode(data)
}
```

## Monitoring and Metrics

### Key Metrics to Track
1. Report generation time
2. API calls per report
3. Cache hit rate
4. Memory usage
5. Error rates

### Success Criteria
- 70% reduction in report generation time
- 75% reduction in API calls
- <5 second generation for weekly reports
- <20 seconds for monthly reports

## Conclusion

The proposed optimizations can deliver significant performance improvements with minimal risk. The phased approach allows for incremental improvements while maintaining system stability. Starting with client-side optimizations provides immediate benefits while longer-term API and architecture changes are planned and implemented. 