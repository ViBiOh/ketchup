package notifier

import "github.com/prometheus/client_golang/prometheus"

func configurePrometheus() (prometheus.Gatherer, prometheus.Gauge, prometheus.Gauge) {
	namespace := "ketchup"
	promRegistry := prometheus.NewRegistry()

	releasesMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "releases",
	})
	promRegistry.MustRegister(releasesMetrics)

	notificationMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "notification",
	})
	promRegistry.MustRegister(notificationMetrics)

	return promRegistry, releasesMetrics, notificationMetrics
}
