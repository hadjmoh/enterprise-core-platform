# Enterprise Core Platform - Protocol Configuration Guide

## Overview

The Enterprise Core Backend supports **25+ protocols** across Network, Application, IoT, SCADA, Infrastructure, Media, and Blockchain domains. All protocols are **environment-based** and can be enabled/disabled via configuration.

---

## Always-On Protocols

These protocols start automatically on server boot:

| Protocol | Port | Description |
|----------|------|-------------|
| **HTTP/REST** | 8080 | Health check (`/health`), HEC API (`/services/collector/event`), Metrics (`/metrics`) |
| **TCP Syslog** | 5140 | RFC3164/5424 syslog ingestion |
| **UDP Syslog** | 5141 | RFC3164/5424 syslog ingestion |
| **gRPC** | 50051 | High-performance RPC ingestion |
| **WebSocket** | 8080 | `/services/collector/ws` - streaming ingestion |
| **SOAP** | 8080 | `/services/collector/soap` - XML-based API |

---

## Optional Protocols (Environment-Based)

### **Messaging & Streaming**

```bash
# MQTT (IoT)
export MQTT_BROKER="tcp://localhost:1883"
export MQTT_TOPIC="logs/#"  # Optional, default: logs/#

# Kafka
export KAFKA_BROKERS="localhost:9092,localhost:9093"
export KAFKA_TOPIC="enterprise-logs"

# AMQP (RabbitMQ)
export AMQP_URL="amqp://guest:guest@localhost:5672/"
export AMQP_QUEUE="logs"  # Optional, default: logs
```

### **Infrastructure Monitoring**

```bash
# SNMP Trap Receiver (Port 162)
export ENABLE_SNMP=true

# DNS Query Log Parser
export DNS_LOG_PATH="/var/log/named/queries.log"

# DHCP Lease Monitor
export DHCP_LOG_PATH="/var/log/dhcpd.log"

# LDAP Authentication Logs
export LDAP_LOG_PATH="/var/log/slapd.log"

# Mail Server Logs (SMTP/IMAP/POP3)
export MAIL_LOG_PATHS="/var/log/mail.log,/var/log/mail.err"

# NTP Statistics
export NTP_LOG_PATH="/var/log/ntpstats/peerstats"
```

### **Industrial (OT/SCADA)**

```bash
# Modbus TCP (Port 502)
export ENABLE_MODBUS=true

# BACnet (Port 47808)
export ENABLE_BACNET=true

# DNP3 Log Parser
export DNP3_LOG_PATH="/var/log/dnp3/events.log"
```

### **Media & VoIP**

```bash
# SIP (Session Initiation Protocol)
export SIP_LOG_PATH="/var/log/asterisk/sip.log"

# RTP/RTSP (Media Streaming)
export RTP_LOG_PATH="/var/log/media/rtp.log"
```

### **Blockchain & Crypto**

```bash
# Bitcoin Node Logs
export BITCOIN_LOG_PATH="/var/log/bitcoin/debug.log"

# Ethereum Node Logs
export ETHEREUM_LOG_PATH="/var/log/geth/geth.log"

# LibP2P Event Streams
export LIBP2P_LOG_PATH="/var/log/ipfs/libp2p.log"
```

---

## Complete Startup Example

```bash
#!/bin/bash
# enterprise-start.sh

# Core Configuration
export LOG_LEVEL=info

# Messaging
export MQTT_BROKER="tcp://mqtt.example.com:1883"
export KAFKA_BROKERS="kafka1:9092,kafka2:9092"
export KAFKA_TOPIC="enterprise-events"
export AMQP_URL="amqp://user:pass@rabbitmq:5672/"

# Infrastructure
export ENABLE_SNMP=true
export DNS_LOG_PATH="/var/log/named/queries.log"
export DHCP_LOG_PATH="/var/log/dhcpd.log"

# Industrial
export ENABLE_MODBUS=true
export ENABLE_BACNET=true

# Start server
./server
```

---

## Protocol Summary

### **Network & Transport** (6)
- TCP, UDP, SCTP, QUIC, TLS, Syslog

### **Web & API** (5)
- HTTP/1.1, HTTP/2, WebSocket, gRPC, SOAP

### **Messaging & IoT** (3)
- MQTT, Kafka, AMQP

### **Infrastructure** (7)
- SNMP, DNS, DHCP, LDAP, Mail (SMTP/IMAP/POP3), NTP, File Watchers

### **Industrial (SCADA)** (3)
- Modbus, DNP3, BACnet

### **Media** (3)
- SIP, RTP, RTSP

### **Blockchain** (3)
- Bitcoin, Ethereum, LibP2P

### **Security & Parsing** (5)
- OAuth2, SAML, Kerberos, RADIUS, PGP/GPG

---

## Data Flow

```
Ingestion → Ring Buffer (10K events) → Disk Writer (100MB segments) → ./data/events/
```

## Monitoring

- **Health Check**: `http://localhost:8080/health`
- **Prometheus Metrics**: `http://localhost:8080/metrics`

### Key Metrics
- `enterprise_events_received_total{protocol="..."}`
- `enterprise_events_processed_total`
- `enterprise_events_dropped_total`
- `enterprise_buffer_size`
- `enterprise_processing_duration_seconds`

---

## File Output

Events are written to `./data/events/events-YYYYMMDD-HHMMSS.jsonl` in JSON-line format.

Example event:
```json
{"Timestamp":"2025-12-17T10:00:00Z","Source":"tcp","Data":{"message":"log event"}}
```

---

## Next Steps

1. **Enable desired protocols** via environment variables
2. **Start the server**: `./server`
3. **Monitor metrics**: `curl http://localhost:8080/metrics`
4. **Check data files**: `ls -lh ./data/events/`

For enterprise features (Enrichment, Correlation, HA, Multi-Tenancy), see Phase 2.5 planning document.
