# DevTrail

**Understand how you build. Learn from what breaks. Become a better engineer.**

DevTrail is a developer-focused engineering intelligence platform designed to track, document and analyze the software development process.

It helps developers understand not only what they build, but also how they build it, the problems they encounter, the decisions they make and the lessons they learn along the way.

DevTrail combines automatic development activity tracking with manual engineering documentation to create a continuous record of software development.

> **Project status:** Initial planning and architecture definition.

---

## 1. The Problem

Software development involves much more than writing code.

Developers spend considerable time:
- Investigating unexpected application behaviour.
- Debugging errors and failed tests.
- Exploring different implementation approaches.
- Making architectural decisions.
- Refactoring existing code.
- Identifying performance and reliability issues.
- Understanding unfamiliar codebases.

However, much of this engineering experience is lost.

Git records changes to source code, but it does not adequately explain the reasoning behind those changes. Traditional task trackers document completed work but rarely preserve the investigation process that led to a solution.

As a result, developers frequently encounter previously solved problems, forget why particular technical decisions were made and struggle to evaluate their engineering progress.

DevTrail aims to address this gap.

## 2. Vision

To help developers continuously improve their engineering capabilities through structured development history, debugging intelligence and actionable engineering insights.

## 3. Mission

To transform everyday software development activities into structured, searchable and meaningful engineering knowledge.

---

## 4. Core Objectives

DevTrail aims to:

1. Capture development activities across multiple repositories.
2. Preserve the history of debugging investigations and their resolutions.
3. Document engineering decisions and their reasoning.
4. Track testing activities and application failures.
5. Identify recurring technical problems.
6. Analyze development patterns over time.
7. Generate actionable recommendations for engineering improvement.
8. Maintain a unified developer profile across projects.

---

## 5. Core Features

### 5.1 Project Management

Developers can register and manage multiple software projects.

Each project maintains its own development history.

Expected capabilities:
- Register local repositories.
- Associate repositories with projects.
- Track project activity.
- View project-specific development sessions.
- Explore historical engineering records.
- Generate project-specific reports.

### 5.2 Development Activity Tracking

DevTrail combines automatic activity collection with manual documentation.

Activities may include:
- Git commits.
- Branch changes.
- File modification metadata.
- Development session events.
- Build results.
- Test execution results.
- Debugging activities.

The system should provide a chronological view of development activities.

### 5.3 Development Sessions

A development session represents a period of focused work on a project.

A session may contain:
- Start and end timestamps.
- Associated project.
- Activities performed.
- Files modified.
- Commits created.
- Tests executed.
- Problems encountered.
- Session notes.

Developers should be able to start, pause, resume and end sessions.

### 5.4 Debugging Journal

The debugging journal preserves the complete problem-solving process.

Each investigation should capture:

| Field | Description |
|---|---|
| Title | Short description of the problem |
| Project | Associated project |
| Context | What the developer was working on |
| Error | Observed error or unexpected behaviour |
| Investigation | Diagnostic process |
| Attempts | Approaches explored |
| Root cause | Identified underlying cause |
| Resolution | Implemented solution |
| Lessons | Knowledge gained |
| Prevention | Potential preventive measures |

The system should support searching and connecting related investigations.

### 5.5 Engineering Decision Records

Engineering decisions should be documented alongside their associated projects.

Each decision should contain:
- Decision title.
- Problem or context.
- Alternatives considered.
- Selected approach.
- Reasoning.
- Consequences.
- Decision status.

This provides a historical record of how the software architecture evolved.

### 5.6 Testing History

DevTrail should maintain a history of test executions.

Expected information:
- Test command.
- Execution timestamp.
- Associated project.
- Execution result.
- Failed test information.
- Duration.
- Related development activity.

The system should help developers identify recurring test failures and understand how changes affect software reliability.

### 5.7 Engineering Intelligence

This is one of DevTrail's central capabilities.

The intelligence engine analyzes collected engineering records to identify patterns.

