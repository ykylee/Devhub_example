package httpapi

import "github.com/gin-gonic/gin"

// testRouter wraps NewRouter with AuthDevFallback enabled so existing handler tests can keep calling endpoints without an Authorization header. Tests that exercise the authentication or authorization policy itself should call NewRouter directly with explicit RouterConfig fields.
func testRouter(cfg RouterConfig) *gin.Engine {
	cfg.AuthDevFallback = true
	return NewRouter(cfg)
}
