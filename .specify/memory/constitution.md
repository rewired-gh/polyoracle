<!--
Sync Impact Report - Constitution Update
==========================================
Version change: N/A → 1.0.0
Modified principles: Initial creation
Added sections:
  - Core Principles (5 principles)
  - Technology Stack
  - Development Workflow
  - Governance
Removed sections: None
Templates requiring updates:
  ✅ .specify/templates/plan-template.md (reviewed - compatible)
  ✅ .specify/templates/spec-template.md (reviewed - compatible)
  ✅ .specify/templates/tasks-template.md (reviewed - compatible)
Follow-up TODOs: None
-->

# Poly Oracle Constitution

## Core Principles

### I. Simplicity and Maintainability

Code MUST be simple, readable, and maintainable. Every design decision MUST prioritize clarity over cleverness.

**Rules:**
- Favor explicit over implicit implementations
- Avoid premature abstraction - prefer duplication when the abstraction isn't clear
- Each function and module MUST have a single, clear responsibility
- Code MUST be self-documenting through clear naming and structure
- Comments MUST explain "why", not "what"

**Rationale:** Simple code is easier to understand, debug, and modify. Maintainability directly impacts long-term project velocity and reduces technical debt.

### II. Go Language

All production code MUST be written in idiomatic Go, following the official Go style guidelines and best practices.

**Rules:**
- Use the latest stable Go version unless project constraints require otherwise
- Follow Effective Go guidelines and Go Code Review Comments
- Use standard library preferentially; external dependencies require justification
- Leverage Go's concurrency primitives (goroutines, channels) appropriately
- Ensure all code passes `go vet` and `golint` without warnings

**Rationale:** Go provides simplicity, performance, and strong tooling. Idiomatic Go ensures consistency across the codebase and leverages the language's strengths.

### III. Latest and Robust Dependencies

External dependencies MUST be carefully evaluated, actively maintained, and minimal in number.

**Rules:**
- Dependencies MUST use semantic versioning with pinned versions in go.mod
- Evaluate dependencies for: maintenance activity, community adoption, security history, and API stability
- Avoid dependencies that duplicate standard library functionality
- Document the rationale for each major dependency in project documentation
- Regularly update dependencies to incorporate security patches
- Prefer well-established libraries over experimental ones

**Rationale:** External dependencies introduce risk and maintenance burden. Careful selection and minimal dependencies reduce attack surface and dependency management overhead.

### IV. Comprehensive Unit Testing

All packages MUST have reasonable unit test coverage with meaningful test cases.

**Rules:**
- Write unit tests for all exported functions and types
- Test coverage MUST include happy paths, edge cases, and error conditions
- Use table-driven tests for scenarios with multiple inputs/outputs
- Tests MUST be independent and idempotent - no shared state between tests
- Aim for high coverage of critical business logic; coverage targets should be pragmatic
- Use `go test` with benchmarking where performance is critical
- Integration and end-to-end tests complement but do not replace unit tests

**Rationale:** Unit tests catch bugs early, document expected behavior, and enable confident refactoring. Reasonable coverage balances quality assurance with development velocity.

### V. Code Quality and Taste

Code MUST demonstrate good taste through clarity, consistency, and attention to detail.

**Rules:**
- Follow consistent formatting enforced by `gofmt` and project linting rules
- Use meaningful names that convey intent - avoid abbreviations and single-letter variables except in obvious contexts
- Functions should be small and focused - if it needs a comment to explain what it does, it should be refactored
- Handle errors explicitly - never ignore errors, always propagate or handle appropriately
- Review code with fresh eyes - if it's hard to understand, rewrite it
- Eliminate dead code, commented-out code, and unnecessary complexity
- Apply the "leave the code better than you found it" principle

**Rationale:** Good code taste is the discipline of writing code that others (and future you) can understand and maintain. It reduces cognitive load and prevents bugs.

## Technology Stack

**Language:** Go (latest stable version)

**Dependency Management:** Go modules with vendoring for reproducible builds

**Testing Framework:** Standard `testing` package with testify for assertions (if needed)

**Linting:** golangci-lint with project-specific configuration

**Documentation:** Go doc comments and project-level documentation in docs/

**Build Tools:** Standard Go toolchain with Make for build automation

## Development Workflow

### Code Review Requirements

- All code changes MUST be reviewed before merging
- Reviewers MUST verify constitution compliance
- PRs MUST pass all automated tests and linting checks
- PRs MUST include tests for new functionality

### Testing Gates

- All unit tests MUST pass before merge
- `go vet` and linting checks MUST pass
- Critical paths should have integration tests where applicable
- Performance benchmarks MUST not regress without documented justification

### Commit Standards

- Commits MUST be atomic and focused
- Commit messages MUST clearly describe the change and motivation
- Reference issues/PRs where applicable

### Refactoring Guidelines

- Refactoring MUST be done in separate commits from feature changes
- Behavior-preserving refactorings MUST include test verification
- Large refactorings require plan documentation

## Governance

This constitution establishes the non-negotiable principles for the Poly Oracle project. All decisions, code reviews, and architectural choices MUST align with these principles.

**Amendment Procedure:**
- Amendments require documentation of the proposed change
- Impact analysis on existing code and practices
- Team discussion and consensus
- Version increment following semantic versioning

**Versioning Policy:**
- MAJOR version: Backward incompatible principle removal or redefinition
- MINOR version: New principle added or materially expanded guidance
- PATCH version: Clarifications, typo fixes, non-semantic refinements

**Compliance Review:**
- All PRs MUST explicitly verify constitution compliance
- Complexity beyond these principles requires documented justification
- Exceptions are temporary and MUST be tracked for resolution

**Version**: 1.0.0 | **Ratified**: 2026-02-16 | **Last Amended**: 2026-02-16
