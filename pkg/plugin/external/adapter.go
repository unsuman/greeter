package external

import (
	"context"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unsuman/greeter/pkg/greetings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "github.com/unsuman/greeter/pkg/plugin/proto"
)

// Server adapts a greetings.Plugin to serve over gRPC
type Server struct {
	pb.UnimplementedGreeterServiceServer
	plugin greetings.Plugin
	logger *logrus.Logger
}

// Run runs a plugin as a standalone executable
func Run(plugin greetings.Plugin) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.Info("Starting plugin: ", plugin.Name())

	if err := plugin.Init(); err != nil {
		logger.Fatalf("Failed to initialize plugin: %v", err)
	}

	defer func() {
		if err := plugin.Close(); err != nil {
			logger.Errorf("Error during plugin cleanup: %v", err)
		}
	}()

	conn := &PipeConn{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}

	listener := NewPipeListener(conn)

	kaProps := keepalive.ServerParameters{
		Time:    5 * time.Second,
		Timeout: 10 * time.Second,
	}

	kaPolicy := keepalive.EnforcementPolicy{
		MinTime:             1 * time.Second,
		PermitWithoutStream: true,
	}

	server := grpc.NewServer(
		grpc.KeepaliveParams(kaProps),
		grpc.KeepaliveEnforcementPolicy(kaPolicy),
	)

	pb.RegisterGreeterServiceServer(server, &Server{
		plugin: plugin,
		logger: logger,
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		logger.Info("Received shutdown signal, stopping server...")
		server.Stop()
		os.Exit(0)
	}()

	logger.Info("Server starting...")
	if err := server.Serve(listener); err != nil {
		logger.Fatalf("Failed to serve: %v", err)
	}
}

// Implement the gRPC service methods
func (s *Server) Hello(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	s.logger.Debug("Received Hello request")
	return &pb.GreetingResponse{Message: s.plugin.Hello()}, nil
}

func (s *Server) GoodMorning(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	s.logger.Debug("Received GoodMorning request")
	return &pb.GreetingResponse{Message: s.plugin.GoodMorning()}, nil
}

func (s *Server) GoodAfternoon(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	s.logger.Debug("Received GoodAfternoon request")
	return &pb.GreetingResponse{Message: s.plugin.GoodAfternoon()}, nil
}

func (s *Server) GoodNight(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	s.logger.Debug("Received GoodNight request")
	return &pb.GreetingResponse{Message: s.plugin.GoodNight()}, nil
}

func (s *Server) GoodBye(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	s.logger.Debug("Received GoodBye request")
	return &pb.GreetingResponse{Message: s.plugin.GoodBye()}, nil
}

// PipeConn implements the net.Conn interface over stdin/stdout
type PipeConn struct {
	Reader io.Reader
	Writer io.Writer
}

func (p *PipeConn) Read(b []byte) (n int, err error) {
	return p.Reader.Read(b)
}

func (p *PipeConn) Write(b []byte) (n int, err error) {
	return p.Writer.Write(b)
}

func (p *PipeConn) Close() error {
	return nil
}

func (p *PipeConn) LocalAddr() net.Addr {
	return pipeAddr{}
}

func (p *PipeConn) RemoteAddr() net.Addr {
	return pipeAddr{}
}

func (p *PipeConn) SetDeadline(t time.Time) error {
	return nil
}

func (p *PipeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (p *PipeConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

// PipeListener implements a simple net.Listener for stdin/stdout
type PipeListener struct {
	conn     net.Conn
	connSent bool
	mu       sync.Mutex
	closed   bool
}

func NewPipeListener(conn net.Conn) *PipeListener {
	return &PipeListener{
		conn:     conn,
		connSent: false,
		closed:   false,
	}
}

func (l *PipeListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil, net.ErrClosed
	}

	if l.connSent {
		select {}
	}

	l.connSent = true
	return l.conn, nil
}

func (l *PipeListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.closed {
		l.closed = true
	}
	return nil
}

func (l *PipeListener) Addr() net.Addr {
	return pipeAddr{}
}
