# Defer Nothing

*Building Backend Systems in Go*

---

# Chapter 2: From Request to Response

Ask an experienced backend engineer why a request is slow, and they rarely start by looking at code. They start by running the request in their head: it arrives here, passes through this, waits on that, comes back this way. Within a minute they have three suspects instead of three thousand lines.

That skill is learnable, and it is the foundation for almost everything later in the book: debugging, performance work, observability, and incident response. This chapter teaches it by following a single request, slowly, from the moment a client sends it to the moment the client receives an answer.

## Two Maps of the Same Request

Chapter 1 gave you a conceptual map: input, interpretation, validation, business decision, state transition, side effects, response. It describes what the system decides.

This chapter gives you the physical map: which components the request actually touches, in what order, and what each one is waiting for. The conceptual map tells you what must happen. The physical map tells you where it can go wrong.

```
Client
  ↓
TCP (and TLS)
  ↓
HTTP
  ↓
Router
  ↓
Middleware
  ↓
Handler
  ↓
Service
  ↓
Repository
  ↓
Database
  ↓
Repository
  ↓
Service
  ↓
Handler
  ↓
HTTP Response
```

One caution before we walk it. The layers named Handler, Service, and Repository are roles, and the book will justify them properly in Chapter 17. When we write the first Forge service in Chapter 4, it will deliberately have less structure than this diagram shows. The diagram is the destination, and we want a name for each stop before we decide how many stops the system needs.

## Walking the Path

We will follow the same request as before:

```
POST /tasks
{ "title": "Write chapter one" }
```

**The connection.** Before any HTTP exists, the client must find the server (DNS), open a TCP connection (a handshake), and usually negotiate encryption (TLS). This happens before your code sees a single byte, and it can fail in ways your code will never log: the name does not resolve, the connection is refused, the handshake times out. From the client's side, these look like "the service is down." From your side, nothing happened. Remember this. Some failures are invisible to the server, which is why clients need their own timeouts and why monitoring must include a view from outside.

**Bytes become HTTP.** The server reads a stream of bytes and parses it into a method, a path, headers, and a body. This is where a server defends itself. How large may a header be? How large may the body be? How long will we wait for a slow client to finish sending? A server with no answers to these questions can be held hostage by a client that sends one byte per second. Chapter 5 covers the limits to configure.

**The router.** The router matches the method and path to the code that should handle them. No match produces a 404. A matching path with the wrong method produces a 405. This step is cheap, and its failures are the easiest to diagnose, which makes it a good first suspect when a new endpoint "doesn't work."

**Middleware.** Before the handler runs, the request passes through a chain of wrappers that deal with concerns shared by every endpoint: recovering from panics, assigning a request ID, logging, authenticating. Middleware sees every request, so a mistake here affects all of them. It is also order-sensitive. Logging before authentication records requests that were rejected, and logging after does not. Chapter 7 treats this carefully.

**The handler.** The handler is the translator between HTTP and the application. It decodes the JSON body, calls the application logic with plain values, and turns the result back into a status code and a body. Its defining property is what it *knows*: it understands HTTP, and it should not understand SQL.

**The service.** The service is where decisions live: is this allowed, is this valid given the current state, what should change? It understands the business rules and knows nothing about HTTP or SQL. That ignorance is deliberate, and it is what lets the same rules serve an HTTP endpoint, a background worker, and a test.

**The repository.** The repository translates between the application's objects and the storage format. It is the only place that knows what the tables look like.

**The database.** Here the request finally changes the world. The database is usually the slowest hop and always the most stateful one: it has locks, connection limits, and its own failure modes.

**The return path.** The result travels back through the same layers, and at each boundary it is translated. A database row becomes a domain object, which becomes a response structure, which becomes JSON bytes, which become an HTTP response.

The same table summarizes who owns what.

