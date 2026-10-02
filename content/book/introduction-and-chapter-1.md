# Defer Nothing

*Building Backend Systems in Go*

---

# Introduction

## The Easy Part Is Easy

Go makes starting feel almost suspiciously simple. A working web server fits in a handful of lines, compiles in a second, and ships as a single binary. For a few hours, it feels like backend engineering has been solved.

Then the service meets reality. A client retries a request that already succeeded. Two users change the same record at the same moment. The database slows down, and every request waiting on it slows down too. A deploy kills a request halfway through a write. At 3 AM, someone asks what happened, and the only evidence is a line of text that says `something failed`.

None of these problems are Go problems. They are backend problems, and they arrive in every language. Go simply gets you to them faster, which is a gift if you are prepared and a shock if you are not. This book is about being prepared.

## Why "Defer Nothing"

In Go, the `defer` statement schedules a call to run when the surrounding function returns. It is one of the language's best ideas: you open a file, and on the very next line you state when it will be closed. The cleanup is decided at the moment the resource is acquired, not left to memory and good intentions.

Backend engineering has the opposite temptation. We defer the hard parts. We will handle errors later. We will add timeouts later. We will think about two requests at once later. We will make it observable later. Later usually arrives as an outage.

The title is a working philosophy: decide how it fails when you decide how it works. Handle the error now. Bound the work now. Make the behavior visible now. Every chapter ahead returns to this habit.

## What This Book Is

This is a practical book about designing, building, testing, deploying, and operating production backend services in Go. It is not a Go language reference, and Go syntax appears only when the system you are building needs it.

It is built around one idea: **start with a request, follow it all the way to a production service, and learn the engineering decisions in between.**

A backend is not an HTTP handler. It is a system that accepts requests, changes state, talks to other systems, fails under imperfect conditions, and must stay understandable while doing all of that. Handlers are the visible tip of it.

## Five Principles

Five principles run through every part of the book.

| Principle | What it means in practice |
| --- | --- |
| Make it work | Start simple. A working system teaches more than a perfect design on paper. |
| Make the boundaries explicit | Know where one responsibility ends and another begins: handler, service, database, network. |
| Make failure ordinary | Treat errors, timeouts, and partial failures as normal inputs, not exceptions to the plan. |
| Make concurrency deliberate | Every goroutine has an owner, a lifetime, and a reason to exist. |
| Make production behavior observable | If you cannot see what the system is doing, you do not control it. |

## The Running Project: Forge

Instead of unrelated toy examples, you will build one system throughout the book: **Forge**, a task and work management API.

Forge begins small. In Part I it has three endpoints, `POST /tasks`, `GET /tasks`, and `GET /tasks/:id`, and nothing else. By the final chapters it has authentication, PostgreSQL, caching, background workers, event streaming, metrics, tracing, a container image, and a deployment pipeline. Each addition arrives at the moment the system genuinely needs it, so you always know why a technique exists before you learn how it works.

## How Each Chapter Works

Most chapters follow the same rhythm: a concrete problem, the simplest implementation that solves it, the failure that implementation eventually meets, the concept that addresses the failure, a refactor, tests, and an honest look at the trade-offs. Concepts are never introduced as abstractions looking for a problem. The problem always comes first.

Two recurring features reinforce this.

**Engineering Lens.** Each chapter closes with five questions: what problem did we solve, what complexity did we introduce, what failure modes appeared, what would break at 10× scale, and when should we not use this technique. The last question matters most. A technique you cannot argue against is a technique you do not yet understand.

**Before and After.** Major chapters end with a problematic implementation next to a production-oriented one, followed by what the change gained and what it cost. The goal is judgment, not syntax.

## Who This Book Is For

You should be comfortable with basic programming: variables, functions, control flow, and reading someone else's code. You do not need prior Go experience, though developers coming from Python, Node, Java, or PHP will find the Go-specific material paced for them. Junior and mid-level backend engineers will find a systematic path from first service to production. Experienced developers can use the book as a structured refresher and a source of checklists.

## A Final Note Before We Begin

Near the end of the book you will meet one question that the whole text is designed to prepare you for:

> If this service breaks at 3 AM, can another engineer understand what happened?

Keep it in mind as you read. It is a better measure of a backend than any benchmark.

---

# Part I: Thinking in Backends

# Chapter 1: What Is a Backend, Really?

Imagine a person tapping a button labeled **Create Task**. A spinner appears for a moment, and then the new task shows up in a list. From the user's side, nothing could be simpler.

Now ask what had to be true for that to work. The request had to arrive. Someone had to decide whether it was well formed, whether the person was allowed to make it, and whether the task it described made sense. Something had to be written somewhere that survives a restart. Perhaps a notification had to go out. The user then needed an answer that told them, accurately, what happened.

All of that sits behind one tap. The backend is everything behind it.

## The Definition That Misleads

