package panurge

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type HealthcheckFunc func(ctx context.Context) error

func StandardInternalMux(
	logger *logrus.Logger, test HealthcheckFunc,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Prometheus metrics
	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/health", HealthcheckHandler(logger, test))

	// PPROF endpoints for live profiles
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))

	// Expose public debug variables
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

func HealthcheckHandler(
	logger *logrus.Logger, test HealthcheckFunc,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		err := test(r.Context())
		if err != nil {
			logger.WithError(err).Error("healthcheck failed")

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintln(w, `{"status": "fail"}`)
			return
		}

		_, _ = fmt.Fprintln(w, `{"status": "pass"}`)
	})
}

func NoopHealthcheck(ctx context.Context) error {
	return nil
}