| Hop | Owns | Must not know | Typical failure | What the client experiences |
| --- | --- | --- | --- | --- |
| Connection | Reaching the server | Anything about the request | DNS failure, refused connection, TLS error | Network error, no HTTP response |
| HTTP parsing | Reading the request safely | Business meaning | Oversized body, malformed request, slow client | 400, 413, 408, or a dropped connection |
| Router | Choosing the handler | Request contents | No matching route | 404 or 405 |
| Middleware | Cross-cutting concerns | Specific endpoint logic | Panic, rejected credentials | 500 or 401 |
| Handler | HTTP in, HTTP out | SQL, storage layout | Undecodable body | 400 |
| Service | Decisions and rules | HTTP, SQL | Rule violation | 409, 403, or 422 |
| Repository | Storage translation | HTTP, business policy | Query error, constraint violation | Usually 500, sometimes 409 |
| Database | Durable state | Everything above it | Timeout, lock wait, connection exhaustion | Slow response, then 500 or 503 |

## Errors Are Translated at Every Boundary

On the return path, the most important thing that happens is translation of errors, and it is where many backends leak or lose information.

Suppose two clients create a record that must be unique, and the second one collides. The database reports a unique-constraint violation, in the database's vocabulary. If that message travels unchanged to the client, the client learns the name of an internal index, and the error means nothing to anyone who is not reading your schema.

The right shape is a chain of translations. The repository recognizes the constraint violation and reports it as "this already exists," a concept the application understands. The service decides what that means for the business. The handler maps it to a response the client can act on: a 409 status with a clear error code.

```
Postgres: unique violation on idx_tasks_...
        ↓  (repository translates)
Application: already exists
        ↓  (service decides)
Application: conflict
        ↓  (handler maps)
HTTP: 409 with { "error": { "code": "ALREADY_EXISTS", ... } }
```

Each layer speaks its own language and translates at the boundary. Chapter 8 builds the full error model. What matters now is the principle: **an error should be described in the vocabulary of the layer that receives it, while the original cause is preserved for the logs.**

## Where the Time Goes

Here is a rough picture of where a typical request spends its time. The figures are illustrative, not measurements, but the shape is realistic.

| Phase | Illustrative time | Notes |
| --- | --- | --- |
| Connection and TLS (new connection) | 10 to 30 ms | Near zero when a connection is reused |
| Parsing, routing, middleware | Under 1 ms | Cheap and rarely the problem |
| Handler and service logic | Under 1 ms | Pure computation is fast |
| Database round trip | 2 to 20 ms | Depends on the query, locks, and load |
| Response writing | Under 1 ms | Unless the body is very large |

The pattern is consistent: the code you write is rarely the slow part. Time is spent *waiting*, on networks and on databases. This observation shapes the second half of the book. If most of a request's life is waiting, then the important engineering questions are how long we are willing to wait, what we do when the wait gets too long, and how many requests we can have waiting at once. Those questions lead directly to timeouts, cancellation, and concurrency.

## The Four-Question Trace

To trace a request, ask four questions at each hop.

1. **What arrives here?** The exact input, including its format and its assumptions.
2. **What leaves here?** The output, and who consumes it next.
3. **What can go wrong here?** The failure modes specific to this hop.
4. **What is this hop waiting on?** The dependency or resource it blocks on, and for how long.

The fourth question is the one beginners skip, and it explains more production behavior than the other three combined.

## A Worked Trace

A teammate reports: "`POST /tasks` returns a 500 about one time in fifty." Rather than reading code, trace.

The connection is unlikely to be the cause, because the client received a 500, which means an HTTP response was produced. The router matched, because a handler ran or a middleware recovered a panic. So the failure is in middleware, handler, service, repository, or database.

Failing only occasionally narrows it further. A bug in decoding or validation would usually depend on the input, so ask whether the failing requests share something. A failure that depends on load or timing points to the database or to concurrency: connection pool exhaustion, a lock wait that times out, or a collision on a unique value.

