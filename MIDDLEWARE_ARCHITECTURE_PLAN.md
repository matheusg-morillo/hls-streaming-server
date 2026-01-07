# Plano de Arquitetura de Middlewares - Matflix HLS Server

**Data:** 2026-01-06
**Objetivo:** Melhorar a arquitetura de middlewares para torná-los auto-contidos, composíveis e idiomáticos em Go

---

## 📊 Análise da Arquitetura Atual

### Implementação Existente

**Arquivo:** `src/middleware/middleware.go`

```go
func Use(before http.HandlerFunc, after http.HandlerFunc, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if before != nil {
            before(w, r)
        }

        if r.Method != http.MethodOptions {
            next.ServeHTTP(w, r)
        }

        if after != nil {
            after(w, r)
        }
    })
}
```

**Uso Atual:**
```go
Routes = map[string]http.Handler{
    "/health": middleware.Use(
        middleware.WithInboundLogging(),  // before
        middleware.WithOutgoingLogging(), // after
        healthHandler(),                  // next
    ),
}
```

### ❌ Problemas Identificados

1. **Não é idiomático em Go**
   - Go favorece `func(http.Handler) http.Handler`
   - Padrão diferente de chi, gorilla, echo, negroni

2. **Falta de composabilidade**
   - Difícil compor múltiplos middlewares
   - Não permite pipelines fluentes
   - Não reutilizável com libs de terceiros

3. **Runtime overhead**
   - Checks de `nil` a cada request
   - Lógica especial para `OPTIONS` (não deveria estar aqui)

4. **After sempre executa**
   - Mesmo se houver erros ou panics
   - Não captura status code da resposta

5. **Arquivos estáticos não seguem padrão**
   - `UseStaticFiles()` é diferente de `Use()`
   - Difícil aplicar mesma pipeline de middlewares
   - Hardcoded no `server.go`

---

## ✅ Proposta de Arquitetura Nova

### 1. Tipo Base de Middleware (Padrão Idiomático)

```go
// types.go
package middleware

import "net/http"

// Middleware é uma função que recebe um Handler e retorna outro Handler wrapeado
// Este é o padrão idiomático em Go para middlewares (closures)
type Middleware func(http.Handler) http.Handler
```

**Por que este padrão?**
- Usado por chi, gorilla/mux, echo, negroni
- Compatível com `net/http` stdlib
- Zero overhead (composição em init time)
- Type-safe e composível

---

### 2. Composição de Middlewares

```go
// Chain compõe múltiplos middlewares em uma única pipeline
// Os middlewares são executados na ordem fornecida (esquerda para direita)
// Exemplo: Chain(Logger, CORS, Auth) executa Logger → CORS → Auth → handler
func Chain(middlewares ...Middleware) Middleware {
    return func(final http.Handler) http.Handler {
        // Aplicar middlewares em ordem reversa para execução correta
        for i := len(middlewares) - 1; i >= 0; i-- {
            final = middlewares[i](final)
        }
        return final
    }
}

// Then é um helper para aplicar um middleware a um handler específico
func Then(mw Middleware, handler http.Handler) http.Handler {
    return mw(handler)
}

// ThenFunc é um helper para aplicar um middleware a um HandlerFunc
func ThenFunc(mw Middleware, handler http.HandlerFunc) http.Handler {
    return mw(handler)
}
```

**Exemplo de uso:**
```go
// Opção 1: Com Chain
handler := Chain(
    CORS(),
    Logger(),
    RateLimit(),
)(finalHandler)

// Opção 2: Manual (mais explícito)
handler := CORS()(
    Logger()(
        RateLimit()(finalHandler)))

// Opção 3: Com Then helper
handler := Then(Chain(CORS(), Logger()), finalHandler)
```

---

### 3. ResponseWriter Wrapper (Para Capturar Status Code)

