package main

import (
	"context"
	"enterprise-core/backend/internal/api"
	"enterprise-core/backend/internal/alerting"
	"enterprise-core/backend/internal/blockchain"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/correlation"
	"enterprise-core/backend/internal/detection"
	"enterprise-core/backend/internal/enrichment"
	"enterprise-core/backend/internal/compliance"
	internalGrpc "enterprise-core/backend/internal/grpc"
	"enterprise-core/backend/internal/ha"
	"enterprise-core/backend/internal/industrial"
	"enterprise-core/backend/internal/infrastructure"
	"enterprise-core/backend/internal/media"
	"enterprise-core/backend/internal/messaging"
	"enterprise-core/backend/internal/network"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/internal/query"
	"enterprise-core/backend/internal/security"
	"enterprise-core/backend/internal/storage"
	"enterprise-core/backend/internal/tenant"
	"enterprise-core/backend/internal/writer"
	"enterprise-core/backend/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Initialize Logger
	logg := logger.New()
	logg.Info("Starting Enterprise Core Backend...")

	// 1. Initialize Ring Buffer
	ringBuffer := buffer.NewRingBuffer(10000, logg)

	// 2. Initialize Enterprise Services (Phase 2.5)
	alertRouter := alerting.NewRouter(logg)
	// Add Slack channel if webhook provided
	if slackURL := os.Getenv("SLACK_WEBHOOK"); slackURL != "" {
		alertRouter.AddChannel(alerting.NewSlackChannel(slackURL))
	}

	tenantMgr := tenant.NewManager(logg)
	idResolver := auth.NewMockIdentityResolver()
	enrichmentPipe := enrichment.NewPipeline(logg, idResolver)
	
	// Initialize Risk Engine for SIEM
	riskEngine := security.NewRiskEngine(logg)
	riskDecay := security.NewRiskDecayService(riskEngine, 1*time.Hour, 0.5, logg)
	go riskDecay.Run(context.Background())

	// Initialize Compliance & Privacy Policies
	policyRegistry := compliance.NewPolicyRegistry(logg)

	// Initialize UEBA Engine
	uebaEngine := security.NewUEBAEngine(logg)

	// Initialize SOAR Orchestrator
	soarOrch := security.NewOrchestrator(auditor, logg)
	soarOrch.RegisterAction(&security.BlockIPAction{Logger: logg})
	soarOrch.RegisterAction(&security.DisableUserAction{Logger: logg})

	// Initialize Hunting Manager
	huntingMgr := security.NewHuntingManager()

	// Audit Logger needed for Correlation and Dispatcher
	auditor := audit.NewLogger("./data/audit.log", logg)
	
	correlationEngine := correlation.NewEngine(5*time.Minute, logg, auditor, riskEngine)
	detectionEngine := detection.NewEngine(logg)
	
	// Initialize MITRE Manager
	mitreMgr := security.NewMitreManager(detectionEngine, correlationEngine)
	
	// Create Storage Engine (Session 1)
	storageEngine := storage.NewFileStorageEngine("./data/storage", logg)
	// Load detection rules from file if exists
	if err := detectionEngine.LoadRulesFromFile("config/detection_rules.yaml"); err != nil {
		logg.Warn("Failed to load detection rules", "error", err)
	}
	
	// Create Search Dispatcher (Phase 4)
	dispatcher := query.NewDispatcher(storageEngine, auditor, policyRegistry, logg)

	backpressure := buffer.NewBackpressureController(10000, 0.8, logg)

	// 3. Initialize Unified Ingestion Pipeline
	ingestPipeline := pipeline.NewIngestionPipeline(
		enrichmentPipe,
		correlationEngine,
		detectionEngine,
		alertRouter,
		ringBuffer,
		storageEngine,
		backpressure,
		tenantMgr,
		uebaEngine,
		riskEngine,
		soarOrch,
		logg,
	)

	// 4. Initialize Disk Writer
	diskWriter, err := writer.NewDiskWriter("./data/events", 100, logg)
	if err != nil {
		log.Fatalf("Failed to create disk writer: %v", err)
	}
	ringBuffer.AddHandler(diskWriter)
	ringBuffer.Start()

	// 5. Initialize Health Monitor (HA)
	healthMonitor := ha.NewHealthMonitor(logg)
	healthMonitor.AddChecker("buffer", func() (string, string) {
		if ringBuffer.Size() > 9000 {
			return "degraded", "Buffer nearly full"
		}
		return "up", ""
	})

	// 6. Start TCP Listener Handlers
	syslogHandler := network.NewSyslogHandler(ingestPipeline, logg)

	// 3. Start TCP Syslog Listener (5140)
	tcpSyslog := network.NewTCPListener(5140, syslogHandler, logg)
	if err := tcpSyslog.Start(context.Background()); err != nil {
		logg.Error("Failed to start TCP Syslog", err)
	}

	// 4. Start UDP Syslog Listener (5141)
	udpSyslog := network.NewUDPListener(5141, syslogHandler, logg)
	if err := udpSyslog.Start(context.Background()); err != nil { // UDP Start is non-blocking in our impl
		logg.Error("Failed to start UDP Syslog", err)
	}

	// 4. Setup HTTP Router
	authSvc := auth.NewMockProvider()
	mux := api.NewRouter(ingestPipeline, dispatcher, authSvc, riskEngine, uebaEngine, soarOrch, policyRegistry, huntingMgr, mitreMgr, logg)
	
	// Add Prometheus metrics endpoint
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	// 6. Configure HTTP Server
	srv := &http.Server{
		Addr:    ":8080", // Standard backend port
		Handler: mux,
	}

	// 5. Start HTTP Server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	logg.Info("Server is listening on port 8080")

	// 6. Start gRPC Server
	grpcServer := internalGrpc.NewServer("50051", logg)
	if err := grpcServer.Start(); err != nil {
		logg.Error("Failed to start gRPC Server", err)
	}

	// 7. Optional: Start MQTT Client (if broker configured)
	var mqttClient *messaging.MQTTClient
	if mqttBroker := os.Getenv("MQTT_BROKER"); mqttBroker != "" {
		mqttTopic := os.Getenv("MQTT_TOPIC")
		if mqttTopic == "" {
			mqttTopic = "logs/#"
		}
		mqttClient = messaging.NewMQTTClient(mqttBroker, "enterprise-ingestor", mqttTopic, ingestPipeline, logg)
		if err := mqttClient.Start(); err != nil {
			logg.Error("Failed to start MQTT Client", err)
		}
	}

	// 8. Optional: Start Kafka Consumer (if brokers configured)
	var kafkaConsumer *messaging.KafkaConsumer
	if kafkaBrokers := os.Getenv("KAFKA_BROKERS"); kafkaBrokers != "" {
		kafkaConsumer = messaging.NewKafkaConsumer(
			strings.Split(kafkaBrokers, ","),
			os.Getenv("KAFKA_TOPIC"),
			"enterprise-ingestor-group",
			ingestPipeline,
			logg,
		)
		if err := kafkaConsumer.Start(); err != nil {
			logg.Error("Failed to start Kafka Consumer", err)
		}
	}

	// 9. Optional: Start AMQP Consumer (if URL configured)
	var amqpConsumer *messaging.AMQPConsumer
	if amqpURL := os.Getenv("AMQP_URL"); amqpURL != "" {
		amqpQueue := os.Getenv("AMQP_QUEUE")
		if amqpQueue == "" {
			amqpQueue = "logs"
		}
		amqpConsumer = messaging.NewAMQPConsumer(amqpURL, amqpQueue, ingestPipeline, logg)
		if err := amqpConsumer.Start(); err != nil {
			logg.Error("Failed to start AMQP Consumer", err)
		}
	}

	// 10. Optional: Start Infrastructure Collectors (env-based)
	var snmpReceiver *infrastructure.SNMPTrapReceiver
	if os.Getenv("ENABLE_SNMP") == "true" {
		snmpReceiver = infrastructure.NewSNMPTrapReceiver(162, ingestPipeline, logg)
		if err := snmpReceiver.Start(); err != nil {
			logg.Error("Failed to start SNMP", err)
		}
	}

	var dnsParser *infrastructure.DNSLogParser
	if dnsLogPath := os.Getenv("DNS_LOG_PATH"); dnsLogPath != "" {
		dnsParser = infrastructure.NewDNSLogParser(dnsLogPath, logg)
		dnsParser.Start()
	}

	var dhcpMonitor *infrastructure.DHCPMonitor
	if dhcpLogPath := os.Getenv("DHCP_LOG_PATH"); dhcpLogPath != "" {
		dhcpMonitor = infrastructure.NewDHCPMonitor(dhcpLogPath, logg)
		dhcpMonitor.Start()
	}

	var ldapCollector *infrastructure.LDAPCollector
	if ldapLogPath := os.Getenv("LDAP_LOG_PATH"); ldapLogPath != "" {
		ldapCollector = infrastructure.NewLDAPCollector(ldapLogPath, logg)
		ldapCollector.Start()
	}

	var mailCollector *infrastructure.MailCollector
	if mailLogPaths := os.Getenv("MAIL_LOG_PATHS"); mailLogPaths != "" {
		mailCollector = infrastructure.NewMailCollector(strings.Split(mailLogPaths, ","), logg)
		mailCollector.Start()
	}

	var ntpCollector *infrastructure.NTPCollector
	if ntpLogPath := os.Getenv("NTP_LOG_PATH"); ntpLogPath != "" {
		ntpCollector = infrastructure.NewNTPCollector(ntpLogPath, logg)
		ntpCollector.Start()
	}

	// 11. Optional: Start Industrial Protocols (env-based)
	var modbusListener *industrial.ModbusListener
	if os.Getenv("ENABLE_MODBUS") == "true" {
		modbusListener = industrial.NewModbusListener(502, ingestPipeline, logg)
		if err := modbusListener.Start(); err != nil {
			logg.Error("Failed to start Modbus", err)
		}
	}

	var bacnetCollector *industrial.BACnetCollector
	if os.Getenv("ENABLE_BACNET") == "true" {
		bacnetCollector = industrial.NewBACnetCollector(47808, ingestPipeline, logg)
		if err := bacnetCollector.Start(); err != nil {
			logg.Error("Failed to start BACnet", err)
		}
	}

	var dnp3Parser *industrial.DNP3Parser
	if dnp3LogPath := os.Getenv("DNP3_LOG_PATH"); dnp3LogPath != "" {
		dnp3Parser = industrial.NewDNP3Parser(dnp3LogPath, ingestPipeline, logg)
		dnp3Parser.Start()
	}

	// 12. Optional: Start Media Protocols (env-based)
	var sipCollector *media.SIPCollector
	if sipLogPath := os.Getenv("SIP_LOG_PATH"); sipLogPath != "" {
		sipCollector = media.NewSIPCollector(sipLogPath, ingestPipeline, logg)
		sipCollector.Start()
	}

	var rtpCollector *media.RTPCollector
	if rtpLogPath := os.Getenv("RTP_LOG_PATH"); rtpLogPath != "" {
		rtpCollector = media.NewRTPCollector(rtpLogPath, logg)
		rtpCollector.Start()
	}

	// 13. Optional: Start Blockchain Collectors (env-based)
	var btcCollector *blockchain.CryptoNodeCollector
	if btcLogPath := os.Getenv("BITCOIN_LOG_PATH"); btcLogPath != "" {
		btcCollector = blockchain.NewCryptoNodeCollector(btcLogPath, "bitcoin", ingestPipeline, logg)
		btcCollector.Start()
	}

	var ethCollector *blockchain.CryptoNodeCollector
	if ethLogPath := os.Getenv("ETHEREUM_LOG_PATH"); ethLogPath != "" {
		ethCollector = blockchain.NewCryptoNodeCollector(ethLogPath, "ethereum", ingestPipeline, logg)
		ethCollector.Start()
	}

	var libp2pCollector *blockchain.LibP2PCollector
	if libp2pLogPath := os.Getenv("LIBP2P_LOG_PATH"); libp2pLogPath != "" {
		libp2pCollector = blockchain.NewLibP2PCollector(libp2pLogPath, logg)
		libp2pCollector.Start()
	}

	// 14. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logg.Info("Shutting down server...")

	// Stop all listeners and collectors
	tcpSyslog.Stop()
	udpSyslog.Stop()
	grpcServer.Stop()
	
	if mqttClient != nil {
		mqttClient.Stop()
	}
	if kafkaConsumer != nil {
		kafkaConsumer.Stop()
	}
	if amqpConsumer != nil {
		amqpConsumer.Stop()
	}

	// Stop infrastructure collectors
	if snmpReceiver != nil {
		snmpReceiver.Stop()
	}
	if dnsParser != nil {
		dnsParser.Stop()
	}
	if dhcpMonitor != nil {
		dhcpMonitor.Stop()
	}
	if ldapCollector != nil {
		ldapCollector.Stop()
	}
	if mailCollector != nil {
		mailCollector.Stop()
	}
	if ntpCollector != nil {
		ntpCollector.Stop()
	}

	// Stop industrial protocols
	if modbusListener != nil {
		modbusListener.Stop()
	}
	if bacnetCollector != nil {
		bacnetCollector.Stop()
	}
	if dnp3Parser != nil {
		dnp3Parser.Stop()
	}

	// Stop media protocols
	if sipCollector != nil {
		sipCollector.Stop()
	}
	if rtpCollector != nil {
		rtpCollector.Stop()
	}

	// Stop blockchain collectors
	if btcCollector != nil {
		btcCollector.Stop()
	}
	if ethCollector != nil {
		ethCollector.Stop()
	}
	if libp2pCollector != nil {
		libp2pCollector.Stop()
	}

	// Stop buffer and writer
	ringBuffer.Stop()
	diskWriter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logg.Error("Server forced to shutdown:", err)
	}

	logg.Info("Server exiting")
}