Examples:
- Frequently encountered problem categories.
- Repeated debugging investigations.
- Components associated with recurring failures.
- Frequently failing tests.
- Changes followed by regressions.
- Areas requiring additional testing.
- Potential engineering knowledge gaps.

Insights should be evidence-based, distinguish observations from assumptions and explain the reasoning behind recommendations.

### 5.8 Reports

DevTrail should generate reports at different levels.

**Daily reports**
- Development activities.
- Sessions completed.
- Problems encountered.
- Solutions implemented.
- Outstanding work.

**Weekly reports**
- Project activity.
- Debugging patterns.
- Testing history.
- Engineering decisions.
- Improvement opportunities.

**Project reports**
- Historical development activity.
- Engineering decisions.
- Debugging history.
- Testing trends.
- Project evolution.

Supported export formats should include Markdown and JSON.

---

## 6. System Architecture

DevTrail consists of three primary components.

### 6.1 CLI

The CLI interacts with the developer's local environment.

Responsibilities:
- Register local repositories.
- Capture Git activity.
- Manage development sessions.
- Collect supported build and test events.
- Record manual engineering notes.
- Synchronize collected information with the backend.

### 6.2 Backend

The backend provides central application services.

Responsibilities:
- Authentication and authorization.
- Project management.
- Activity ingestion.
- Development session management.
- Debugging record management.
- Engineering decision management.
- Testing history.
- Insight generation.
- Reporting.

### 6.3 Frontend

The frontend provides the interface for exploring development history.

It will use native web technologies without a frontend framework.

**Technology:**
- HTML5 for document structure.
- CSS3 for styling and responsive layouts.
- Vanilla JavaScript for interactions and API communication.

Responsibilities:
- Display development dashboards.
- Visualize project activity.
- Present debugging investigations.
- Display engineering decisions.
- Explore testing history.
- Present engineering insights.
- Generate and display reports.
- Manage user settings.

---

## 7. Proposed Architecture Diagram

```text
                    DEVELOPER
                        |
           +------------+------------+
           |                         |
        DEVTRAIL CLI            WEB APPLICATION
           |                         |
       LOCAL TRACKING          HTML + CSS + JS
           |                         |
        SQLITE                  REST API CLIENT
           |                         |
           +------------+------------+
                        |
                   GO BACKEND
                        |
                 EVENT PROCESSING
                        |
           +------------+------------+
           |            |            |
        Projects     Activities    Sessions
           |            |            |
           +------------+------------+
                        |
                   POSTGRESQL
                        |
                  ANALYSIS ENGINE
                        |
           +------------+------------+
           |            |            |
        Patterns     Insights     Reports
```

The CLI maintains local persistence using SQLite and synchronizes records with the central PostgreSQL database.

---

## 8. Technology Stack

| Component | Technology |
|---|---|
| Backend | Go |
| Frontend | HTML5 |
| Styling | CSS3 |
| Frontend interactions | Vanilla JavaScript |
| Database | PostgreSQL |
| Local storage | SQLite |
| API | REST |
| Version control | Git |
| Repository integration | GitHub API |
| Backend deployment | Render |
| Frontend deployment | Render Static Sites |

No frontend frameworks or UI libraries are required for the initial implementation.

---

## 9. Repository Structure

Proposed monorepo structure:

```text
devtrail/
│
├── backend/
│   ├── cmd/
│   │   └── api/
│   ├── internal/
│   │   ├── user/
│   │   ├── workspace/
│   │   ├── project/
│   │   ├── activity/
│   │   ├── session/
│   │   ├── git/
│   │   ├── debugging/
│   │   ├── decision/
│   │   ├── testing/
│   │   ├── insights/
│   │   └── reports/
│   ├── migrations/
│   ├── configs/
│   └── go.mod
│
├── frontend/
│   ├── index.html
│   │
│   ├── pages/
│   │   ├── login.html
│   │   ├── register.html
│   │   ├── overview.html
│   │   ├── activity.html
│   │   ├── projects.html
│   │   ├── sessions.html
│   │   ├── debugging.html
│   │   ├── decisions.html
│   │   ├── testing.html
│   │   ├── insights.html
│   │   ├── reports.html
│   │   └── settings.html
│   │
│   ├── css/
│   │   ├── main.css
│   │   ├── layout.css
│   │   ├── components.css
│   │   └── pages/
│   │
│   ├── js/
│   │   ├── app.js
│   │   ├── api.js
│   │   ├── auth.js
│   │   ├── components/
│   │   └── pages/
│   │
│   └── assets/
│       ├── icons/
│       └── images/
│
├── cli/
│   ├── cmd/
│   ├── internal/
│   └── go.mod
│
├── docs/
│   ├── architecture/
│   ├── decisions/
│   ├── requirements/
│   └── research/
│
└── README.md
```

