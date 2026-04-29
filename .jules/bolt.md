## 2024-04-29 - Removed Default Gin Logger for Backend Core
**Learning:** `gin.Default()` automatically attaches a logger middleware that writes to standard out on every request. This I/O operation can bottleneck high-throughput endpoints like health checks.
**Action:** Use `gin.New()` with `gin.Use(gin.Recovery())` instead of `gin.Default()` in production, and explicitly log only what's necessary to save CPU and I/O overhead.
