package server

import (
	"fmt"
	pb "github.com/KYVENetwork/kyve-rdk/runtime/tendermint-bsync-go/proto/kyverdk/runtime/v1"
	"github.com/KYVENetwork/kyve-rdk/runtime/tendermint-bsync-go/utils"
	"google.golang.org/grpc"
	"net"
)

const (
	RuntimeName    = "@kyvejs/tendermint-bsync"
	RuntimeVersion = "1.1.7"
	host           = "0.0.0.0"
	maxMessageSize = 2 * 1024 * 1024 * 1024 // 2 GB
)

var logger = utils.Logger()

func StartServer(port int32, debug bool) {
	// Initialize the gRPC server and listen on port 50051
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		logger.Fatal().Msg(fmt.Sprintf("Failed to listen on port %d: %s", port, err))
		return
	}

	// Create a new gRPC server instance
	server := grpc.NewServer(grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize))

	// Register the Tendermint service with the gRPC server
	pb.RegisterRuntimeServiceServer(server, &TendermintBsyncGoServer{debug: debug, logger: logger})

	// Start serving incoming connections
	logger.Info().Msg(fmt.Sprintf("Server is running on http://%s:%d...", host, port))
	logger.Info().Msg("Press Ctrl + C to exit.")

	if err := server.Serve(listener); err != nil {
		logger.Fatal().Msg(fmt.Sprintf("Failed to serve gRPC server: %s", err))
	}
}
