// Package app wires configuration, adapters, services and the HTTP server, and
// owns the process lifecycle. It contains no business logic.
package app

import (
	"context"
	"errors"
	"net/http"

	"modelrail/internal/app/config"
	applicationinference "modelrail/internal/application/inference"
	"modelrail/internal/constant"
	httpcontroller "modelrail/internal/controller/http"
	domaininference "modelrail/internal/domain/inference"
	providermock "modelrail/internal/infrastructure/provider/mock"
	"modelrail/internal/logger"
)

// Run loads the configuration, wires the application and serves until ctx is
// cancelled. A startup failure or an unexpected server error is returned.
func Run(ctx context.Context, configPath string) (err error) {
	var cfg config.Config
	if cfg, err = config.Load(configPath); err != nil {
		return
	}

	var log *logger.Logger
	if log, err = logger.New(logger.Config{
		Level:    cfg.Logging.Level,
		Encoding: cfg.Logging.Encoding,
	}); err != nil {
		return
	}
	logger.SetGlobal(log)
	defer func() {
		// Sync errors are platform-specific and intentionally ignored.
		_ = log.Sync()
	}()

	provider := providermock.New(providermock.Config{
		ProviderName: cfg.Providers.Mock.Name,
		ResponseText: cfg.Providers.Mock.ResponseText,
		ChunkSize:    cfg.Providers.Mock.ChunkSize,
		Latency:      cfg.Providers.Mock.Latency.Std(),
		TTFT:         cfg.Providers.Mock.TTFT.Std(),
		ChunkDelay:   cfg.Providers.Mock.ChunkDelay.Std(),
	})

	inferenceDomainService := domaininference.NewInferenceService(provider)
	inferenceApplicationService := applicationinference.NewInferenceService(inferenceDomainService)
	router := httpcontroller.NewRouter(httpcontroller.Deps{Inference: inferenceApplicationService})

	server := &http.Server{
		Addr:              cfg.Server.PublicAddr,
		Handler:           router,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout.Std(),
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	log.Info("gateway started",
		logger.String(constant.FieldAddress, cfg.Server.PublicAddr),
		logger.String(constant.FieldEnvironment, cfg.App.Environment),
	)

	select {
	case <-ctx.Done():
		log.Info("shutdown requested")
	case serveErr := <-serverErr:
		if !errors.Is(serveErr, http.ErrServerClosed) {
			err = serveErr
			return
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownGrace.Std())
	defer cancel()
	if err = server.Shutdown(shutdownCtx); err != nil {
		return
	}
	if serveErr := <-serverErr; serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		err = serveErr
	}
	log.Info("gateway stopped")
	return
}
