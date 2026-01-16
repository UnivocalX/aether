Other useful operators you might consider:

Take / TakeWhile

Limit the stream to a certain number of items or until a condition is false.

Use: throttling, sampling, or early exit.

Skip / SkipWhile

Ignore the first N items or until a condition is false.

Use: ignore warm-up items or irrelevant data.

FlatMap / Expand

Map each item to multiple items (or a stream) and flatten the results.

Use: when one item generates many downstream items.

Buffer / Window

Collect items in batches of N or by time window.

Use: batch processing, efficient I/O, or rate-limited calls.

Distinct / Unique

Remove duplicates from the stream.

Use: deduplication of events.

Retry / Recover

Reattempt failed items or provide fallback.

Use: robustness in unreliable pipelines (network requests, DB).

Throttle / RateLimit

Control the flow rate of the stream.

Use: prevent overload downstream or API rate limits.