# Defer Nothing

*Building Backend Systems in Go*

---

# Chapter 4: The First Go Service

Three chapters in, you have a vocabulary, a physical map of a request, and a design for Forge. This chapter writes the code.

We will not study Go as a language. We will build the smallest version of Forge that honors the design from Chapter 3, and each piece of Go appears at the moment the system needs it. By the end you will have a running service, a test suite, and an honest list of what the service still gets wrong.

You will need Go 1.22 or newer, because the router we use relies on method-aware patterns introduced in that release.

## The Smallest Possible Server

Here is a complete web server:

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	http.ListenAndServe(":8080", nil)
}
```

Run it with `go run .` and visit `localhost:8080/health`. It works, which is exactly the problem.

Look at the last line. `ListenAndServe` returns an error, and we ignored it. If port 8080 is already taken, the program exits silently and nothing tells us why. This is the first small act of deferring: we assumed success and postponed the question of failure. The Go compiler allowed it without complaint, which is a reminder that Go gives you tools and trusts you to use them.

The same line has a second problem. The default server has no timeouts, so a client that connects and sends nothing holds a connection open indefinitely. Chapter 2 warned about slow clients, and this is where that warning applies. Here is the version we will actually use:

```go
httpServer := &http.Server{
	Addr:              ":8080",
	Handler:           srv.routes(),
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      10 * time.Second,
	IdleTimeout:       60 * time.Second,
}
```

Two lines of difference, and the server now bounds how long it will wait on any client. This is the book's habit in miniature: decide how it fails at the moment you decide how it works.

## Setting Up the Project

Go organizes code into **packages** and packages into **modules**. A package is a directory of files that share a name and are compiled together. A module is a collection of packages with a version, described by a `go.mod` file at its root.

```
mkdir forge && cd forge
go mod init forge
```

The `main` package is special: it produces an executable, and its `main` function is where execution starts. For now everything in Forge lives in one package. This is a deliberate under-structuring. Chapter 17 splits the code into packages once there is a reason to, and the pain of this single package will be part of that reason.

The service will have four files.

| File | Responsibility |
| --- | --- |
| `task.go` | The Task type, statuses, and title validation |
| `store.go` | Where tasks are kept in memory |
| `handlers.go` | HTTP handlers and the router |
| `main.go` | Starting the server |

## The Task

Chapter 3 defined a task with five fields and five statuses. Go expresses this with a **struct**, a named collection of fields, and a set of constants.

```go
package main