The most common definition is that a backend receives HTTP requests and returns responses. It is accurate and nearly useless, because it describes the packaging and says nothing about the work. It is like defining a restaurant as a place where plates move between a kitchen and tables.

It also quietly trains bad habits. If a backend is "a thing that handles requests," then the natural unit of design is the handler, and the natural place for logic is inside it. Chapter 6 will show where that road ends: a single function that parses input, applies rules, talks to the database, writes logs, and formats errors, all tangled together.

A more useful definition starts from what the system does rather than how it is reached.

> **A backend is a state-transforming system.** It takes input, decides whether and how that input should change the world, makes the change, and reports the outcome.

Notice what moved to the center: *state*, and *decisions about state*. HTTP is how the input arrives. It is not what the system is.

## The Anatomy of a Request

Every meaningful request, whether it arrives over HTTP, gRPC, or a message queue, passes through the same stages in some form.

```
Input
  ↓
Interpretation
  ↓
Validation
  ↓
Business decision
  ↓
State transition
  ↓
Side effects
  ↓
Response
```

Each stage answers a distinct question, and each fails in a distinct way.

| Stage | Question it answers | Typical failure |
| --- | --- | --- |
| Input | What exactly did we receive? | Truncated body, wrong content type, oversized payload |
| Interpretation | What is the caller trying to do? | Unknown fields, ambiguous meaning, malformed JSON |
| Validation | Is this acceptable on its face? | Missing title, negative quantity, invalid format |
| Business decision | Should this be allowed right now? | Task already completed, user lacks permission, quota exceeded |
| State transition | What changes, and is the result valid? | Half-written data, lost update, constraint violation |
| Side effects | What else must happen because of this? | Email not sent, event not published, cache left stale |
| Response | What do we tell the caller? | Success reported for a failure, or failure for a success |

The distinction between the middle stages deserves attention. **Validation** checks the request in isolation: is a title present, is the ID well formed? **A business decision** checks the request against the current state of the world: is this task allowed to move from `pending` to `completed` given what it is right now? The first needs only the input. The second needs the system's memory. Mixing them is one of the most common sources of confused code.

## Following One Request Through Forge

Take the first thing Forge will ever do. A client sends:

```
POST /tasks
{ "title": "Write chapter one" }
```

**Input and interpretation.** The server receives bytes, recognizes them as JSON, and extracts a title. At this point it knows what the caller said, not whether it is valid.

**Validation.** The title is present and within a sensible length. If it were missing, the request would be rejected here, cheaply, before anything expensive happens.

**Business decision.** Is this caller allowed to create tasks? Have they exceeded a limit? For now, Forge has no rules here, but a place for them already exists.

**State transition.** A new task is created in the `pending` state and stored. This is the moment the system's memory changes. Everything before it was thinking. This is acting.

**Side effects.** Perhaps an audit record is written, or an event is emitted for other parts of the system. These are consequences of the change that live outside the change itself.

**Response.** The caller learns the task's ID and its state. The response must reflect what truly happened, not what was intended.

Written out this way, the request looks orderly. The rest of this chapter is about everything that makes it less orderly in practice.

## State Is What Remains

A request is temporary. It arrives, is processed, and disappears. **State** is what remains afterward and shapes how future requests behave.

When Forge creates a task, the request is gone within milliseconds, but the task persists. The next request, perhaps from a different client on a different day, will find it. That persistence is what makes the system useful, and also what makes it dangerous, because the cost of an incorrect state transition is paid long after the request that caused it.

This leads to a question that Chapter 11 will treat in depth: *where does the state actually live, and which copy is the truth?* A task might exist in a database, in a cache, in a queue message, and in the client's memory at once. When they disagree, one of them is right, and the system must know which. For now, hold the idea that every backend is, at its core, a custodian of state, and its first duty is to keep that state valid.

## Business Rules and Invariants

A rule such as "a completed task cannot be started again" is not an implementation detail. It is a statement about what states and transitions are legal in the system. Rules of this kind, which must hold no matter how the system is reached, are called **invariants**.

Consider the lifecycle Forge will enforce:

```
pending
  │
  ▼
running
  │
  ├──────► failed
  │
  ▼
completed
```

This small diagram contains real decisions. Can a `failed` task be retried? Can a `pending` task jump straight to `completed`? Who is allowed to cancel? Each answer is a business rule, and each rule must be enforced somewhere that cannot be bypassed. If enforcement lives only in one handler, the next handler, background job, or script that touches tasks can break it silently.

Chapter 3 designs these rules before any code is written. Chapter 18 shows how to make the code itself refuse invalid transitions. Here the point is simpler: a backend exists to decide, and its decisions must be consistent.

## Dependencies and Side Effects

Almost no backend works alone. Forge will talk to a database, a cache, a message broker, and perhaps external services for email or payments. These are **dependencies**, and each one introduces something your code does not control: its speed, its availability, its behavior under load.

A **side effect** is any consequence of a request that happens outside the system's own state, such as sending an email, publishing an event, or charging a card. Side effects are especially important because they are often impossible to undo. A row can be rolled back. An email cannot be unsent.

