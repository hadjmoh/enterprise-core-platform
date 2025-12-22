package main

import (
	"context"
	"enterprise-core/backend/internal/api"
	"enterprise-core/backend/internal/alerting"
	"enterprise-core/backend/internal/analytics"
	"enterprise-core/backend/internal/blockchain"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/correlation"
	"enterprise-core/backend/internal/detection"
	"enterprise-core/backend/internal/enrichment"
	"enterprise-core/backend/internal/governance"
	"enterprise-core/backend/internal/audit"
	"enterprise-core/backend/internal/auth"
	"enterprise-core/backend/internal/compliance"
	"enterprise-core/backend/internal/cluster"
	internalGrpc "enterprise-core/backend/internal/grpc"
	"enterprise-core/backend/internal/ha"
	"enterprise-core/backend/internal/industrial"
	"enterprise-core/backend/internal/infrastructure"
	"enterprise-core/backend/internal/media"
	"enterprise-core/backend/internal/messaging"
	"enterprise-core/backend/internal/network"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/internal/query"
	"enterprise-core/backend/internal/query/cost"
	"enterprise-core/backend/internal/query/pilot"
	"enterprise-core/backend/internal/security"
	"enterprise-core/backend/internal/security/risk"
	"enterprise-core/backend/internal/storage"
	"enterprise-core/backend/internal/simulation"
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
	riskEngine := risk.NewRiskEngine(logg)
	riskDecay := risk.NewRiskDecayService(riskEngine, 1*time.Hour, 0.5, logg)
	go riskDecay.Run(context.Background())

	// Initialize Compliance & Privacy Policies
	policyRegistry := compliance.NewPolicyRegistry(logg)
	
	// Initialize Data Governance Engines (Session 7.6)
	vault := compliance.NewPrivacyVault("enterprise-salt-secure-random")
	merkleTree := compliance.NewMerkleTree(nil)

	// Initialize Cost & Safety Engine (Session 7.2)
	costEngine := cost.NewPolicyEngine()

	// Initialize UEBA Engine
	uebaEngine := security.NewUEBAEngine(logg)

	// Initialize Analytics & Feature Store (Session 7.4)
	featureStore := analytics.NewFeatureStore("./data/features.json")
	featureExtractor := analytics.NewFeatureExtractor(featureStore)
	anomalyDetector := analytics.NewAnomalyDetector()
	explainerEngine := analytics.NewExplainerEngine()
	feedbackStore := analytics.NewFeedbackStore("./data/feedback.json")

	// Initialize Trust Graph & Lineage (Session 7.8)
	lineageTracker := security.NewLineageTracker("enterprise-secret-key-123", "ingest-master-01")
	trustGraph := security.NewTrustGraph()

	// Initialize Security Control Plane (Session 7.10)
	governor := governance.NewGovernor(logg)
	policyEngine := governance.NewPolicyEngine()
	driftDetector := governance.NewDriftDetector(map[string]interface{}{
		"max_ingest_buffer": 10000,
		"retention_days":    90,
		"encryption_enabled": true,
	}, logg)

	// Initialize Pilot Engine (Session 7.2)
	pilotEngine := pilot.NewPilotEngine(costEngine)

	// Initialize Cluster Intelligence (Session 7.7 & 8.1)
	clusterMonitor := cluster.NewResourceMonitor()
	autoScaler := cluster.NewAutoScaler(clusterMonitor, costEngine, logg)
	go autoScaler.Run(context.Background())

	nodeMgr := cluster.NewNodeManager()
	shardMgr := cluster.NewShardManager()
	shcMgr := cluster.NewSHCManager("node-primary-01", nodeMgr, logg)
	loadBalancer := cluster.NewLoadBalancer(cluster.StrategyLeastLoaded, logg)
	mrEngine := cluster.NewMapReduceEngine(nodeMgr, shardMgr, loadBalancer, logg)
	discovery := cluster.NewPeerDiscovery(nodeMgr, []string{"127.0.0.1:50051"}, logg) // Self and static peers
	hbWorker := cluster.NewHeartbeatWorker("node-primary-01", "127.0.0.1:50051", discovery, nodeMgr, logg)
	hbWorker.Start()

	clusterMaster := cluster.NewClusterMaster(nodeMgr, shardMgr, logg)
	clusterMaster.Start(context.Background())

	deploySrv := cluster.NewDeploymentServer(nodeMgr, logg)
	txCoord := cluster.NewTransactionCoordinator(nodeMgr, logg)
	rollbackMgr := cluster.NewRollbackManager(deploySrv, nodeMgr, txCoord, logg)
	
	// Multi-site and DR (Session 8.8)
	siteMgr := cluster.NewSiteManager(logg)
	replMgr := cluster.NewReplicationManager(siteMgr, cluster.ReplicationAsync, logg)
	drCoord := cluster.NewDRCoordinator(siteMgr, replMgr, txCoord, logg)

	// Audit Logger needed for Correlation and Dispatcher
	auditor := audit.NewLogger("./data/audit.log", logg)

	// Initialize SOAR Orchestrator
	soarOrch := security.NewOrchestrator(auditor, logg)
	soarOrch.RegisterAction(&security.BlockIPAction{Logger: logg})
	soarOrch.RegisterAction(&security.DisableUserAction{Logger: logg})

	// Initialize Hunting Manager
	huntingMgr := security.NewHuntingManager()

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
	
	// Create Search Dispatcher (Phase 4 & 8)
	dispatcher := query.NewDispatcher(storageEngine, auditor, policyRegistry, governor, mrEngine, logg)

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
		featureExtractor,
		vault,
		merkleTree,
		lineageTracker,
		trustGraph,
		governor,
		logg,
	)

	// Initialize Simulation & Replay Engine (Session 7.9)
	simEngine := simulation.NewEngine(storageEngine, ingestPipeline, logg)

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

	// Initialize Cluster gRPC Server (Session 8.1 & 8.2)
	_ = cluster.NewGRPCServer(nodeMgr, dispatcher, shcMgr, logg)
	// Note: In a real production app, we would start the gRPC listener here

	// 4. Setup HTTP Router
	authSvc := auth.NewMockProvider()
	mux := api.NewRouter(ingestPipeline, dispatcher, authSvc, riskEngine, uebaEngine, soarOrch, policyRegistry, huntingMgr, mitreMgr, costEngine, pilotEngine, featureStore, anomalyDetector, explainerEngine, feedbackStore, merkleTree, autoScaler, clusterMonitor, nodeMgr, shardMgr, shcMgr, deploySrv, rollbackMgr, drCoord, trustGraph, simEngine, governor, driftDetector, policyEngine, logg)
	
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