import (
	"errors"
	"fmt"
	"time"
	"unicode/utf8"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

const maxTitleLength = 200

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func validateTitle(title string) error {
	if title == "" {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(title) > maxTitleLength {
		return fmt.Errorf("title must be at most %d characters", maxTitleLength)
	}
	return nil
}
```

Several Go ideas appear here, each doing real work.

| Idea | Where it appears | What it does |
| --- | --- | --- |
| Named type | `type Status string` | Gives status values their own type, so the compiler distinguishes them from arbitrary strings |
| Struct tags | `` `json:"created_at"` `` | Tells the JSON encoder which field name to use, matching our API contract |
| Functions returning `error` | `validateTitle` | Reports failure as an ordinary return value instead of an exception |
| `nil` | `return nil` | Means "no error." An error is just a value, and `nil` is the absence of one |

A note on the title check. We count *runes* (characters) with `utf8.RuneCountInString` rather than bytes with `len`. A title in Bengali or Japanese uses several bytes per character, and counting bytes would reject titles that are well under 200 characters. The Chapter 3 invariant said 200 characters, so the code must count characters.

Notice also what the code does *not* enforce. `Status` is a distinct type, but Go will still happily let you assign `Status("banana")`. The invariant "status is always one of the defined states" is documented, not guaranteed. Chapter 18 closes that gap.

## The Store

Forge needs somewhere to keep tasks. For now that is a map in memory, which is the simplest thing that satisfies "create, fetch, list."

```go
package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Store struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

func NewStore() *Store {
	return &Store{tasks: make(map[string]Task)}
}

func newID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "tsk_" + hex.EncodeToString(b), nil
}

func (s *Store) Create(title string) (Task, error) {
	id, err := newID()
	if err != nil {
		return Task{}, err
	}

	now := time.Now().UTC()
	task := Task{
		ID:        id,
		Title:     title,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[id] = task
	return task, nil
}

func (s *Store) Get(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return task, nil
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].CreatedAt.Equal(tasks[j].CreatedAt) {
			return tasks[i].ID < tasks[j].ID
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})
	return tasks
}
```

This file introduces more Go than any other in the chapter, so it is worth slowing down.

**Maps and slices.** `map[string]Task` is a lookup table from ID to task. `[]Task` is a slice, a growable list. `List` copies the map's values into a slice because maps have no defined order, and an API that returns tasks in random order is a bug waiting for a client to depend on it. We sort by creation time, and break ties by ID so the order is fully deterministic. Notice `make([]Task, 0, len(s.tasks))`: it creates an empty but non-nil slice. That matters, because a nil slice encodes to JSON as `null`, and our contract promises a list, so an empty store must produce `[]`.

**Pointers and methods.** `NewStore` returns `*Store`, a *pointer* to a Store, and the methods are declared on `*Store`. A pointer refers to the one real Store rather than a copy. If the methods received copies, each would operate on its own private map and nothing would ever be shared. Methods are functions attached to a type, and the `(s *Store)` before the name is the receiver, the thing the method acts on.

**Errors.** `ErrNotFound` is a *sentinel error*, a named, package-level error value that callers can recognize. Go has no exceptions. A function that can fail returns an `error` alongside its result, and the caller checks it immediately. This feels repetitive at first, and it is deliberate: every failure point is visible in the code. `error` is itself an interface, a type defined by behavior rather than structure, and we will use that fact more in Chapter 28.

**Returning copies.** `Get` returns a `Task` value, not a pointer into the map. The caller receives a copy, so nothing outside the store can modify stored data without going through the store. This is a small decision that protects an invariant.

**The mutex and `defer`.** Here the book's title finally appears in code. A `sync.RWMutex` is a lock: any number of readers may hold it at once, but a writer holds it alone. And look at how it is used:

```go
s.mu.Lock()
defer s.mu.Unlock()
```

`defer` schedules a call to run when the surrounding function returns, by any path. We acquire the lock and, on the very next line, state when it will be released. Whether the function returns normally, returns early, or panics, the lock is released. Compare the alternative, where an `Unlock` placed at the bottom of the function is skipped by an early `return` someone adds next year. This is the idea the book is named after: the cleanup is decided at the moment of acquisition.

You may wonder why a lock is here at all, since we have not written any concurrent code. We did not need to. Go's HTTP server handles every incoming request in its own goroutine, so our handlers already run concurrently, and a plain map shared between them is a data race. Part V explains goroutines and locks properly. For now, accept that the lock is required and verify it yourself in the exercises.

## The Handlers and the Router

Now the layer that speaks HTTP. This is the handler role from Chapter 2: decode the request, call the application, encode the response.

```go
package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	store *Store
}

