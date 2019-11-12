package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kazeburo/consul-service-has-ip/accesslog"
	"github.com/kazeburo/consul-service-has-ip/consulclient"
	"go.uber.org/zap"

	flags "github.com/jessevdk/go-flags"
	ss "github.com/lestrrat/go-server-starter-listener"
)

// Version set in compile
var Version string

type cmdOpts struct {
	Listen            string        `short:"l" long:"listen" default:"0.0.0.0" description:"address to bind"`
	Port              string        `short:"p" long:"port" default:"3000" description:"Port number to bind"`
	Version           bool          `short:"v" long:"version" description:"Show version"`
	ReadTimeout       time.Duration `long:"read-timeout" default:"30s" description:"timeout of reading request"`
	WriteTimeout      time.Duration `long:"write-timeout" default:"90s" description:"timeout of writing response"`
	ShutdownTimeout   time.Duration `long:"shutdown-timeout" default:"10s" description:"Timeout to wait for all connections to be closed"`
	Timeout           time.Duration `long:"timeout" default:"30s" description:"timeout to reques to consul"`
	LogDir            string        `long:"access-log-dir" default:"" description:"directory to store logfiles"`
	LogRotate         int64         `long:"access-log-rotate" default:"30" description:"Number of day before remove logs"`
	ConsulAPIEndpoint string        `long:"consul-api-endpoint" description:"api endpoint of consul. required" required:"true"`
}

func wrapLogHandler(h http.Handler, logDir string, logRotate int64, logger *zap.Logger) http.Handler {
	al, err := accesslog.New(logDir, logRotate)
	if err != nil {
		logger.Fatal("could not init accesslog", zap.Error(err))
	}
	return al.WrapHandleFunc(h)
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}

func handleHasCheck(w http.ResponseWriter, r *http.Request, cc *consulclient.Client, logger *zap.Logger) {
	vars := mux.Vars(r)
	service := vars["service"]
	ip := vars["ip"]
	ok, err := cc.HasIP(r.Context(), service, ip)
	if err != nil {
		logger.Info("Error in consulclient", zap.Error(err))
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": 1, "messages": err.Error()})
		return
	}
	if !ok {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": 1, "messages": fmt.Sprintf("ip:%s is not in service:%s", ip, service)})
		return
	}
	w.Write([]byte(`{"error":0}` + "\n"))
}

func main() {
	opts := cmdOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf(`consul-service-include-jp %s
Compiler: %s %s
`,
			Version,
			runtime.Compiler,
			runtime.Version())
		os.Exit(0)
	}

	logger, _ := zap.NewProduction()
	cc := consulclient.New(opts.ConsulAPIEndpoint, opts.Timeout)

	m := mux.NewRouter()
	m.HandleFunc("/live", handleHello)
	m.HandleFunc("/", handleHello)
	m.HandleFunc("/has/{service:[A-Za-z0-9_-]+}/{ip:.+}", func(w http.ResponseWriter, r *http.Request) {
		handleHasCheck(w, r, cc, logger)
	})
	handler := wrapLogHandler(m, opts.LogDir, opts.LogRotate, logger)

	server := http.Server{
		Handler:      handler,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM)
		<-sigChan
		logger.Info("Signal received. Start to shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), opts.ShutdownTimeout)
		if es := server.Shutdown(ctx); es != nil {
			logger.Info("shutdown error", zap.Error(es))
		}
		cancel()
		close(idleConnsClosed)
		logger.Info("Waiting for all connections to be closed")
	}()

	l, err := ss.NewListener()
	if l == nil || err != nil {
		// Fallback if not running under Server::Starter
		l, err = net.Listen("tcp", fmt.Sprintf("%s:%s", opts.Listen, opts.Port))
		if err != nil {
			logger.Fatal("Failed to listen to port",
				zap.Error(err),
				zap.String("listen", opts.Listen),
				zap.String("porrt", opts.Port),
			)
		}
	}

	if err := server.Serve(l); err != http.ErrServerClosed {
		logger.Error("Error in Serve", zap.Error(err))
	}

	<-idleConnsClosed

}