```go
// response.go
package middleware

import "net/http"

// ResponseWriter wraps http.ResponseWriter para capturar status code e bytes escritos
type ResponseWriter struct {
    http.ResponseWriter
    StatusCode   int
    BytesWritten int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
    return &ResponseWriter{
        ResponseWriter: w,
        StatusCode:     http.StatusOK, // default
    }
}

func (rw *ResponseWriter) WriteHeader(code int) {
    rw.StatusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
    if rw.StatusCode == 0 {
        rw.StatusCode = http.StatusOK
    }
    n, err := rw.ResponseWriter.Write(b)
    rw.BytesWritten += n
    return n, err
}

// Unwrap retorna o ResponseWriter original (para http.Hijacker, http.Flusher, etc)
func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
    return rw.ResponseWriter
}
```

**Uso em middlewares:**
```go
func LoggingWithStatus() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            wrapped := NewResponseWriter(w)

            start := time.Now()
            next.ServeHTTP(wrapped, r)
            duration := time.Since(start)

            // Agora temos acesso ao status code!
            log.Printf("%s %s -> %d (%d bytes) [%v]",
                r.Method, r.URL.Path,
                wrapped.StatusCode, wrapped.BytesWritten,
                duration)
        })
    }
}
```

---

### 4. Refatoração dos Middlewares Existentes

#### CORS Middleware

```go
// cors.go
package middleware

import (
    "log"
    "net/http"
)

// CORS adiciona headers CORS e trata preflight requests (OPTIONS)
func CORS() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Println("Applying CORS headers")

            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "*")
            w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")

            // Preflight request - retorna sem chamar next
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// CORSWithConfig permite customização
type CORSConfig struct {
    AllowOrigins []string
    AllowMethods []string
    AllowHeaders []string
}

func CORSWithConfig(config CORSConfig) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Implementação customizada...
            next.ServeHTTP(w, r)
        })
    }
}
```

#### Logging Middleware

```go
// logging.go
package middleware

import (
    "log"
    "net/http"
    "time"
)

// Logger cria um middleware que loga requests e responses
func Logger() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Before: log inbound request
            log.Printf("→ %s %s", r.Method, r.URL.Path)

            // Wrap response writer para capturar status
            wrapped := NewResponseWriter(w)
            next.ServeHTTP(wrapped, r)

            // After: log outbound response com status e duração
            duration := time.Since(start)
            log.Printf("← %s %s -> %d [%v]",
                r.Method, r.URL.Path, wrapped.StatusCode, duration)
        })
    }
}

// SimpleLogger apenas loga sem capturar status (mais rápido)
func SimpleLogger() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Printf("%s %s", r.Method, r.URL.Path)
            next.ServeHTTP(w, r)
        })
    }
}

// LoggerWithFormat permite customização do formato
func LoggerWithFormat(format string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            wrapped := NewResponseWriter(w)

            next.ServeHTTP(wrapped, r)

            // Use format string customizado
            log.Printf(format, r.Method, r.URL.Path, wrapped.StatusCode, time.Since(start))
        })
    }
}
```

#### HLS Headers Middleware

```go
// hls.go
package middleware

import (
    "net/http"
    "strings"
)

// HLSHeaders adiciona headers necessários para HLS streaming
func HLSHeaders() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Accept-Ranges é crítico para adaptive bitrate (ABR)
            w.Header().Set("Accept-Ranges", "bytes")

            next.ServeHTTP(w, r)
        })
    }
}

// HLSCacheControl aplica cache strategy diferente para .m3u8 vs .ts
func HLSCacheControl() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            path := r.URL.Path

            if strings.HasSuffix(path, ".m3u8") {
                // Playlists mudam constantemente - NO CACHE
                w.Header().Set("Cache-Control", "no-cache")
            } else if strings.HasSuffix(path, ".ts") {
                // Segmentos são imutáveis - CACHE AGRESSIVO
                w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
            }

            next.ServeHTTP(w, r)
        })
    }
}

// HLSRangeLogging loga range requests (útil para debug ABR)
func HLSRangeLogging() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
                log.Printf("Range request: %s for %s", rangeHeader, r.URL.Path)
            }

            next.ServeHTTP(w, r)
        })
    }
}

// HLSComplete combina todos os middlewares HLS
func HLSComplete() Middleware {
    return Chain(
        HLSHeaders(),
        HLSCacheControl(),
        HLSRangeLogging(),
    )
}
```

