package notifier

import "github.com/prometheus/client_golang/prometheus"

func configurePrometheus() (prometheus.Gatherer, *prometheus.GaugeVec) {
	namespace := "ketchup"
	promRegistry := prometheus.NewRegistry()

	metrics := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "metrics",
	}, []string{"item"})
	promRegistry.MustRegister(metrics)

	return promRegistry, metrics
}
