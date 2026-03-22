# GoTwitter Improvements Proposal

## Goal

This document outlines the highest-value improvements for the current GoTwitter API based on the existing implementation in this repository. The focus is on improving security, API quality, performance, maintainability, and product depth.

## Current State Summary

The project already supports:

- user signup, login, logout, and profile CRUD
- tweet creation, listing, retrieval, update, and deletion
- automatic hashtag extraction and tag association
- tag listing, popularity lookup, detail view, and deletion

The API has a solid base, but several important areas should be improved before the project is treated as production-ready.

## Priority Overview

| Priority | Area | Why it matters |
| :--- | :--- | :--- |
| P0 | Security and authorization | Some sensitive routes are exposed without ownership checks |
| P0 | Data consistency | Tweet and hashtag writes are not transactional |
| P1 | API correctness and DX | Responses lack metadata and validation is inconsistent |
| P1 | Performance | Tweet listing triggers repeated tag lookups |
| P1 | Testing | There are currently no automated tests in the repo |
| P2 | Product features | The app is missing core social features beyond CRUD |
| P2 | Observability and operations | Logging and runtime diagnostics are minimal |

## 1. Security and Authorization

### Problems observed

- `PUT /users/{id}` and `DELETE /users/{id}` are not protected by authentication middleware.
- There is no ownership check to ensure a user can only update or delete their own account.
- `DELETE /tags/{id}` is exposed without authentication or role-based control.
- Authentication uses JWT cookies, but the API does not include CSRF protection for state-changing requests.

### Improvements

- Require authentication for all user update and delete operations.
- Enforce `claims.UserID == route user id` for user self-service actions.
- Restrict tag deletion to admins, or remove public tag deletion entirely.
- Add CSRF protection if cookie-based auth remains the main session mechanism.
- Add tighter cookie configuration guidance for production, including `Secure=true` and environment-specific settings.

### Expected outcome

- Users cannot modify or delete other user accounts.
- Destructive endpoints are no longer publicly accessible.
- The auth model becomes safer for browser-based clients.

## 2. Data Consistency and Integrity

### Problems observed

- Tweet creation writes the tweet first, then creates and associates tags in separate operations.
- Tweet updates delete all old tag associations and recreate them without a transaction.
- Failures during tag creation or association are silently ignored.
- Deleting tweets can leave unused tags behind.

### Improvements

- Wrap tweet creation and tweet update plus hashtag association in a single database transaction.
- Stop swallowing tag-association errors; return a proper service error when persistence is incomplete.
- Add cleanup logic for orphaned tags, either synchronously or via a background cleanup job.
- Add unique constraints and foreign-key rules where missing to protect integrity at the database level.

### Expected outcome

- A tweet and its hashtags are saved consistently.
- Partial writes become much less likely.
- Tag data stays clean over time.

## 3. API Design and Developer Experience

### Problems observed

- List endpoints support pagination inputs but do not return pagination metadata.
- Query parsing silently falls back to defaults when invalid values are passed.
- The API mixes raw slices and ad hoc maps in responses instead of using stable response DTOs.
- Update handlers fetch updated records after writing and ignore fetch errors.

### Improvements

- Return pagination metadata such as `page`, `page_size`, `count`, and optionally `total`.
- Validate query params and return `400` for invalid input instead of silently coercing values.
- Introduce response DTOs for tweets, tags, and list payloads.
- Standardize success and error payload contracts across the API.
- Consider versioning the API once the contract stabilizes.

### Expected outcome

- Clients can paginate reliably.
- API behavior becomes more predictable.
- Frontend and external integrations become easier to build.

## 4. Performance Improvements

### Problems observed

- `ListTweets` loads tags with one extra repository call per tweet, creating an N+1 query pattern.
- Tag detail responses fetch tweets, but those tweets do not appear to include their tags.
- Popular tag and list endpoints may become slow without supporting indexes as data grows.

### Improvements

- Replace per-tweet tag lookups with a bulk tag fetch or a joined query strategy.
- Add indexes for frequent filters such as `tweets.user_id`, `tags.name`, and the tweet-tag join table.
- Consider cursor pagination for large tweet feeds.
- Add request timeouts and database timeout handling for slower queries.

### Expected outcome

- Tweet listing scales better.
- Response latency is more predictable.
- The API remains usable as data volume increases.

## 5. Testing and Quality Gates

### Problems observed

- The repository currently has no test files.
- Critical logic such as auth, tweet ownership, hashtag extraction, and pagination defaults is unverified.

### Improvements

- Add unit tests for services, especially:
  - user creation and login
  - tweet ownership checks
  - hashtag extraction and deduplication
  - pagination and filtering rules
- Add repository integration tests against a test database.
- Add handler tests for auth-protected routes and validation failures.
- Run tests in CI on every push and pull request.

### Expected outcome

- Regressions are caught earlier.
- Refactors become safer.
- Security-sensitive behavior is continuously verified.

## 6. Product-Level Feature Improvements

The current API is functional but still closer to a CRUD backend than a social product. The next feature set should deepen the product model.

### Recommended product features

- likes and retweets
- replies and conversation threads
- follow and unfollow relationships
- user timeline and home feed endpoints
- mentions and notifications
- media attachments
- pinned tweets
- soft delete or archive support

### Best next feature

If only one product feature should be built next, implement follows plus a home timeline. That moves the project from isolated tweet CRUD into an actual social experience.

## 7. Observability and Operations

### Problems observed

- Logging is minimal and not structured.
- There is no request tracing, metrics, or health diagnostics beyond the basic API shape.
- There is no visible rate limiting or abuse protection.

### Improvements

- Add structured request logging with request IDs.
- Add metrics for request counts, latency, DB failures, and auth failures.
- Add rate limiting for login and tweet creation endpoints.
- Add a health endpoint that checks database connectivity.
- Add environment-based config validation at startup.

### Expected outcome

- Production issues are easier to detect and debug.
- Abuse and brute-force behavior are reduced.
- Deployments become safer to operate.

## Suggested Delivery Plan

### Phase 1: Stabilize the backend

- protect user update and delete routes
- add ownership checks for user mutations
- protect or remove tag deletion
- add transaction handling for tweet and tag writes
- stop ignoring repository errors
- add tests for auth and tweet flows

### Phase 2: Improve API quality

- add pagination metadata
- validate query parameters strictly
- remove N+1 tag loading in tweet lists
- introduce response DTOs for stable contracts

### Phase 3: Expand the product

- add follow relationships
- add home timeline endpoint
- add replies
- add likes and notifications

## Recommended First Sprint

The best immediate sprint would focus on shipping the following:

1. secure user and tag mutation routes
2. make tweet plus hashtag persistence transactional
3. add tests for auth, tweet ownership, and hashtag extraction
4. optimize tweet listing to avoid one tag query per tweet

This gives the largest gain in correctness and production readiness without changing the public feature set too aggressively.