---

### 5. Solução para Arquivos Estáticos

#### Opção A: Helper Function (Recomendado)

```go
// static.go
package middleware

import (
    "log"
    "net/http"
    "os"
)

// ServeStatic cria um handler para servir arquivos estáticos com middlewares
func ServeStatic(dir, routePrefix string, middlewares ...Middleware) http.Handler {
    // Valida que diretório existe
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        log.Fatalf("Directory %s does not exist", dir)
    }

    // Cria file server
    fs := http.FileServer(http.Dir(dir))

    // Remove prefix da URL antes de servir
    handler := http.StripPrefix(routePrefix, fs)

    // Aplica todos os middlewares
    return Chain(middlewares...)(handler)
}

// Uso:
// mux.Handle("/hls/", ServeStatic(".upload", "/hls/",
//     HLSComplete(),
//     Logger(),
//     CORS(),
// ))
```

#### Opção B: Registrar Diretamente no Mux

```go
// RegisterStatic é um helper para registrar diretamente no mux
func RegisterStatic(mux *http.ServeMux, dir, route string, middlewares ...Middleware) {
    handler := ServeStatic(dir, route, middlewares...)
    mux.Handle(route, handler)
}

// Uso:
// RegisterStatic(mux, ".upload", "/hls/",
//     HLSComplete(),
//     Logger(),
// )
```

#### Opção C: Middleware Genérico de Headers

```go
// WithHeaders cria middleware que adiciona headers customizados
func WithHeaders(headers map[string]string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for k, v := range headers {
                w.Header().Set(k, v)
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Uso para arquivos estáticos:
// mux.Handle("/hls/", WithHeaders(map[string]string{
//     "Accept-Ranges": "bytes",
//     "Cache-Control": "no-cache",
// })(http.StripPrefix("/hls/", http.FileServer(http.Dir(".upload")))))
```

---

## 🔄 Migração do Código Existente

### Antes (Atual)

```go
// http_in.go
var Routes = map[string]http.Handler{
    "/health": middleware.Use(
        middleware.WithInboundLogging(),
        middleware.WithOutgoingLogging(),
        healthHandler(),
    ),
}

// server.go
func SetupServer() (http.Handler, error) {
    mux := http.NewServeMux()
    hlsDir := os.Getenv("HLS_DIR")

    for endpoint, handler := range Routes {
        mux.Handle(endpoint, handler)
    }

    middleware.UseStaticFiles(mux, hlsDir, "/hls/")

    handler := middleware.Use(middleware.Cors(), nil, mux)
    return handler, nil
}
```

### Depois (Novo)

```go
// http_in.go
var Routes = map[string]http.Handler{
    "/health": middleware.Chain(
        middleware.Logger(),
    )(healthHandler()),
}

// server.go
func SetupServer() (http.Handler, error) {
    mux := http.NewServeMux()
    hlsDir := os.Getenv("HLS_DIR")

    // Registra rotas normais
    for endpoint, handler := range Routes {
        mux.Handle(endpoint, handler)
    }

    // Registra arquivos estáticos HLS com middlewares específicos
    mux.Handle("/hls/", middleware.ServeStatic(hlsDir, "/hls/",
        middleware.HLSComplete(),  // Headers + Cache + Range logging
        middleware.Logger(),       // Logging
    ))

    // Aplica middlewares globais
    handler := middleware.Chain(
        middleware.CORS(),
    )(mux)

    return handler, nil
}
```

---

## 📂 Estrutura de Arquivos Proposta

