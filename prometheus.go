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

//var defaultMetricPath = "/metrics"

// ListenerHandler url label
//type ListenerHandler func(ctx *fasthttp.RequestCtx) string

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	//listenAddress string
	//MetricsPath   string
	//Handler       fasthttp.RequestHandler
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

//// NewPrometheus generates a new set of metrics with a certain subsystem name
//func NewPrometheus(subsystem string) *Prometheus {
//	p := &Prometheus{
//		MetricsPath: defaultMetricPath,
//	}
//	p.registerMetrics(subsystem)
//
//	return p
//}
//
//// SetListenAddress for exposing metrics on address. If not set, it will be exposed at the
//// same address of api that is being used
//func (p *Prometheus) SetListenAddress(address string) {
//	p.listenAddress = address
//	if p.listenAddress != "" {
//		p.router = router.New()
//	}
//}
//
//// SetListenAddressWithRouter for using a separate router to expose metrics. (this keeps things like GET /metrics out of
//// your content's access log).
//func (p *Prometheus) SetListenAddressWithRouter(listenAddress string, r *router.Router) {
//	p.listenAddress = listenAddress
//	if len(p.listenAddress) > 0 {
//		p.router = r
//	}
//}
//
//// SetMetricsPath set metrics paths for Custom path
//func (p *Prometheus) SetMetricsPath(r *router.Router) {
//	if p.listenAddress != "" {
//		r.GET(p.MetricsPath, prometheusHandler())
//		p.runServer()
//	} else {
//		r.GET(p.MetricsPath, prometheusHandler())
//	}
//}
//
//func (p *Prometheus) runServer() {
//	if p.listenAddress != "" {
//		go fasthttp.ListenAndServe(p.listenAddress, p.router.Handler)
//	}
//}

const (
	reqsName    = "requests_total"
	latencyName = "request_duration_milliseconds"
)

var (
	dflBuckets     = []float64{300, 1200, 5000}
	DefaultBuckets = []float64{.005, .01, .02, 0.04, .06, 0.08, .1, 0.15, .25, 0.4, .6, .8, 1, 1.5, 2, 3, 5}
)

func (p *Prometheus) registerHisto(name, prefix string, buckets ...float64) {

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
}

func (p *Prometheus) registerCount(name, prefix string) {
	p.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        addPrefixIfNeeded(reqsName, prefix),
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.reqs)
}

// NewMiddleware is
func NewMiddleware(name, prefix string, buckets ...float64) func(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			p := &Prometheus{}
			p.registerCount(name, prefix)
			p.registerHisto(name, prefix, buckets...)
			start := time.Now()
			next(ctx)
			p.reqs.WithLabelValues(strconv.Itoa(ctx.Response.StatusCode()), zeroconv.B2S(ctx.Method()), zeroconv.B2S(ctx.URI().Path())).Inc()
			p.latency.WithLabelValues(strconv.Itoa(ctx.Response.StatusCode()),
				zeroconv.B2S(ctx.Method()), zeroconv.B2S(ctx.URI().Path())).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)

			//	uri := string(ctx.Request.URI().Path())
			//	if uri == p.MetricsPath {
			//		// next
			//		p.router.Handler(ctx)
			//		return
			//	}
			//	start := time.Now()
			//	// next
			//	p.router.Handler(ctx)
			//
			//	status := strconv.Itoa(ctx.Response.StatusCode())
			//	elapsed := float64(time.Since(start)) / float64(time.Second)
			//	// get route pattern of url
			//	routeList := p.router.List()
			//	paths, ok := routeList[string(ctx.Request.Header.Method())]
			//	handler, _ := p.router.Lookup(string(ctx.Request.Header.Method()), uri, ctx)
			//	if ok {
			//		for _, v := range paths {
			//			tmp, _ := p.router.Lookup(string(ctx.Request.Header.Method()), v, ctx)
			//			if fmt.Sprintf("%v", tmp) == fmt.Sprintf("%v", handler) {
			//				uri = v
			//				break
			//			}
			//		}
			//	}
			//	ep := string(ctx.Method()) + "_" + uri
			//	ob, err := p.reqDur.GetMetricWithLabelValues(status, ep)
			//	if err != nil {
			//		log.Printf("Fail to GetMetricWithLabelValues: %s\n", err)
			//		return
			//	}
			//	ob.Observe(elapsed)
		}
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