type createTaskRequest struct {
	Title string `json:"title"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", s.createTask)
	mux.HandleFunc("GET /tasks", s.listTasks)
	mux.HandleFunc("GET /tasks/{id}", s.getTask)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "request body must be valid JSON")
		return
	}

	title := strings.TrimSpace(req.Title)
	if err := validateTitle(title); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
		return
	}

	task, err := s.store.Create(title)
	if err != nil {
		log.Printf("create task: %v", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	task, err := s.store.Get(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "task not found")
			return
		}
		log.Printf("get task: %v", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	writeJSON(w, http.StatusOK, task)
}
```

Read `createTask` as a sequence of stages from Chapter 1. It limits the input size (**input**), decodes JSON (**interpretation**), trims and validates the title (**validation**), stores the task (**state transition**), and responds (**response**). The business decision stage is empty for now, since Forge has no rules about who may create tasks. The structure is already there, waiting for the rules.

A few details connect directly to earlier chapters.

**The error shape is the contract.** `writeError` produces exactly the JSON from Chapter 3, so every failure uses one format. Handlers never build error responses by hand.

**Errors are translated at the boundary.** `getTask` receives `ErrNotFound` from the store, which is storage vocabulary, and converts it to a 404 with a client-facing code. It uses `errors.Is` rather than `==`, so the check still works when an error has been wrapped with extra context, a technique Chapter 28 develops. Any *other* error becomes a generic 500 with the detail logged server-side and kept away from the client.

**Interfaces appear naturally.** `routes()` returns an `http.Handler`, which is an interface: any type with a `ServeHTTP` method qualifies. `http.ResponseWriter` is also an interface. We have used interfaces for two chapters without defining one ourselves, which is typical. Go interfaces describe what a value can do, and the standard library is full of them.

**Routing is declarative.** Patterns such as `"GET /tasks/{id}"` bind a method and a path to a handler, and `r.PathValue("id")` retrieves the wildcard. The router returns 404 or 405 for everything else, as Chapter 2 described.

## Starting the Server

```go
package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	srv := &Server{store: NewStore()}

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           srv.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("forge listening on :8080")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

This time we check the error. If the port is taken, the program says so and exits with a non-zero status.

Run it with `go run .` and try the contract:

```
$ curl -s -X POST localhost:8080/tasks -d '{"title":"Write chapter one"}'
{"id":"tsk_8f3a1c9d2b47","title":"Write chapter one","status":"pending","created_at":"2026-10-03T09:14:00.512Z","updated_at":"2026-10-03T09:14:00.512Z"}

$ curl -s localhost:8080/tasks/tsk_8f3a1c9d2b47
{"id":"tsk_8f3a1c9d2b47", ... }

$ curl -s -X POST localhost:8080/tasks -d '{}'
{"error":{"code":"INVALID_ARGUMENT","message":"title is required"}}

$ curl -s localhost:8080/tasks/tsk_nope
{"error":{"code":"NOT_FOUND","message":"task not found"}}
```

Each response matches the contract we wrote before there was any code. That is the payoff of Chapter 3: the code has a target, and you can check it against the target.

## Testing It

A service you cannot test is a service you cannot safely change. Go's standard library includes everything needed to test a handler without starting a real server.

```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func newTestHandler() http.Handler {
	return (&Server{store: NewStore()}).routes()
}

func TestCreateTask(t *testing.T) {
	h := newTestHandler()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid", `{"title":"Write chapter one"}`, http.StatusCreated},
		{"missing title", `{}`, http.StatusBadRequest},
		{"blank title", `{"title":"   "}`, http.StatusBadRequest},
		{"too long", `{"title":"` + strings.Repeat("a", 201) + `"}`, http.StatusBadRequest},
		{"not json", `not json`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	h := newTestHandler()

	createReq := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"Write chapter one"}`))
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, createReq)

	var created Task
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created task: %v", err)
	}
	if created.Status != StatusPending {
		t.Fatalf("status = %q, want %q", created.Status, StatusPending)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/tasks/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRec.Code, http.StatusOK)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/tasks/tsk_nope", nil)
	missingRec := httptest.NewRecorder()
	h.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want %d", missingRec.Code, http.StatusNotFound)
	}
}

func TestStoreConcurrentCreate(t *testing.T) {
	store := NewStore()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := store.Create("task"); err != nil {
				t.Errorf("create: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := len(store.List()); got != 50 {
		t.Fatalf("task count = %d, want 50", got)
	}
}
```

The first test is *table-driven*: a list of cases and one loop that runs them. Adding a case costs one line, which is why this is the standard Go style. Chapter 33 develops testing fully.

The last test uses `sync.WaitGroup` and `go func()` to create fifty tasks at once. We use them without full explanation, because the point is the result, and Chapter 23 explains them. Run the suite twice:

```
go test ./...
go test -race ./...
```

The `-race` flag turns on the race detector, which watches for unsynchronized access to shared memory. It is the single most valuable testing flag in Go, and it should be part of your habits from the first day.

## What the Service Still Gets Wrong

The service works, and it passes its tests. It is also nowhere near production. Compare it honestly to the design from Chapter 3.

| Chapter 3 commitment | Status in this chapter |
| --- | --- |
| No task silently lost | **Violated.** Tasks live in memory, so a restart erases everything |
| One consistent error format | **Partly met.** Our handlers comply, but the router's own 404 and 405 responses are plain text |
| Title rules enforced | Met |
| Status always a valid value | Documented only; nothing prevents an invalid value |
| Allowed transitions enforced | Not started: no endpoint changes a task's status |
| Retried creates do not duplicate | Open: a repeated `POST` creates a second task |
| Failures bounded and visible | **Weak.** No request logging, no request IDs, no graceful shutdown |

Other gaps are quieter. `GET /tasks` returns every task, so with a million tasks it returns a million tasks. An oversized body currently produces a 400 about invalid JSON when 413 would be the accurate answer. The server cannot be stopped cleanly: a deploy would cut off requests in flight.

None of this is a failure of the chapter. It is the point of the chapter. Every row in that table is a future chapter with a concrete reason to exist, and you can now see each one in running code instead of in a bullet list. This is how the book's rhythm works: a simple implementation, the production problem it eventually meets, and then the concept that solves it.

## What Part I Gave You

Four chapters, and no framework. You now have a way of thinking about a backend as a system that transforms state, a physical map for tracing a request through it, a design written before the code, and a working service whose weaknesses you can name. Part II begins with the first of those weaknesses that every client will notice: HTTP itself. We stop treating it as magic, learn what the protocol actually guarantees, and rebuild these handlers with that understanding.

## Engineering Lens

**What problem did we solve?** We turned the Chapter 3 design into a working service with a tested contract, using only the standard library.

**What complexity did we introduce?** A lock, because our handlers run concurrently, and a small amount of Go-specific vocabulary: packages, pointers, errors as values, `defer`. We introduced no frameworks and no abstractions beyond what the code needed.

**What failure modes appeared?** Data loss on restart, requests cut off on shutdown, unbounded list responses, duplicate creates on retry, and inconsistent errors from the router. All are named in the table above.

**What would break at 10× scale?** The in-memory store would exhaust memory, `GET /tasks` would become unusably slow, and running a second instance would give each instance its own private, disagreeing set of tasks.

**When should we not build this way?** An in-memory store is only appropriate for a prototype, a test, or a cache. The moment the data matters, it needs durable storage, which is where Part III begins.

## Exercises

1. Remove the mutex from `Store` (delete the `Lock`, `RLock`, and their `defer` lines) and run `go test -race ./...`. Read the race detector's report. Then run the service and hit it with many concurrent `POST` requests. What happens?
2. Add a `GET /health` endpoint that returns `{"status":"ok"}`. Where does it belong, and does it need the store?
3. Restart the service after creating a few tasks and list them. Which sentence from the Chapter 3 requirements does this violate? Write down what a fix would need to guarantee.
4. Send a `POST /tasks` body with an unknown field, such as `{"title":"x","colour":"red"}`. What happens? Look up `DisallowUnknownFields` on the JSON decoder and decide whether strict or lenient is the better contract. What does each choice cost clients that evolve over time?
5. Send a body larger than 1 MiB. What status do you get, and what status should the contract promise? Sketch how the handler could tell the two failures apart.