package reducer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"gitlab.com/mataelang/report-command-service/internal/pb"
)

type Metric map[string]interface{}

func CleanseMetric(metric *pb.Metric) Metric {
	cleansedMetric := Metric{
		// "snort_timestamp":   metric.SnortTimestamp,
		"snort_src_address": metric.SnortSrcAddress,
		"snort_dst_address": metric.SnortDstAddress,
		"snort_src_port":    metric.SnortSrcPort,
		"snort_dst_port":    metric.SnortDstPort,
	}
	return cleansedMetric
}

func hashMetric(metric Metric) string {
	hasher := sha256.New()
	keys := make([]string, 0, len(metric))
	for key := range metric {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		hasher.Write([]byte(key))
		switch v := metric[key].(type) {
		case string:
			hasher.Write([]byte(v))
		case *string:
			if v != nil {
				hasher.Write([]byte(*v))
			}
		case int, int8, int16, int32, int64:
			hasher.Write([]byte(fmt.Sprintf("%d", v)))
		case uint, uint8, uint16, uint32, uint64:
			hasher.Write([]byte(fmt.Sprintf("%d", v)))
		case float32, float64:
			hasher.Write([]byte(fmt.Sprintf("%f", v)))
		default:
			hasher.Write([]byte(fmt.Sprintf("%v", v)))
		}
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func CountMetrics(metrics []Metric) []Metric {
	counts := make(map[string]int)
	cleansedMetrics := make(map[string]Metric)

	for _, metric := range metrics {
		metricHash := hashMetric(metric)
		counts[metricHash]++
		cleansedMetrics[metricHash] = metric
	}

	var result []Metric
	for metricHash, count := range counts {
		metric := cleansedMetrics[metricHash]
		metric["count"] = count
		result = append(result, metric)
	}

	return result
}

func CreateOutputData(consumedMessage *pb.SensorEvent, countedMetrics []Metric) map[string]interface{} {
	return map[string]interface{}{
		"metrics":              countedMetrics,
		"sensor_id":            consumedMessage.SensorId,
		"event_metrics_count":  consumedMessage.EventMetricsCount,
		"snort_classification": consumedMessage.SnortClassification,
		"snort_message":        consumedMessage.SnortMessage,
		"snort_priority":       consumedMessage.SnortPriority,
	}
}
