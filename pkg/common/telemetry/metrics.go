package telemetry

import (
	"context"
)

func Count(ctx context.Context, component, entity, action string) {
	var increment int64
	increment = 1

	meter := InitiateMeterProvider(ctx)

	metric_name := FormatLabel(component, entity, action)
	updown_counter, err := meter.SyncInt64().UpDownCounter(
		metric_name,
	)
	if err != nil {
		panic(err)
	}

	if action == Remove {
		increment = -1
	}

	updown_counter.Add(ctx, increment)
}
