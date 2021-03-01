package main

import (
	"flag"
	"io/ioutil"
	"net"
	"os"

	sdk "github.com/merefield/grpc-user-api/sdk"

	grpclog "google.golang.org/grpc/grpclog"

	"github.com/merefield/grpc-user-api/pkg/config"
	pgsql "github.com/merefield/grpc-user-api/pkg/postgres"

	//"github.com/merefield/grpc-user-api/pkg/role"

	// Static files
	_ "github.com/merefield/grpc-user-api/statik"
)

// TLSConfig points to the cert files needed for HTTPS
// type TLSConfig struct {
// 	enabled bool
// 	// CertFile is the path to the cert file
// 	CertFile string
// 	// KeyFile is the path to the key file
// 	KeyFile string
// }

// SecurityConfig provides configuration for SDK auth
// type SecurityConfig struct {
// 	// Role implementation
// 	Role role.RoleManager
// 	// Tls configuration
// 	TLS *TLSConfig
// 	// Authenticators per issuer. You can register multple authenticators
// 	// based on the "iss" string in the string. For example:
// 	// map[string]auth.Authenticator {
// 	//     "https://accounts.google.com": googleOidc,
// 	//     "openstorage-sdk-auth: selfSigned,
// 	// }
// 	Authenticators map[string]auth.Authenticator
// }

// ServerConfig provides the configuration to the SDK server
// type ServerConfig struct {
// 	// Net is the transport for gRPC: unix, tcp, etc.
// 	// For the gRPC Server. This value goes together with `Address`.
// 	Net string
// 	// Address is the port number or the unix domain socket path.
// 	// For the gRPC Server. This value goes together with `Net`.
// 	Address string
// 	// port is the port number at which remote SdkGrpcServer is running.Same
// 	// across cluster. Exampl: 9020
// 	port string
// 	// RestAdress is the port number. Example: 9110
// 	// For the gRPC REST Gateway.
// 	RestPort string
// 	// Unix domain socket for local communication. This socket
// 	// will be used by the REST Gateway to communicate with the gRPC server.
// 	// Only set for testing. Having a '%s' can be supported to use the
// 	// name of the driver as the driver name.
// 	Socket string
// 	// (optional) Location for audit log.
// 	// If not provided, it will go to /var/log/openstorage-audit.log
// 	AuditOutput io.Writer
// 	// (optional) Location of access log.
// 	// This is useful when authorization is not running.
// 	// If not provided, it will go to /var/log/openstorage-access.log
// 	AccessOutput io.Writer
// 	// (optional) The OpenStorage driver to use
// 	// DriverName string
// 	// (optional) Cluster interface
// 	// Cluster cluster.Cluster
// 	// AlertsFilterDeleter
// 	// AlertsFilterDeleter alerts.FilterDeleter
// 	// StoragePolicy Manager
// 	// StoragePolicy policy.PolicyManager
// 	// Security configuration
// 	Security *SecurityConfig
// 	// ServerExtensions allows you to extend the SDK gRPC server
// 	// with callback functions that are sequentially executed
// 	// at the end of Server.Start()
// 	//
// 	// To add your own service to the SDK gRPC server,
// 	// just append a function callback that registers it:
// 	//
// 	// s.config.ServerExtensions = append(s.config.ServerExtensions,
// 	// 		func(gs *grpc.Server) {
// 	//			api.RegisterCustomService(gs, customHandler)
// 	//		})
// 	//	GrpcServerExtensions []func(grpcServer *grpc.Server)

// 	// RestServerExtensions allows for extensions to be added
// 	// to the SDK Rest Gateway server.
// 	//
// 	// To add your own service to the SDK REST Server, simply add your handlers
// 	// to the RestSererExtensions slice. These handlers will be registered on the
// 	// REST Gateway http server.
// 	//	RestServerExtensions []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
// }

func main() {

	cfgPath := flag.String("p", "./conf.local.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)

	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	checkErr(log, err)

	db, err := pgsql.New(cfg.DB.Dev.PSN, cfg.DB.Dev.LogQueries, cfg.DB.Dev.TimeoutSeconds)

	checkErr(log, err)

	addr := "0.0.0.0:10000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// // Setup https if certs have been provided
	// opts := make([]grpc.ServerOption, 0)
	// if cfg.Security.TLS.Enabled != false {
	// 	creds, err := credentials.NewServerTLSFromFile(
	// 		cfg.Security.TLS.Certfile,
	// 		cfg.Security.TLS.Keyfile)
	// 	if err != nil {
	// 		//	return fmt.Errorf("Failed to create credentials from cert files: %v", err)
	// 		log.Fatalln("Failed to create credentials from cert files: %v", err)
	// 	}
	// 	opts = append(opts, grpc.Creds(creds))
	// 	log.Info("TLS enabled")
	// }

	// s := grpc.NewServer(
	// 	// TODO: Replace with your own certificate!
	// 	grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)),
	// )

	// Setup authentication and authorization using interceptors if auth is enabled
	// if len(cfg.Security.Authenticators) != 0 {
	// 	opts = append(opts, grpc.UnaryInterceptor(
	// 		grpc_middleware.ChainUnaryServer(
	// 			s.rwlockUnaryIntercepter,
	// 			grpc_auth.UnaryServerInterceptor(s.auth),
	// 			s.authorizationServerUnaryInterceptor,
	// 			s.loggerServerUnaryInterceptor,
	// 			grpc_prometheus.UnaryServerInterceptor,
	// 		)))
	// 	opts = append(opts, grpc.StreamInterceptor(
	// 		grpc_middleware.ChainStreamServer(
	// 			s.rwlockStreamIntercepter,
	// 			grpc_auth.StreamServerInterceptor(s.auth),
	// 			s.authorizationServerStreamInterceptor,
	// 			s.loggerServerStreamInterceptor,
	// 			grpc_prometheus.StreamServerInterceptor,
	// 		)))
	// } else {
	// 	opts = append(opts, grpc.UnaryInterceptor(
	// 		grpc_middleware.ChainUnaryServer(
	// 			s.rwlockUnaryIntercepter,
	// 			s.loggerServerUnaryInterceptor,
	// 			grpc_prometheus.UnaryServerInterceptor,
	// 		)))
	// 	opts = append(opts, grpc.StreamInterceptor(
	// 		grpc_middleware.ChainStreamServer(
	// 			s.rwlockStreamIntercepter,
	// 			s.loggerServerStreamInterceptor,
	// 			grpc_prometheus.StreamServerInterceptor,
	// 		)))
	// }

	// Create a gRPC server on a unix domain socket
	restServer, err := sdk.New(&cfg)
	if err != nil {
		return nil, err
	}

	restServer.Start()

	//pbAPI.RegisterUserServiceServer(s, server.New(db, cfg))

	// Serve gRPC Server
	// log.Info("Serving gRPC on https://", addr)
	// go func() {
	// 	log.Fatal(s.Serve(lis))
	// }()

	// err = gateway.Run("dns:///" + addr)
	// log.Fatal(err)
}

func checkErr(log grpclog.LoggerV2, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