Now you know what to look for: do the failing requests share an input pattern, and did the database report a timeout or a constraint error at those moments? You have turned an open-ended hunt into two specific checks, and you did it without opening an editor. This is the skill. It works because the physical map gave you a list of places where intermittent failures actually live.

## Engineering Lens

**What problem did we solve?** We gave ourselves a physical map of a request and a method for reasoning about it, so that debugging and design start from structure instead of guesswork.

**What complexity did we introduce?** A set of named layers, which may be more structure than a small service needs. We accept this now because naming the layers makes their trade-offs discussable later.

**What failure modes appeared?** Failures invisible to the server, such as DNS and connection problems. Failures hidden by badly translated errors. Slow requests dominated by waiting rather than computation.

**What would break at 10× scale?** The waiting. Connection pools, database capacity, and slow dependencies become the limiting factors long before CPU does.

**When should we not use this model?** For a tiny internal tool with one endpoint and one user, a full layered view is overkill. The tracing habit still applies even when the layers are collapsed into one function.

## Exercises

1. Choose one endpoint from any API you have used. Draw its path through the hops above. Which hop is most likely to be slow? Which is most likely to fail in a way the server never logs?
2. Take the error chain for a unique-constraint violation and write the equivalent chain for "the database is unreachable." What should the client see, and what should the logs record?
3. A request logs "completed in 45 ms" on the server, but the client measured 400 ms. List three places the missing time could have gone.
4. For each hop, write one sentence describing what it is waiting on. Which of those waits have no timeout today?

Chapter 3 moves from understanding a request to designing a system for it. Before any Go is written, we decide what Forge must do, what it must never do, and what happens when it fails.

---

# Chapter 3: Designing Before Coding

The expensive bugs in a backend are rarely typos. They are correct implementations of unclear decisions. The code does exactly what it was written to do, and what it was written to do was never properly thought through.

Consider a pattern that appears in many young systems. A task has a status, and someone adds it as a free-form string because that is the quickest option. Over the following months, different parts of the system write `done`, `completed`, `finished`, and `complete`. No single change was wrong. The system just never decided what the legal states were, so every developer decided independently. Cleaning it up later means a data migration, a compatibility period, and an apology to everyone who built on the old values.

This chapter is about preventing that, by deciding the important things before writing code. It is not about producing long documents.

## The Right Size of Design

Design before coding does not mean designing everything. It means making the decisions that are expensive to change, and recording them in a form small enough that people actually read it.

There is an apparent tension with this book's title, since "defer nothing" could be read as "decide everything now." It means something narrower: decide *how things fail and what must always be true* now, and leave cheap decisions, such as file layout and internal naming, for when you have more information.

For Forge's first slice, the design is six short artifacts. Each answers one question.

| Artifact | The question it answers |
| --- | --- |
| Requirements | What must the system do, and what is out of scope? |
| Entities | What things does the system keep track of? |
| Invariants | What must always be true? |
| State transitions | How is a thing allowed to change? |
| API contracts | What exactly do clients send and receive? |
| Failure modes | What happens when something goes wrong? |

## Requirements: Including What We Will Not Build

Write requirements as short statements, and write the exclusions with equal care. An explicit "not now" prevents arguments later.

| Kind | Requirement |
| --- | --- |
| Functional | A client can create a task with a title |
| Functional | A client can retrieve one task by ID |
| Functional | A client can list tasks |
| Functional | A task moves through a defined lifecycle |
| Non-functional | Each endpoint responds within a reasonable time under normal load |
| Non-functional | No task is silently lost or duplicated by a successful response |
| Non-functional | Every failure produces a clear, consistent error response |
| Out of scope for now | Users and authentication (Chapter 9), pagination (Chapter 10), background execution (Chapter 25) |

The non-functional rows matter most and are the ones most often omitted. "No task is silently lost" is a requirement that will later force decisions about transactions and storage. Writing it now means we notice when a design violates it.

## Entities: What the System Remembers

Forge begins with one entity, the **Task**.

