package grpc

import (
	"context"
	pb "enterprise-core/backend/internal/proto"
	"enterprise-core/backend/pkg/logger"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedIngestorServer
	logger *logger.Logger
	port   string
	server *grpc.Server
}

func NewServer(port string, logger *logger.Logger) *Server {
	return &Server{
		port:   port,
		logger: logger,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	pb.RegisterIngestorServer(s.server, s)

	s.logger.Info("gRPC Server started", "port", s.port)

	go func() {
		if err := s.server.Serve(lis); err != nil {
			s.logger.Error("gRPC Serve error", err)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

func (s *Server) SendEvent(ctx context.Context, req *pb.LogEvent) (*pb.IngestResponse, error) {
	s.logger.Info("Received gRPC Event", "host", req.Host, "source", req.Source)
	return &pb.IngestResponse{Success: true, Message: "Ack"}, nil
}

func (s *Server) StreamEvents(stream pb.Ingestor_StreamEventsServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		s.logger.Info("Received gRPC Stream Event", "host", req.Host)
		if err := stream.Send(&pb.IngestResponse{Success: true, Message: "Ack"}); err != nil {
			return err
		}
	}
}
