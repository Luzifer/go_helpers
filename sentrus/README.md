# Sentrus

Sentrus is a small Logrus hook which sends selected entries as Sentry-compatible
issue events. It implements only the envelope event path and does not depend on
the Sentry SDK.

```go
hook, err := sentrus.NewHook(
	sentrus.WithDSN(os.Getenv("SENTRY_DSN")),
	sentrus.WithEnvironment("production"),
	sentrus.WithRelease("my-service@1.2.3"),
)
if err != nil {
	return err
}

logrus.AddHook(hook)
```

By default, Error, Fatal, and Panic entries are submitted. `WithLevels` replaces
that exact list and accepts every valid Logrus level.

## Delivery behavior

Logging never waits for the remote service. Events enter a bounded in-memory
queue; an event is dropped when that queue is full. `Dropped` reports queue
drops, while `Failed` reports events which could not be encoded or delivered.
Both counters are safe to read concurrently.

Delivery is best effort. A process exit can discard queued events unless the
application explicitly flushes them. Register a bounded exit handler when using
`logrus.Fatal`:

```go
logrus.RegisterExitHandler(func() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = hook.Flush(ctx)
})
```

For graceful shutdown and panic unwinding in the same goroutine, register an
equivalent deferred flush after creating the hook:

```go
defer func() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = hook.Flush(ctx)
}()
```

`Flush` waits for events accepted before the call, but does not turn failed
deliveries into successes and does not stop the worker.

Log fields are sent as event extras without redaction. Do not put secrets into
fields submitted through this hook.
