package fasthttpprom

import (
	zeroconv "github.com/savsgio/gotils/strconv"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

const (
	reqsName    = "requests_total"
	latencyName = "request_duration_milliseconds"
)

var (
	dflBuckets     = []float64{300, 1200, 5000}
	DefaultBuckets = []float64{.005, .01, .02, 0.04, .06, 0.08, .1, 0.15, .25, 0.4, .6, .8, 1, 1.5, 2, 3, 5}
)

//func (p *Prometheus) registerHisto(name, prefix string, buckets ...float64) {
//
//
//}
//
//func (p *Prometheus) registerCount(name, prefix string) {
//	p.reqs = prometheus.NewCounterVec(
//		prometheus.CounterOpts{
//			Name:        addPrefixIfNeeded(reqsName, prefix),
//			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
//			ConstLabels: prometheus.Labels{"service": name},
//		},
//		[]string{"code", "method", "path"},
//	)
//}

// NewMiddleware is
func NewMiddleware(name, prefix string, buckets ...float64) func(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	var p Prometheus
	p.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        addPrefixIfNeeded(reqsName, prefix),
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.reqs)

	if len(buckets) == 0 {
		buckets = dflBuckets
	}
	p.latency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        addPrefixIfNeeded(latencyName, prefix),
			Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
			Buckets:     buckets,
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.latency)
	return p.handler
}

func (p Prometheus) handler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		next(ctx)
		p.reqs.WithLabelValues(strconv.Itoa(ctx.Response.StatusCode()), zeroconv.B2S(ctx.Method()), zeroconv.B2S(ctx.URI().Path())).Inc()
		p.latency.WithLabelValues(strconv.Itoa(ctx.Response.StatusCode()),
			zeroconv.B2S(ctx.Method()), zeroconv.B2S(ctx.URI().Path())).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
	}
}

// since prometheus/client_golang use net/http we need this net/http adapter for fasthttp

func PrometheusHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
}
func addPrefixIfNeeded(name, prefix string) string {
	if prefix == "" {
		return name
	}
	return prefix + "_" + name
}