```
middleware/
├── types.go           # Middleware type, Chain(), Then(), ThenFunc()
├── response.go        # ResponseWriter wrapper para capturar status
├── cors.go           # Middleware CORS (refatorado)
├── logging.go        # Middlewares de logging (refatorado)
├── hls.go            # Middlewares específicos para HLS
├── static.go         # Helpers para static files
├── recovery.go       # Middleware de panic recovery (novo)
├── ratelimit.go      # Middleware de rate limiting (novo)
└── examples_test.go  # Exemplos de uso e testes
```

---

## 👥 Perspectivas de Diferentes Profissionais

### 1. Arquiteto de Software Go Sênior

**Opinião:**
> "Seu `Use(before, after, next)` funciona, mas não é o padrão da comunidade Go. O problema principal é que você está lutando contra a linguagem em vez de abraçá-la."

**Recomendações:**
- Go favorece **composição sobre configuração**
- Middlewares devem ser **closures que retornam closures**
- O padrão `func(http.Handler) http.Handler` é universal
- Separe before/after **dentro** do middleware, não externamente

**Para arquivos estáticos:**
```go
staticHandler := http.FileServer(http.Dir(dir))
mux.Handle("/hls/", Chain(
    HLSHeaders(),
    Logger(),
)(http.StripPrefix("/hls/", staticHandler)))
```

---

### 2. Desenvolvedor de Framework Web (Chi/Echo)

**Opinião:**
> "Você está reinventando a roda. Olhe como frameworks maduros fazem. Eles separaram claramente: (1) definição de middleware, (2) composição, (3) aplicação."

**Princípios de Design:**

1. **Middleware deve ser função pura** (sem estado, sem side effects na definição)
2. **Composição deve ser explícita e flexível**
3. **Developer Experience (DX) importa**
   - Middlewares devem ser plug-and-play
   - Documentação clara de ordem de execução
   - Evite "magia"

**Exemplo de composição:**
```go
// Chi-style (se tivéssemos um router)
r.Use(CORS, Logger, RateLimit)

// Ou manual (com stdlib)
handler = CORS(Logger(RateLimit(finalHandler)))

// Ou com Chain helper
handler = Chain(CORS, Logger, RateLimit)(finalHandler)
```

---

### 3. Engenheiro de Performance/SRE

**Opinião:**
> "Seu design atual tem overhead desnecessário. Cada request passa por `Use()` que faz checks de `if before != nil` e `if r.Method != http.MethodOptions`. Isso escala mal."

**Análise de Performance:**

```go
// ❌ Atual (runtime checks a cada request)
func Use(before, after, next) {
    return func(w, r) {
        if before != nil { ... }      // ← Check em runtime
        if r.Method != OPTIONS { ... } // ← Check em runtime
        if after != nil { ... }        // ← Check em runtime
    }
}

// ✅ Idiomático (composição em init time, zero overhead)
func Logger() Middleware {
    logger := log.New(os.Stdout, "[HTTP] ", log.LstdFlags) // ← UMA VEZ

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w, r) {
            logger.Printf("→ %s", r.URL.Path)
            next.ServeHTTP(w, r)
            logger.Printf("← %s", r.URL.Path)
        })
    }
}
```

**Benchmarks esperados:**
- Alocações por request: 0 extras
- Latência: Chain pre-computa cadeia (custo único)
- Memória: Closures capturam apenas o necessário

---

### 4. Autor de Biblioteca Open Source

**Opinião:**
> "Se você tornar isso uma lib, precisa pensar em **extensibilidade** e **interoperabilidade**. Seu design atual não se integra bem com outras bibliotecas."

**Checklist de Design de API:**

✅ **Interface Compatibility**
```go
// DEVE ser compatível com http.Handler
type Middleware func(http.Handler) http.Handler
```

✅ **Composabilidade com terceiros**
```go
import "github.com/chi/middleware"

Chain(
    CORS(),
    middleware.Logger,    // ← Chi middleware
    middleware.Recoverer, // ← Chi middleware
    CustomMiddleware(),   // ← Seu middleware
)(handler)
```

