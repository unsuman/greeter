package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "github.com/unsuman/greeter/pkg/plugin/proto"
)

// HindiGreeter implements the Greeter interface in Hindi
type JapaneseGreeter struct {
	pb.UnimplementedGreeterServiceServer
}

func (g *JapaneseGreeter) Hello(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	log.Println("Received Hello request")
	return &pb.GreetingResponse{Message: "こんにちは! (Konnichiwa)"}, nil
}

func (g *JapaneseGreeter) GoodMorning(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	log.Println("Received GoodMorning request")
	return &pb.GreetingResponse{Message: "おはようございます! (Ohayou gozaimasu)"}, nil
}

func (g *JapaneseGreeter) GoodAfternoon(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	log.Println("Received GoodAfternoon request")
	return &pb.GreetingResponse{Message: "こんにちは! (Konnichiwa)"}, nil
}

func (g *JapaneseGreeter) GoodNight(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	log.Println("Received GoodNight request")
	return &pb.GreetingResponse{Message: "おやすみなさい! (Oyasumi nasai)"}, nil
}

func (g *JapaneseGreeter) GoodBye(ctx context.Context, empty *pb.Empty) (*pb.GreetingResponse, error) {
	log.Println("Received GoodBye request")
	return &pb.GreetingResponse{Message: "さようなら! (Sayounara)"}, nil
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
	return nil // Not supported for pipes
}

func (p *PipeConn) SetReadDeadline(t time.Time) error {
	return nil // Not supported for pipes
}

func (p *PipeConn) SetWriteDeadline(t time.Time) error {
	return nil // Not supported for pipes
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
		// Block forever once the connection is sent
		select {} // This will block indefinitely
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

func main() {
	log.Println("Starting Hindi greeter plugin")

	// Set up the connection over stdin/stdout
	conn := &PipeConn{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}

	// Create a listener that will return our connection
	listener := NewPipeListener(conn)

	// Configure keepalive settings
	kaProps := keepalive.ServerParameters{
		Time:    5 * time.Second,
		Timeout: 10 * time.Second,
	}

	kaPolicy := keepalive.EnforcementPolicy{
		MinTime:             1 * time.Second,
		PermitWithoutStream: true,
	}

	// Create the gRPC server with keepalive settings
	server := grpc.NewServer(
		grpc.KeepaliveParams(kaProps),
		grpc.KeepaliveEnforcementPolicy(kaPolicy),
	)

	// Register our greeting service
	pb.RegisterGreeterServiceServer(server, &JapaneseGreeter{})

	// Handle shutdown signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Received shutdown signal, stopping server...")
		server.Stop()
		os.Exit(0)
	}()

	// Start serving
	log.Println("Server starting...")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
