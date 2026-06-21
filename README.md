<!--
Realized by @luckignolo32 (GitHub) - MIT License
-->

# SQL Injection — Attack diary and applied countermeasures

This repository contains the deliverable for the **Information Security and
Cybersecurity** (SIC) course. The project starts from the academic codebase
*Fantastic Coffee (decaffeinated)* — a messaging application with a Go
backend (`net/http` + `httprouter` + SQLite) and a Vue.js frontend served by
Vite — and uses it as the target of a SQL Injection chain executed on the
`vulnerabile` branch and subsequently remediated on the `master` branch.

This README summarises **how the attack was carried out** and **which
countermeasures were applied** so that it can no longer be reproduced. The
full attack diary, including every payload and every response body observed,
is kept in the separate file `Diario_di_una_SQL_Injection.md`.

---

## 1. Target architecture

| Component | Technology | Relevant detail |
|---|---|---|
| Frontend | Vue.js + Vite | SPA on `http://localhost:5173` |
| Backend  | Go (`database/sql`, `httprouter`) | JSON API on `http://localhost:3000` |
| Database | SQLite (`./data/decaf.db`) | Schema with tables `User`, `UserUsername`, `Login`, `Conversation`, `Message`, `Comment`, ... |
| Endpoint under attack | `POST /session?isLogin=true` | The only surface exposed before authentication |

The schema relevant to the attack is the following:

```sql
CREATE TABLE User(
    userId   INTEGER PRIMARY KEY AUTOINCREMENT,
    password VARCHAR(40) NOT NULL
);

CREATE TABLE UserUsername(
    updateId INTEGER PRIMARY KEY AUTOINCREMENT,
    time     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    userId   INTEGER REFERENCES User(userId) NOT NULL,
    username TEXT NOT NULL
);
```

The *current* username of a user is the one associated with the maximum
`time` for the corresponding `userId`. The login query in the vulnerable
version therefore joined `User` and `UserUsername` and verified the password
at the application layer.

---

## 2. The attack (summary)

The attack was conducted **without any prior knowledge** of the stack, the
DBMS or the source code: every piece of information was obtained through
active reconnaissance against the single public login endpoint. The three
canonical SQL Injection techniques — **tautology**, **comment termination**
and **piggybacked queries** — were used in sequence to achieve the goals
listed below.

### 2.1 Recon and DBMS fingerprinting

By inserting a single quote in the `name` field, the backend returned the
engine error verbatim:

```
{"error":"near \"adfadfadsf\": syntax error"}
```

The signature `near "X": syntax error` made it possible to fingerprint
**SQLite** and to reconstruct the shape of the server-side query, while at
the same time revealing an **information disclosure** vulnerability (SQL
parsing errors were leaked verbatim in the response).

### 2.2 Authentication bypass via `UNION SELECT`

After confirming that the classic `' OR 1=1 -- ` was not enough (the
password was also checked at the application layer), a `UNION SELECT` was
injected so as to produce a row with an attacker-controlled password value.
To neutralise the residual `AND` clause of the original query — placed on a
separate line in the source and therefore not covered by `--` — a multi-line
comment `/*` was used instead. The winning payload was:

```text
name     = ' UNION SELECT 1 AS password /*
password = x
```

Backend response: `HTTP/1.1 201 Created` with a valid `userId` and session
token.

### 2.3 Error-based exfiltration channel

Replacing the constant with a sub-query revealed that the Go backend typed
`userId` as `int`. Returning a non-numeric string in the `UNION` caused the
`Scan` of `database/sql` to fail with an error message that included
**verbatim the value read from the database**:

```
sql: Scan error on column index 0, name "userId":
converting driver.Value type string ("Comment") to a int: invalid syntax
```

This channel turned any arbitrary `SELECT` into an in-band exfiltration
primitive. A `group_concat` over `sqlite_master` returned the whole schema
in a single request; a `User`-`UserUsername` join exfiltrated every username
together with the corresponding password (some of which were stored in
plaintext).

### 2.4 Piggybacked queries: Integrity and Availability violations

After verifying that the SQLite driver accepted multiple statements separated
by `;`, DML operations were appended to the injection payload. The password
of `alice_dev` was surgically overwritten:

```text
name = '; UPDATE User SET password='pwned_by_sqli' WHERE userId=1 /*
```

and a **legitimate login** was subsequently performed with the new
credentials (indistinguishable, from the application logs, from a normal
sign-in). In the same way, the two rows corresponding to user
`davidedecazzo` were deleted from `UserUsername` and `User`.

### 2.5 CIA summary

| CIA property | Technique used | Outcome |
|---|---|---|
| **Confidentiality** | `UNION SELECT` + error-based via Go cast to `int` | Exfiltration of the schema and of every user credential |
| **Integrity** | Piggybacked `UPDATE` | Surgical password rewrite + subsequent legitimate login |
| **Availability** | Piggybacked `DELETE` | Removal of a user from `User` and `UserUsername` |