| Field | Meaning |
| --- | --- |
| `id` | Unique identifier, opaque to clients |
| `title` | Short human-readable description |
| `status` | Current lifecycle state |
| `created_at` | When the task was created |
| `updated_at` | When the task last changed |

One decision hides in this table: the ID is *opaque*. Clients treat it as a label, never as a number to increment or guess. That freedom lets us change how IDs are generated later without breaking anyone, and it avoids exposing how many tasks exist. A decision that costs nothing now would be painful to reverse once clients depend on the opposite.

Later chapters add users and task events. We design them when we need them.

## Invariants: What Must Always Be True

An invariant is a statement that holds at every moment, regardless of which request, worker, or script touched the data. Writing them down is the most valuable design step, because they become the checklist against which every later piece of code is judged.

| Invariant | Where it will eventually be enforced |
| --- | --- |
| A title is non-empty and at most 200 characters | Validation, and a database constraint as a backstop |
| Status is always one of the defined states | Domain model and database constraint |
| A completed or cancelled task never changes again | Domain model |
| Status changes only along allowed transitions | Domain model, and conditional writes in storage |
| `updated_at` is never earlier than `created_at` | Storage layer |

The right-hand column shows a pattern worth noticing. Important invariants are enforced in more than one place, because any single layer can be bypassed by a bug, a script, or a future developer. Defense in depth is cheap for the invariants that matter and wasteful for the ones that do not, so choose deliberately.

## State Transitions: The Lifecycle

Chapter 1 showed a first sketch. Now we make the decisions that sketch left open. Forge tasks have five states:

```
            ┌──────────► cancelled
            │
pending ────┼──► running ──┬──► completed
            │              │
            │              └──► failed
            │                     │
            └◄────────────────────┘
                   (retry)
```

Each arrow is a decision, and the table records it.

| From | To | Allowed | Notes |
| --- | --- | --- | --- |
| pending | running | Yes | Work begins |
| pending | cancelled | Yes | Caller abandons it before it starts |
| running | completed | Yes | Terminal |
| running | failed | Yes | Work did not succeed |
| running | cancelled | Yes | Abandoned mid-flight |
| failed | pending | Yes | Retry: re-queues the task |
| pending | completed | No | A task cannot finish without running |
| completed | anything | No | Terminal states are final |
| cancelled | anything | No | Terminal states are final |

Two choices deserve a defence. First, `failed` is *not* terminal, because retrying failed work is a normal operation, and modelling it as a transition back to `pending` keeps the lifecycle a single clean loop. Second, `completed` *is* terminal. If someone later needs to reopen finished work, that is a new feature with its own rules, not a loophole.

Three questions, asked about any lifecycle, will find most design flaws. They are the questions that turn a diagram into a design.

> What states are legal? What transitions are legal? What happens if two requests attempt the same transition at once?

The first two are answered above. The third is the one that separates a sketch from a system.

## When Two Requests Collide

Suppose two clients both ask to start the same `pending` task at the same moment. Each reads `pending`, each is permitted to proceed, and each issues the change. If nothing prevents it, both succeed, and the task starts twice.

There are only a few ways to respond, and choosing among them is a design decision, not an implementation detail.

| Policy | Behavior | Verdict for starting a task |
| --- | --- | --- |
| Last write wins | The second change silently overwrites the first | Unacceptable: the first caller is misled |
| Reject the loser | The second request fails with a conflict | Correct: exactly one caller wins, the other is told |
| Queue and retry | The second waits, then re-checks | Sometimes right, but adds complexity |

For Forge, we choose to **reject the loser**. A transition is only valid if the task is still in the state the caller expected. The loser receives a conflict response and can re-read the task to see what happened. Chapter 13 shows the mechanism, a write that is conditional on the current state. The point here is that the *behavior* was decided before the code, so the code has a target.

## API Contracts: The Promise to Clients

The API is a contract. Once clients depend on it, changing it is expensive, so it deserves precision before implementation.

