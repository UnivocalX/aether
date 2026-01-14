Got it — let’s focus only on the new terms/utilities you may want to introduce that are not yet implemented, and show short description, input/output, and suggested file/category.

---

1️⃣ Transformation / Stage utilities (stage.go)

Term Description Input → Output Category/File

Map Transform values in a stream (pure) stream of Envelope[T] → stream of Envelope[T] stage.go
Filter Pass values only if predicate is true stream of Envelope[T] → filtered stream stage.go
Tap / SideEffectStage Perform side-effects without changing values stream of Envelope[T] → same stream stage.go
Partition Split stream into two based on predicate stream → 2 streams stage.go
Retry Retry failed envelopes a number of times stream → stream stage.go
Batch / Window Group values into slices of N or time window stream → stream of slices stage.go
MergeSorted / Join Merge multiple streams with ordering or join multiple streams → single ordered stream stage.go

---

2️⃣ Terminal / Sink (sink.go)

Term Description Input → Output Category/File

Sink / ForEach Consume stream, perform action stream → none sink.go
Drain Consume stream just to prevent leaks stream → none sink.go
Collect Collect all envelopes into slice stream → []Envelope[T] sink.go
AbortOnError Stop pipeline on first error stream → stream (cancels context) sink.go

---

3️⃣ Pipeline / Orchestration (pipeline.go)

Term Description Input → Output Category/File

Pipeline Holds stages + context stages + source → orchestrated stream pipeline.go
Then / Chain Append a stage to pipeline pipeline + stage → extended pipeline pipeline.go
ThenConcurrent Append stage with fan-out workers pipeline + stage → extended pipeline pipeline.go
Run Execute pipeline pipeline + source → output stream pipeline.go
Collect Convenience to gather results pipeline → []Envelope[T] pipeline.go

---

✅ Summary / principle

stage.go → transformations, side-effects, filters, batching, retries

sink.go → terminal consumers, error policies, collection, drains

pipeline.go → fluent orchestration (Then, ThenConcurrent, Run, Collect)

Everything else you already implemented (Generator, FanIn, FanOut, OrDone, Tee, Bridge, Stage) stays in stream.go / stage.go.

---

If you want, I can make a final full “planned API map”, showing which primitives, stages, sinks, and pipeline methods you’ll have in the universe package — essentially a cheat sheet for implementation.

Do you want me to do that?
