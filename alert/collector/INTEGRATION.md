# Alert Collector Integration Guide

This guide shows how to integrate the alert collection module into Nightingale's pushgw router.

## Integration Steps

### 1. Import the collector package

Add the following import to `/workspace/pushgw/router/router.go`:

```go
import (
    "github.com/ccfos/nightingale/v6/alert/collector"
)
```

### 2. Add AlertCollectorHandler to Router struct

Extend the Router struct in `/workspace/pushgw/router/router.go`:

```go
type Router struct {
    // ... existing fields ...
    
    // New field for alert collector
    AlertCollector *collector.AlertCollectorHandler
}
```

### 3. Initialize the handler in New function

Update the `New` function in `/workspace/pushgw/router/router.go`:

```go
func New(...) *Router {
    rt := &Router{
        // ... existing initialization ...
    }
    
    // Initialize alert collector
    alertRegistrar := collector.NewAlertMetricsRegistrar("n9e")
    rt.AlertCollector = collector.NewAlertCollectorHandler(
        alertRegistrar,
        rt.Writers,
    )
    
    return rt
}
```

### 4. Register HTTP endpoints in Config function

Add new routes in the `Config` function in `/workspace/pushgw/router/router.go`:

```go
func (rt *Router) Config(r *gin.Engine) {
    // ... existing routes ...
    
    // Alert collector endpoints
    service.POST("/alerts", rt.AlertCollector.CollectAlerts)
    service.POST("/alert", rt.AlertCollector.CollectSingleAlert)
}
```

## Usage Examples

### Send a single alert event

```bash
curl -X POST http://localhost:17000/v1/n9e/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "ident": "web-server-01",
    "name": "HighCPU",
    "severity": "critical",
    "value": 95.5,
    "alert_type": "system",
    "source": "zabbix",
    "start_time": 1715184000,
    "labels": {
      "host_ip": "192.168.1.100",
      "environment": "production",
      "service": "api"
    }
  }'
```

### Send multiple alert events

```bash
curl -X POST http://localhost:17000/v1/n9e/alerts \
  -H "Content-Type: application/json" \
  -d '[
    {
      "ident": "web-server-01",
      "name": "HighCPU",
      "severity": "critical",
      "value": 95.5,
      "alert_type": "system",
      "source": "zabbix",
      "start_time": 1715184000
    },
    {
      "ident": "web-server-02",
      "name": "DiskFull",
      "severity": "warning",
      "value": 85.0,
      "alert_type": "system",
      "source": "prometheus",
      "start_time": 1715184000
    }
  ]'
```

## API Response

### Success Response

```json
{
  "status": "success",
  "total": 2,
  "metrics": 4,
  "errors": 0
}
```

### Error Response

```json
{
  "error": "invalid request body",
  "details": "json: unexpected end of JSON input"
}
```

## Generated Metrics

The alert collector generates the following metrics:

### 1. n9e_alert_count_total
Total count of alert events received.

Labels:
- `alert_name`: Name of the alert
- `severity`: Alert severity (critical, warning, info)
- `alert_type`: Type of alert (system, application, business)
- `source`: Alert source (prometheus, zabbix, etc.)
- `ident`: Target identifier

### 2. n9e_alert_firing
Current firing alerts with their values.

Labels:
- Same as `n9e_alert_count_total`

### 3. n9e_alert_resolved_total
Duration of resolved alerts in seconds.

Labels:
- Same as `n9e_alert_count_total`

## Target Registration

The alert collector can automatically register targets based on alert events.

To enable target registration, modify the handler initialization:

```go
func New(...) *Router {
    // ...
    
    // Enable target registration
    targetRegistrar := collector.NewTargetRegistrar(ctx, 1, "alert-collector")
    
    // Create handler with target registration
    rt.AlertCollector = collector.NewAlertCollectorHandler(
        alertRegistrar,
        writer,
        targetRegistrar,
    )
    
    return rt
}
```

## Configuration

No additional configuration is required. The alert collector uses sensible defaults.

Optional configuration in `config.toml`:

```toml
[AlertCollector]
enabled = true
metric_prefix = "n9e"
default_group_id = 1
engine_name = "alert-collector"
```