Dependencies and side effects are where most real-world complexity enters. A function that only transforms data in memory is easy to reason about. A function that writes to a database, calls a payment provider, and publishes an event has three ways to fail partway through.

## Failure Is Ordinary

Return to the creation of a task and add one realistic detail: after storing the task, Forge sends a notification.

```
1. Store the task          → succeeds
2. Send the notification   → fails
3. Respond to the caller   → ???
```

What should step 3 say? If it reports success, the caller believes the notification went out when it did not. If it reports failure, the caller may retry, and now the task exists twice. If the process crashes between steps 1 and 2, no response is sent at all, and the system is left in a state nobody chose.

There is no clever trick that makes this go away. The lesson is a change of attitude: **partial failure is not an edge case, it is the normal condition of any system that does more than one thing.** Good backends are not those where nothing fails. They are those where every failure leaves the system in a state that is valid, understood, and recoverable. The second principle of this book, make failure ordinary, begins here.

## Concurrency: The World Does Not Wait Its Turn

Everything above assumed one request at a time. Real traffic does not cooperate.

Suppose two workers both read a task, find it `pending`, and decide to start it. Each is acting correctly given what it saw. Both then write `running`. The task has now been started twice, and neither worker did anything wrong. The bug is in the gap between reading and writing, and it appears only when timing aligns badly, which makes it rare in testing and common in production.

A backend serves many callers at once, so every piece of shared state is a place where two actors might collide. Part V treats concurrency in depth, but the habit begins now: whenever you read a decision from the current state and then act on it, ask what happens if someone else changes that state in between.

## Observability: Can You See What Happened?

Imagine one of those failures has occurred. A customer reports that their task is stuck. What can you learn?

If the system produced nothing but silence, you can only guess. If it recorded which request arrived, what decision was made, which dependency was slow, and where the error occurred, you can reconstruct the story. The difference between those two situations is **observability**: the property that lets you understand a system's internal behavior from the outside.

It is easy to treat this as an operational afterthought. It is better to treat it as a design requirement, because a backend you cannot inspect is one you cannot safely change. Part VI covers logs, metrics, and traces. For now, accept the idea that a production backend must be able to explain itself.

## The Eight Concerns

Over the course of this chapter, eight concerns have appeared. Together they form the real subject of backend engineering, and the rest of the book unfolds them in order.

| Concern | The core question | Where the book treats it |
| --- | --- | --- |
| Requests | How does input enter and get understood? | Part II |
| State | Where is the truth, and how is it kept valid? | Part III |
| Business rules | What decisions must always hold? | Part IV |
| Persistence | How does state survive restarts and crashes? | Part III |
| Dependencies | What do we rely on that we do not control? | Parts III, V, VI |
| Failures | What happens when something goes wrong halfway? | Part VI |
| Concurrency | What happens when many things happen at once? | Part V |
| Observability | Can we see and explain what the system did? | Part VI |

A handler touches the first of these and quietly depends on all the others. That is why a backend cannot be understood by studying handlers alone.

## The Key Idea

> **A backend is a state-transforming system.** Its job is to accept input, make correct decisions about state, change that state safely, and stay understandable when things go wrong.

Everything else in this book, from routing to Kafka to Kubernetes, is either a way to do that job at larger scale or a way to notice when it is going wrong.

## Engineering Lens

**What problem did we solve?** We replaced the shallow idea of a backend as a request handler with a model built around state, decisions, and consequences, and we gave ourselves eight concerns to organize everything that follows.

**What complexity did we introduce?** A larger vocabulary before any code. Readers who prefer to start typing will feel the delay. The vocabulary is worth it because every later technique attaches to one of these concerns.

**What failure modes appeared?** Partial completion across multiple steps, duplicate effects from retries, lost updates from concurrent access, and invisible failures from missing observability. Each reappears later with a concrete solution.

**What would break at 10× scale?** The concerns that seem theoretical at low traffic become the dominant ones. Concurrency collisions move from rare to routine, slow dependencies cascade, and the absence of observability turns every incident into guesswork.

**When should we not use this way of thinking?** For a throwaway script or a prototype that will be discarded in a week, this much rigor is overkill. The model pays off once the system holds state that matters and serves callers you cannot control.

## Exercises

1. Pick a backend you use daily, such as a food delivery app or a banking app. Trace one action through the seven stages. Where does validation end and the business decision begin?
2. For the Forge lifecycle shown above, list every possible transition between states and mark which you would allow. What new states would you add for cancelled tasks?
3. Write down two side effects of creating a task that cannot be rolled back. What would you do if the second one failed after the first succeeded?
4. Describe a scenario where two requests could corrupt a task's state if they arrived at the same moment. What would the system need to know to prevent it?

In the next chapter, we follow a single request at the level of the network and the code, from the first byte arriving to the last byte leaving, and learn to trace execution the way an experienced backend engineer does.