| Method and path | Purpose | Success | Common errors |
| --- | --- | --- | --- |
| `POST /tasks` | Create a task | 201 with the task | 400 invalid input |
| `GET /tasks/{id}` | Fetch one task | 200 with the task | 404 not found |
| `GET /tasks` | List tasks | 200 with a list | none expected |

A created task looks like this:

```json
{
  "id": "tsk_8f3a1c",
  "title": "Write chapter one",
  "status": "pending",
  "created_at": "2026-10-03T09:14:00Z",
  "updated_at": "2026-10-03T09:14:00Z"
}
```

Every error uses one consistent shape, so clients write one error handler:

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "title is required"
  }
}
```

Notice what the contract fixes: field names, the `pending` initial status, timestamp format, and the error structure. It deliberately leaves out how tasks are stored and how IDs are generated. A good contract promises behavior and hides mechanism.

## Failure Modes: Deciding Before It Happens

For every endpoint, ask what happens when each thing goes wrong, and write down the intended behavior. Doing this before coding means failure handling is designed instead of accidental.

| Situation | Intended behavior |
| --- | --- |
| Body is not valid JSON | 400 with `INVALID_ARGUMENT` |
| Title is missing or too long | 400 with a message naming the field |
| Task ID does not exist | 404 with `NOT_FOUND` |
| Transition is not allowed | 409 with a code naming the conflict |
| Database is unavailable | 503, with nothing partially written |
| Client disconnects mid-request | Stop work; do not leave half-finished state |
| Client retries a create after a timeout | Open question, see below |

## Open Questions Belong in the Design

A design should record what it has *not* solved. The last row above is a real problem: if a client times out and retries `POST /tasks`, the first request may have succeeded, and now there are two identical tasks. Forge's first version does not prevent this. Writing the question down makes the gap visible, so we solve it deliberately in Chapter 10 with idempotency, rather than discovering it in production.

## The One-Page Design

All of this fits on a single page. For any new feature, a page with these headings is enough:

```
Goal            One sentence on what this enables
Requirements    Functional, non-functional, out of scope
Entities        What is stored, and the decisions hidden in the fields
Invariants      What must always be true, and where it is enforced
Transitions     States, allowed moves, and the collision policy
Contract        Endpoints, shapes, status codes, error format
Failure modes   Each failure and the intended behavior
Open questions  What we know we have not solved
```

If a design takes more than a page for a feature this size, it is probably trying to decide things that are cheap to decide later.

## Engineering Lens

**What problem did we solve?** We made the decisions that are expensive to reverse, the lifecycle, the invariants, the contract, and the collision policy, before they could be made accidentally in code.

**What complexity did we introduce?** A small amount of up-front writing and the discipline to keep it current. A design that drifts from the code is worse than none.

**What failure modes appeared?** Unclear states, silent overwrites under concurrency, duplicate creates from retries, and half-written state when dependencies fail. Each is now a named item with intended behavior or an open question.

**What would break at 10× scale?** Collisions that were theoretical become constant, so the reject-the-loser policy will be exercised thousands of times a day. A lifecycle with unclear rules would be producing corrupt data at the same rate.

**When should we not design this much?** For a prototype you will throw away, or a change small enough to understand in one sitting, a full page is excessive. Scale the design to the cost of being wrong.

## Exercises

1. Add a `blocked` state to the Forge lifecycle, for tasks waiting on something external. Update the transition table. Which existing decisions does it force you to revisit?
2. Write the one-page design for a new feature: assigning a task to a user. What are the invariants? What collision policy applies when two people assign the same task at once?
3. Pick a real API you use and find one place where its contract leaks implementation detail. What would the cost be of fixing it now?
4. Choose one open question from this chapter and propose two possible answers. What would each cost, and what would each protect against?

With the design settled, Chapter 4 finally opens an editor. We write the first Forge service in Go, teaching only the language features the system needs, and discover quickly which of today's decisions the code makes easy and which it makes hard.