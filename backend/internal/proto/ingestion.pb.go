package proto

import (
	"context"
	"google.golang.org/grpc"
)

// LogEvent mimics the proto message
type LogEvent struct {
	Time       int64             `json:"time"`
	Host       string            `json:"host"`
	Source     string            `json:"source"`
	Sourcetype string            `json:"sourcetype"`
	Index      string            `json:"index"`
	Data       string            `json:"data"`
	Fields     map[string]string `json:"fields"`
}

// IngestResponse mimics the proto response
type IngestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// IngestorServer is the server API
type IngestorServer interface {
	SendEvent(context.Context, *LogEvent) (*IngestResponse, error)
	StreamEvents(Ingestor_StreamEventsServer) error
}

// UnimplementedIngestorServer for forward compatibility
type UnimplementedIngestorServer struct{}

func (UnimplementedIngestorServer) SendEvent(context.Context, *LogEvent) (*IngestResponse, error) {
	return nil, nil
}
func (UnimplementedIngestorServer) StreamEvents(Ingestor_StreamEventsServer) error {
	return nil
}

// Ingestor_StreamEventsServer is the stream interface
type Ingestor_StreamEventsServer interface {
	Send(*IngestResponse) error
	Recv() (*LogEvent, error)
	grpc.ServerStream
}

// RegisterIngestorServer registers the server
func RegisterIngestorServer(s *grpc.Server, srv IngestorServer) {
	s.RegisterService(&_Ingestor_serviceDesc, srv)
}

// _Ingestor_serviceDesc is a mock descriptor
var _Ingestor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ingestion.Ingestor",
	HandlerType: (*IngestorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendEvent",
			Handler:    nil, // In a real mock we'd need the handler, but this allows compiling 'RegisterService' calls usually
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamEvents",
			Handler:       nil,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "ingestion.proto",
}