✅ **Documentação por exemplos**
```go
func ExampleCORS() {
    handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    }))

    http.ListenAndServe(":8080", handler)
    // Output: Server running with CORS enabled
}
```

---

### 5. Tech Lead de Plataforma (Streaming Expert)

**Opinião:**
> "Para HLS especificamente, há concerns únicos: byte-range requests, cache control, CORS preflight. Seu middleware precisa ser **aware** disso."

**Recomendações HLS-Específicas:**

#### Cache Strategy por Tipo
```go
func HLSCacheControl() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w, r) {
            if strings.HasSuffix(r.URL.Path, ".m3u8") {
                // Playlists mudam constantemente
                w.Header().Set("Cache-Control", "no-cache")
            } else if strings.HasSuffix(r.URL.Path, ".ts") {
                // Segmentos são imutáveis
                w.Header().Set("Cache-Control", "public, max-age=31536000")
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

#### CORS Específico para HLS
```go
func CORS() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w, r) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
            // IMPORTANTE: Exponha headers de range para ABR
            w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")

            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return // ← Não chama next para OPTIONS
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

#### Pipeline Recomendada para HLS
```go
mux.Handle("/hls/", Chain(
    CORS(),              // 1. Trata CORS e OPTIONS
    Logger(),            // 2. Loga requests
    HLSByteRange(),      // 3. Headers de range
    HLSCacheControl(),   // 4. Cache strategy inteligente
)(http.StripPrefix("/hls/", http.FileServer(http.Dir(dir)))))
```

---

### 6. DevRel/Educador Técnico

**Opinião:**
> "Seu código precisa ser **ensinável**. Quando outro dev olhar, deve entender imediatamente o que está acontecendo."

**Princípios de Código Limpo:**

1. **Nomeação Clara**
```go
// ❌ Confuso
Use(before, after, next)

// ✅ Auto-explicativo
Chain(Logger(), CORS(), Auth())(handler)
```

2. **Fluxo Visível**
```go
handler := Logger()(      // ← 1º a executar
    CORS()(               // ← 2º a executar
        RateLimit()(      // ← 3º a executar
            finalHandler))) // ← último
```

3. **Documentação Inline**
```go
// Chain compõe múltiplos middlewares da esquerda para direita.
// Exemplo: Chain(A, B, C) executa A → B → C → handler
func Chain(middlewares ...Middleware) Middleware
```

---

## 🎯 Recomendação Final

### Consenso de Todos os Perfis

1. **Adote `func(http.Handler) http.Handler`**
   - Idiomático em Go
   - Compõe com stdlib e terceiros
   - Zero overhead
   - Auto-contido

2. **Use composição explícita**
   ```go
   Chain(mw1, mw2, mw3)(handler)
   ```

3. **Trate arquivos estáticos como handlers normais**
   ```go
   mux.Handle("/hls/", Chain(
       HLSComplete(),
       Logger(),
   )(http.StripPrefix("/hls/", http.FileServer(http.Dir(dir)))))
   ```

4. **Crie middlewares específicos para HLS**
   - Cache diferenciado (.m3u8 vs .ts)
   - Logging de range requests
   - CORS otimizado

5. **Implemente ResponseWriter wrapper** (quando precisar de status code)

---

## 📋 Checklist de Implementação

### Fase 1: Infraestrutura Base
- [ ] Criar `types.go` com `Middleware` type e `Chain()`
- [ ] Criar `response.go` com `ResponseWriter` wrapper
- [ ] Adicionar testes unitários para `Chain()`

### Fase 2: Refatorar Middlewares Existentes
- [ ] Refatorar `cors.go` para novo padrão
- [ ] Refatorar `logging.go` para capturar status code
- [ ] Criar `hls.go` com middlewares específicos HLS
- [ ] Atualizar `static.go` com helpers