For the full step-by-step history (including failed payloads and the
diagnostic reasoning) refer to the file `Diario_di_una_SQL_Injection.md`.

---

## 3. Applied countermeasures

The `master` branch contains the remediated version. All changes are aimed
at removing the vulnerability **at its root** (not just sanitising the
symptom): the query is no longer built by string concatenation, the
separation between authentication and registration is now explicit, and the
error messages returned to the client no longer leak details about the SQL
engine.

### 3.1 Parameterised prepared statements

The login query has been rewritten so that every user-supplied input is
bound through a **`?` placeholder**: the argument is passed to the driver
as a bound parameter, never concatenated into the SQL string. The grammar
of the query is therefore fixed statically at compile time and no input can
alter its structure.

`service/database/CreateSession.go`:

```go
query := `SELECT u.userId, u.password
          FROM User u
          JOIN UserUsername uu ON u.userId = uu.userId
          WHERE uu.username = ?
          ORDER BY uu.updateId DESC
          LIMIT 1`

err := db.c.QueryRow(query, username).Scan(&userId, &password)
```

The same pattern is applied to `INSERT INTO User`, `INSERT INTO
UserUsername` and `INSERT INTO Login`. Direct consequences:

- the tautology `' OR 1=1 -- ` no longer alters the `WHERE` clause;
- the `UNION SELECT` is no longer parsed as part of the statement;
- piggybacked queries (`; UPDATE ...`, `; DELETE ...`) are no longer
  executed, because the argument is a value rather than SQL.

### 3.2 Splitting `/session` (login) and `/register` (signup)

In the vulnerable version the same endpoint `POST /session?isLogin=true`
served both login and registration, branching on a query-string flag and on
two code paths inside the same DB function. This overlap caused the login
branch to issue a concatenated `SELECT` even when the user did not exist.

`master` introduces two distinct endpoints:

| Endpoint | Handler | DB function |
|---|---|---|
| `POST /session`  | `service/api/session.go::postSession`  | `database.CreateSession` |
| `POST /register` | `service/api/register.go::postRegister` | `database.CreateUser`    |

Each path has a clear responsibility, its own sentinel errors
(`"password errata"`, `"username già esistente"`) and no branching on the
query string.

### 3.3 Suppression of the information-disclosure channel

In the vulnerable version, the session handler did:

```go
_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
```

returning **any** error to the client — including SQLite parser errors and
`Scan` errors from `database/sql`. This is the channel that made DBMS
fingerprinting and error-based exfiltration possible.

`master` filters known cases explicitly and masks everything else behind
generic responses:

```go
if err.Error() == "password errata" {
    w.WriteHeader(http.StatusConflict)
    _ = json.NewEncoder(w).Encode(map[string]string{"error": "Password errata"})
    return
}

context.Logger.WithError(err).Error("Error creating session")
http.Error(w, "Failed to create session", http.StatusInternalServerError)
```

As a result:

- SQL parser errors no longer reach the client;
- the attacker can no longer tell "user does not exist" apart from
  "SQL error" by looking at the response body;
- `Scan` errors (the error-based channel that relied on the `int` cast) are
  no longer visible from the outside.

### 3.4 Operational confirmation

Replaying the payloads from the diary against the `master` branch:

- `' UNION SELECT 1 AS password /*` no longer yields HTTP 201, but HTTP 409
  with `"Password errata"` (the `UNION` is treated as a literal string in
  the `name` field, which matches no existing username);
- `'; UPDATE User SET password='pwned_by_sqli' WHERE userId=1 /*` no longer
  modifies any row;
- single-quote injections no longer produce any `near "...": syntax error`
  messages.

---

## 4. Project structure

* `cmd/` — executables (web server and healthcheck entrypoints);
* `service/api/` — HTTP handlers, now with `session.go` and `register.go`
  cleanly separated;
* `service/database/` — DB middleware, with `CreateSession.go` and
  `CreateUser.go` rewritten on top of prepared statements;
* `webui/` — Vue.js + Vite SPA;
* `doc/` — OpenAPI specification of the API (the **full attack diary** is
  the separate document `Diario_di_una_SQL_Injection.md`);
* `data/decaf.db` — persistent SQLite database (created on first start).

---

## 5. Build & run

Backend (development):

```shell
go run ./cmd/webapi/
```

Build with the WebUI embedded:

```shell
./open-node.sh
yarn run build-embed
exit
go build -tags webui ./cmd/webapi/
```

Frontend in dev mode (in a second terminal):

```shell
./open-node.sh
yarn run dev
```

### Database persistence

The SQLite file is written by default to `./data/decaf.db` relative to the
working directory. The path can be overridden via the `--db-filename` flag
or the `CFG_DB_FILENAME` environment variable:

```shell
go run ./cmd/webapi/ --db-filename ./data/custom.db
```

In Docker, mount a volume on `/app/data` to preserve the DB across restarts:

```shell
docker run -p 3000:3000 -v "$(pwd)/data:/app/data" <image>
```

---

## 6. License

Distributed under the MIT License. See [LICENSE](LICENSE).
