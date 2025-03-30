package plugin

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/unsuman/greeter/pkg/plugin/proto"
)

// GRPCClient is a client for communicating with a gRPC server over stdin/stdout
type GRPCClient struct {
	Stdin      io.WriteCloser
	Stdout     io.ReadCloser
	Conn       *grpc.ClientConn
	GreeterSvc pb.GreeterServiceClient
	logger     *logrus.Entry
}

// NewGRPCClient creates a new GRPCClient
func NewGRPCClient(stdin io.WriteCloser, stdout io.ReadCloser, logger *logrus.Entry) (*GRPCClient, error) {
	// Create a pipe that connects stdin/stdout to a gRPC client
	clientConn := newPipeConn(stdin, stdout)

	// Create a gRPC client connection
	conn, err := grpc.NewClient("pipe",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return clientConn, nil
		}),
	)
	if err != nil {
		logger.Errorf("Failed to create gRPC client connection: %v", err)
		return nil, err
	}

	// Create the service client
	greeterClient := pb.NewGreeterServiceClient(conn)

	return &GRPCClient{
		Stdin:      stdin,
		Stdout:     stdout,
		Conn:       conn,
		GreeterSvc: greeterClient,
		logger:     logger,
	}, nil
}

// Close closes the client connection
func (c *GRPCClient) Close() error {
	return c.Conn.Close()
}

// GetGreeting calls the appropriate gRPC method based on greeting type
func (c *GRPCClient) GetGreeting(ctx context.Context, greetingType string) (string, error) {
	var response *pb.GreetingResponse
	var err error

	// Set a timeout for the RPC call
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch greetingType {
	case "hello":
		response, err = c.GreeterSvc.Hello(ctx, &pb.Empty{})
	case "goodmorning":
		response, err = c.GreeterSvc.GoodMorning(ctx, &pb.Empty{})
	case "goodafternoon":
		response, err = c.GreeterSvc.GoodAfternoon(ctx, &pb.Empty{})
	case "goodnight":
		response, err = c.GreeterSvc.GoodNight(ctx, &pb.Empty{})
	case "goodbye":
		response, err = c.GreeterSvc.GoodBye(ctx, &pb.Empty{})
	default:
		c.logger.Errorf("Unknown greeting type: %s", greetingType)
		return "", ErrUnknownGreetingType
	}

	if err != nil {
		c.logger.Errorf("gRPC call failed: %v", err)
		return "", err
	}

	return response.Message, nil
}

// PipeConn implements net.Conn over stdin/stdout
type PipeConn struct {
	reader io.Reader
	writer io.Writer
}

func newPipeConn(writer io.WriteCloser, reader io.ReadCloser) *PipeConn {
	return &PipeConn{
		reader: reader,
		writer: writer,
	}
}

func (c *PipeConn) Read(b []byte) (n int, err error)   { return c.reader.Read(b) }
func (c *PipeConn) Write(b []byte) (n int, err error)  { return c.writer.Write(b) }
func (c *PipeConn) Close() error                       { return nil }
func (c *PipeConn) LocalAddr() net.Addr                { return pipeAddr{} }
func (c *PipeConn) RemoteAddr() net.Addr               { return pipeAddr{} }
func (c *PipeConn) SetDeadline(_ time.Time) error      { return nil }
func (c *PipeConn) SetReadDeadline(_ time.Time) error  { return nil }
func (c *PipeConn) SetWriteDeadline(_ time.Time) error { return nil }

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

// Errors
var (
	ErrUnknownGreetingType = grpc.Errorf(codes.InvalidArgument, "unknown greeting type")
)