### Fase 3: Migrar Código de Aplicação
- [ ] Atualizar `http_in.go` para usar `Chain()`
- [ ] Atualizar `server.go` para novo padrão
- [ ] Remover função `Use()` antiga
- [ ] Atualizar registros de rotas estáticas

### Fase 4: Testes e Validação
- [ ] Adicionar testes de integração
- [ ] Validar streaming HLS continua funcionando
- [ ] Validar CORS em todas as rotas
- [ ] Benchmarks de performance

### Fase 5: Documentação
- [ ] Adicionar exemplos em `examples_test.go`
- [ ] Documentar cada middleware com godoc
- [ ] Atualizar README com novos padrões

---

## 🔗 Referências

**Frameworks Go que usam este padrão:**
- [Chi Router](https://github.com/go-chi/chi) - `func(http.Handler) http.Handler`
- [Gorilla Mux](https://github.com/gorilla/mux) - `MiddlewareFunc`
- [Negroni](https://github.com/urfave/negroni) - Middleware interface
- [Echo](https://echo.labstack.com/) - Similar pattern

**Artigos e Recursos:**
- [Writing HTTP Middleware in Go](https://www.alexedwards.net/blog/making-and-using-middleware)
- [Go HTTP Middleware Best Practices](https://golang.org/doc/articles/wiki/)
- [HLS Specification - RFC 8216](https://tools.ietf.org/html/rfc8216)

---

## 💡 Exemplos Completos

### Exemplo 1: Endpoint Simples com Logging

```go
// Define handler
func healthHandler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
            "timestamp": time.Now().Format(time.RFC3339),
        })
    })
}

// Registra com middlewares
mux.Handle("/health", Chain(
    Logger(),
)(healthHandler()))
```

### Exemplo 2: Arquivos Estáticos HLS com Pipeline Completa

```go
hlsDir := os.Getenv("HLS_DIR")

mux.Handle("/hls/", ServeStatic(hlsDir, "/hls/",
    CORS(),              // CORS headers
    Logger(),            // Request/response logging
    HLSHeaders(),        // Accept-Ranges: bytes
    HLSCacheControl(),   // Smart caching (.m3u8 vs .ts)
))
```

### Exemplo 3: Server Completo

```go
func SetupServer() (http.Handler, error) {
    mux := http.NewServeMux()

    // Rotas API
    mux.Handle("/health", Chain(Logger())(healthHandler()))
    mux.Handle("/api/videos", Chain(Logger(), Auth())(videosHandler()))

    // Arquivos estáticos HLS
    hlsDir := os.Getenv("HLS_DIR")
    mux.Handle("/hls/", ServeStatic(hlsDir, "/hls/",
        HLSComplete(),
        Logger(),
    ))

    // Middlewares globais
    handler := Chain(
        CORS(),
        Recovery(), // Panic recovery
    )(mux)

    return handler, nil
}
```

### Exemplo 4: Middleware Customizado

```go
// Crie seu próprio middleware facilmente
func RateLimit(requestsPerSecond int) Middleware {
    limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Uso
mux.Handle("/api/", Chain(
    RateLimit(100), // 100 req/s
    Logger(),
)(apiHandler()))
```

---

## ✨ Benefícios da Nova Arquitetura

### 1. Idiomático
- Segue padrões da comunidade Go
- Compatível com stdlib e frameworks populares

### 2. Composível
- Fácil criar pipelines de middlewares
- Reutilizável entre diferentes endpoints

### 3. Performático
- Zero overhead (composição em init time)
- Sem alocações extras por request

### 4. Extensível
- Fácil adicionar novos middlewares
- Compõe com libs de terceiros

### 5. Testável
- Cada middleware é uma unidade isolada
- Fácil escrever testes unitários

### 6. Específico para HLS
- Cache strategy inteligente
- Headers corretos para ABR
- CORS otimizado para streaming

---

**Próximos Passos:** Implementar a Fase 1 (Infraestrutura Base) e validar com testes.