---

## 10. Architectural Principles

### Event-driven activity recording

Relevant development activities should be represented as events.

Events should be preserved as historical records rather than overwritten whenever possible.

### Separation of concerns

The CLI, backend and frontend should have clearly defined responsibilities.

The frontend should handle presentation and user interactions, while the backend owns business logic and data processing.

### Modular architecture

Each backend module should have a clearly defined responsibility and explicit boundaries.

### Lightweight frontend

The frontend should use native HTML, CSS and JavaScript.

Reusable components should be implemented without introducing unnecessary framework dependencies.

### Local-first data collection

The CLI should be capable of collecting and temporarily storing information locally, even when synchronization is unavailable.

### Privacy by design

DevTrail should not collect source code, credentials, environment secrets or sensitive file contents by default.

Developers should control what information is captured and synchronized.

### Evidence-based intelligence

Recommendations should be based on recorded evidence.

The system must distinguish observed facts from inferred explanations.

---

## 11. Development Roadmap

Development will take place incrementally.

| Phase | Focus | Expected Outcome |
|---|---|---|
| 1 | Requirements and architecture | Defined system boundaries |
| 2 | Project management | Repository registration |
| 3 | Activity tracking | Structured development events |
| 4 | Session management | Development session history |
| 5 | Git integration | Repository activity history |
| 6 | Debugging journal | Structured investigations |
| 7 | Engineering decisions | Decision documentation |
| 8 | Testing history | Test execution records |
| 9 | Reporting | Historical engineering reports |
| 10 | Intelligence engine | Pattern detection and recommendations |
| 11 | CLI synchronization | Connected local and central systems |
| 12 | Advanced analysis | Intelligent engineering assistance |

---

## 12. Initial MVP Scope

The first usable version should focus on:

- User registration and authentication.
- Project registration.
- Local repository identification.
- Development session tracking.
- Git activity collection.
- Manual activity documentation.
- Debugging investigations.
- Activity timeline.
- Basic project reports.

Advanced analytics and AI-powered recommendations should come after the underlying data collection has been validated.

---

## 13. Privacy and Security

Because DevTrail interacts with development environments, security is a fundamental requirement.

The system should:
- Exclude credentials and environment secrets.
- Avoid uploading source code by default.
- Allow developers to configure tracked repositories.
- Provide controls over collected information.
- Authenticate synchronization requests.
- Protect user and project data.
- Support local data retention and deletion.
- Maintain clear audit records for sensitive operations.

---

## 14. Long-term Direction

DevTrail should eventually evolve into an engineering intelligence platform.

Potential future capabilities include:
- AI-assisted debugging analysis.
- Automatic correlation of code changes and test failures.
- Recurring bug detection.
- Engineering knowledge retrieval.
- Cross-project development analysis.
- Personalized learning recommendations.
- Team engineering analytics.
- Integration with development environments and CI/CD platforms.

The long-term goal is to help developers learn from their actual engineering experiences rather than relying exclusively on external tutorials and documentation.

---

## 15. Guiding Principle

> Don't just track the code. Understand the engineering behind it.

DevTrail is built around the belief that every debugging investigation, technical decision, failed test and successful implementation contains knowledge worth preserving.

**Build. Investigate. Understand. Improve.**
