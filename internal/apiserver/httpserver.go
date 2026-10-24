package apiserver

import (
	"context"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
	genericmw "github.com/onexstack/onexstack/pkg/middleware/gin"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/onexdemo/iamv2/internal/pkg/errno"
	mw "github.com/onexdemo/iamv2/internal/pkg/middleware/gin"
)

// ginServer implements the server.Server interface using the Gin framework.
type ginServer struct {
	srv server.Server
}

var _ server.Server = (*ginServer)(nil)

// NewGinServer initializes and returns a new HTTP server based on Gin.
func (c *ServerConfig) NewGinServer() (server.Server, error) {
	engine := gin.New()

	engine.Use(
		gin.Recovery(),
		mw.RequestIDMiddleware(),
		mw.NoCache,
		mw.Cors(mw.DefaultCorsConfig()),
		mw.Secure,
		otelgin.Middleware(serviceName, otelgin.WithFilter(shouldRecordTelemetry)),
		genericmw.Observability(),
		mw.Context(),
	)

	c.InstallRESTAPI(engine)

	httpsrv := server.NewHTTPServer(c.InsecureServingOptions, c.SecureServingOptions, engine)

	return &ginServer{srv: httpsrv}, nil
}

// InstallRESTAPI registers every HTTP route on the engine.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	InstallGenericAPI(engine)
	c.InstallAuthAPI(engine)

	v1 := engine.Group("/v1")
	c.Handler.ApplyTo(v1, mw.Authn(c.Authenticator))
}

// InstallAuthAPI registers login, logout and token-refresh endpoints. They are
// exposed both at the root (legacy iam) and under /v1 (OneX convention).
func (c *ServerConfig) InstallAuthAPI(engine *gin.Engine) {
	authn := mw.Authn(c.Authenticator)

	engine.POST("/login", c.Handler.Login)
	engine.POST("/logout", c.Handler.Logout)
	engine.POST("/refresh", authn, c.Handler.RefreshToken)

	engine.POST("/v1/login", c.Handler.Login)
	engine.POST("/v1/logout", c.Handler.Logout)
	engine.POST("/v1/refresh", authn, c.Handler.RefreshToken)
}

// InstallGenericAPI registers non-business routes such as pprof, metrics,
// health checks and the 404 handler.
func InstallGenericAPI(engine *gin.Engine) {
	pprof.Register(engine)

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{EnableOpenMetrics: true},
	)))

	engine.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, nil, errno.ErrPageNotFound)
	})
}

// Run starts the Gin server and returns when it stops.
func (s *ginServer) Run(ctx context.Context) error {
	return s.srv.Run(ctx)
}

// GracefulStop gracefully shuts down the server.
func (s *ginServer) GracefulStop(ctx context.Context) error {
	return s.srv.GracefulStop(ctx)
}

func shouldRecordTelemetry(r *http.Request) bool {
	return r.URL.Path != "/metrics"
}
