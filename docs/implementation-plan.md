# Release Catalog 実装計画書

## Phase 1 & 2 対象

- **Phase 1**: Branch / Release 基盤（Product, Main/Release Line, Current, Snapshot, Root Project, Hash, Source Revision, DT Adapter, OpenAPI）
- **Phase 2**: BOM-Link / Graph（Root SBOM 取得, BOM-Link 解析/再帰解決, 未解決検出, Graph API, 循環参照対策）

---

## 1. 技術スタック

| 項目 | 選定 |
|------|------|
| 言語 | Go 1.22+ |
| Web Framework | Echo v4 |
| DB (デフォルト) | SQLite (CGo: mattn/go-sqlite3) |
| DB 抽象化 | Repository Interface パターン（将来 PostgreSQL 等に差し替え可能） |
| マイグレーション | golang-migrate/migrate |
| OpenAPI | oapi-codegen（スキーマファースト） |
| DI | 手動 Wire / コンストラクタインジェクション |
| テスト | Go 標準 testing + testify |
| HTTP Client (DT) | net/http + 独自 Adapter |
| ログ | slog (Go 1.21+ 標準) |
| コンテナ | Docker / Docker Compose |
| Dependency-Track | v5.0 (2026-06 GA) — 別起動 |

---

## 2. Dependency-Track 環境

### 2.1 DT 起動方法（参考）

```yaml
# docker-compose.dt.yml (Release Catalog とは別管理)
services:
  dependency-track-api:
    image: dependencytrack/apiserver:5.0
    ports:
      - "8081:8080"
    environment:
      ALPINE_DATABASE_MODE: external
      ALPINE_DATABASE_URL: jdbc:postgresql://dt-postgres:5432/dtrack
      ALPINE_DATABASE_DRIVER: org.postgresql.Driver
      ALPINE_DATABASE_USERNAME: dtrack
      ALPINE_DATABASE_PASSWORD: dtrack
    depends_on:
      - dt-postgres

  dependency-track-frontend:
    image: dependencytrack/frontend:5.0
    ports:
      - "8082:8080"
    environment:
      API_BASE_URL: http://localhost:8081

  dt-postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: dtrack
      POSTGRES_USER: dtrack
      POSTGRES_PASSWORD: dtrack
    volumes:
      - dt-pgdata:/var/lib/postgresql/data

volumes:
  dt-pgdata:
```

開発時は DT スタブを利用するため、実 DT インスタンスは不要。結合テスト時のみ上記で起動する。

---

## 3. プロジェクト構成

```
release-catalog/
├── cmd/
│   └── server/
│       └── main.go              # エントリポイント
├── internal/
│   ├── config/
│   │   └── config.go            # 環境変数・設定読み込み
│   ├── domain/
│   │   ├── product.go           # Product エンティティ
│   │   ├── branch.go            # BranchLine エンティティ
│   │   ├── current_state.go     # BranchCurrentState
│   │   ├── snapshot.go          # Snapshot (Main/Release)
│   │   ├── graph.go             # Graph ノード/エッジ
│   │   └── errors.go            # ドメインエラー定義
│   ├── repository/
│   │   ├── interfaces.go        # Repository インターフェース群
│   │   └── sqlite/
│   │       ├── product.go
│   │       ├── branch.go
│   │       ├── current_state.go
│   │       ├── snapshot.go
│   │       └── migrations/
│   │           ├── 001_initial.up.sql
│   │           └── 001_initial.down.sql
│   ├── service/
│   │   ├── product_service.go
│   │   ├── branch_service.go
│   │   ├── release_service.go
│   │   └── graph_service.go     # Phase 2
│   ├── dtclient/
│   │   ├── client.go            # DependencyTrackClient インターフェース
│   │   ├── http_client.go       # 実 DT HTTP 実装
│   │   ├── stub_client.go       # 開発用スタブ
│   │   └── models.go            # DT レスポンスモデル
│   ├── bomlink/
│   │   ├── resolver.go          # BOM-Link 解析・再帰解決 (Phase 2)
│   │   ├── parser.go            # CycloneDX BOM-Link 抽出
│   │   └── graph_builder.go     # Graph 構築
│   ├── handler/
│   │   ├── product_handler.go
│   │   ├── branch_handler.go
│   │   ├── current_handler.go
│   │   ├── snapshot_handler.go
│   │   ├── graph_handler.go     # Phase 2
│   │   ├── error_handler.go     # 統一エラーレスポンス
│   │   └── middleware/
│   │       ├── request_id.go
│   │       ├── logging.go
│   │       └── auth_placeholder.go  # 将来認証用スケルトン
│   └── openapi/
│       └── spec.yaml            # OpenAPI 3.1 仕様
├── migrations/
│   └── sqlite/
│       ├── 001_initial.up.sql
│       └── 001_initial.down.sql
├── api/
│   └── openapi.yaml             # OpenAPI 仕様 (公開用コピー)
├── docker-compose.yml           # Release Catalog 起動用
├── docker-compose.dt.yml        # DT 参考起動用
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── .env.example
└── docs/
    ├── implementation-plan.md   # 本文書
    └── architecture.md          # アーキテクチャ説明
```

---

## 4. DB 抽象化設計

### 4.1 Repository Interface

```go
// internal/repository/interfaces.go

type ProductRepository interface {
    Create(ctx context.Context, p *domain.Product) error
    FindByID(ctx context.Context, id string) (*domain.Product, error)
    FindAll(ctx context.Context, opts ListOptions) ([]*domain.Product, error)
    Update(ctx context.Context, p *domain.Product) error
    Delete(ctx context.Context, id string) error
}

type BranchRepository interface {
    Create(ctx context.Context, b *domain.BranchLine) error
    FindByID(ctx context.Context, id string) (*domain.BranchLine, error)
    FindByProductID(ctx context.Context, productID string) ([]*domain.BranchLine, error)
    Update(ctx context.Context, b *domain.BranchLine) error
}

type CurrentStateRepository interface {
    Upsert(ctx context.Context, cs *domain.BranchCurrentState) error
    FindByBranchID(ctx context.Context, branchID string) (*domain.BranchCurrentState, error)
}

type SnapshotRepository interface {
    Create(ctx context.Context, s *domain.Snapshot) error
    FindByID(ctx context.Context, id string) (*domain.Snapshot, error)
    FindByBranchID(ctx context.Context, branchID string, opts ListOptions) ([]*domain.Snapshot, error)
}
```

### 4.2 将来の DB 切り替え

```
internal/repository/
├── interfaces.go          # 共通インターフェース
├── sqlite/                # SQLite 実装
│   └── ...
└── postgres/              # 将来追加
    └── ...
```

起動時に `config.DatabaseDriver` を見て実装を切り替える。

---

## 5. DT Adapter 設計

### 5.1 インターフェース

```go
// internal/dtclient/client.go

type DependencyTrackClient interface {
    // Project
    GetProject(ctx context.Context, uuid string) (*DTProject, error)
    ProjectExists(ctx context.Context, uuid string) (bool, error)

    // SBOM
    GetBOM(ctx context.Context, projectUUID string) (*CycloneDXBOM, error)

    // Vulnerability (Phase 3 だが Interface は定義)
    GetVulnerabilities(ctx context.Context, projectUUID string) ([]*DTVulnerability, error)
}
```

### 5.2 スタブ実装

開発・テスト時は `StubClient` を利用。設定ファイルまたはインメモリデータで応答を返す。

```go
// internal/dtclient/stub_client.go

type StubClient struct {
    Projects map[string]*DTProject
    BOMs     map[string]*CycloneDXBOM
}
```

### 5.3 実クライアント

```go
// internal/dtclient/http_client.go

type HTTPClient struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
    timeout    time.Duration
}
```

DT v5.0 の REST API エンドポイントに対応:
- `GET /api/v1/project/{uuid}`
- `GET /api/v1/bom/cyclonedx/project/{uuid}`
- `GET /api/v1/vulnerability/project/{uuid}`

---

## 6. データモデル（SQLite DDL）

```sql
-- 001_initial.up.sql

CREATE TABLE product (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE branch_line (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('MAIN', 'RELEASE')),
    name TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    source_branch_line_id TEXT REFERENCES branch_line(id),
    forked_from_snapshot_id TEXT,
    status TEXT NOT NULL DEFAULT 'active'
        CHECK(status IN ('active','maintenance','security_only','end_of_support','closed')),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    closed_at TEXT,
    UNIQUE(product_id, name)
);

CREATE TABLE branch_current_state (
    branch_line_id TEXT PRIMARY KEY REFERENCES branch_line(id) ON DELETE CASCADE,
    root_dt_project_uuid TEXT,
    root_bom_serial_number TEXT,
    root_bom_version INTEGER,
    root_bom_sha256 TEXT,
    source_revision TEXT,
    updated_at TEXT NOT NULL
);

CREATE TABLE snapshot (
    id TEXT PRIMARY KEY,
    branch_line_id TEXT NOT NULL REFERENCES branch_line(id) ON DELETE CASCADE,
    snapshot_type TEXT NOT NULL CHECK(snapshot_type IN ('MAIN_SNAPSHOT', 'RELEASE')),
    version TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK(status IN ('draft','testing','approved','released','deprecated','end_of_support')),
    root_dt_project_uuid TEXT,
    root_bom_serial_number TEXT,
    root_bom_version INTEGER,
    root_bom_sha256 TEXT,
    source_revision TEXT,
    created_at TEXT NOT NULL,
    released_at TEXT
);

-- Phase 2: BOM-Link Index (性能最適化用、正本ではない)
CREATE TABLE bom_link_index (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_dt_project_uuid TEXT NOT NULL,
    source_bom_serial TEXT,
    source_bom_version INTEGER,
    source_bom_ref TEXT,
    target_serial TEXT,
    target_bom_version INTEGER,
    target_bom_ref TEXT,
    target_dt_project_uuid TEXT,
    resolution_status TEXT NOT NULL DEFAULT 'pending'
        CHECK(resolution_status IN ('resolved','missing_project','missing_bom','missing_bom_ref','invalid','pending')),
    updated_at TEXT NOT NULL
);

CREATE INDEX idx_bom_link_source ON bom_link_index(source_dt_project_uuid);
CREATE INDEX idx_bom_link_target ON bom_link_index(target_dt_project_uuid);
```

---

## 7. API エンドポイント一覧 (Phase 1 & 2)

### Phase 1

| Method | Path | 概要 |
|--------|------|------|
| POST | /api/v1/products | Product 作成 |
| GET | /api/v1/products | Product 一覧 |
| GET | /api/v1/products/{productId} | Product 詳細 |
| PATCH | /api/v1/products/{productId} | Product 更新 |
| DELETE | /api/v1/products/{productId} | Product 削除 |
| GET | /api/v1/products/{productId}/branches | Branch 一覧 |
| GET | /api/v1/branches/{branchId} | Branch 詳細 |
| POST | /api/v1/products/{productId}/release-lines | Release Line 作成 |
| PATCH | /api/v1/branches/{branchId} | Branch 更新 |
| GET | /api/v1/branches/{branchId}/current | Current State 取得 |
| PUT | /api/v1/branches/{branchId}/current | Current State 更新 |
| POST | /api/v1/branches/{branchId}/snapshots | Main Snapshot 作成 |
| POST | /api/v1/branches/{branchId}/releases | Release Snapshot 作成 |
| GET | /api/v1/branches/{branchId}/releases | Release 一覧 |
| GET | /api/v1/releases/{releaseId} | Release 詳細 |

### Phase 2

| Method | Path | 概要 |
|--------|------|------|
| GET | /api/v1/branches/{branchId}/current/graph | Current State の Release Graph |
| GET | /api/v1/releases/{releaseId}/graph | Release の Release Graph |

Query パラメータ: `maxDepth`, `maxNodes`, `timeout`

---

## 8. Phase 2: BOM-Link / Graph 詳細設計

### 8.1 BOM-Link 解決フロー

```
1. Root Project UUID → DT API で SBOM 取得
2. SBOM 内の components から BOM-Link (urn:cdx:...) を抽出
3. BOM-Link を解析: serialNumber / version / bom-ref
4. DT 上で参照先 Project Version を特定
5. 参照先 SBOM を取得し再帰的に探索
6. visited set で循環検出
7. maxDepth / maxNodes で制限
8. Graph (nodes + edges) として返却
```

### 8.2 Graph レスポンス構造

```json
{
  "rootNodeId": "uuid-of-root",
  "nodes": [
    {
      "id": "node-uuid",
      "projectUUID": "dt-project-uuid",
      "projectName": "Backend",
      "projectVersion": "6.4.3",
      "bomSerialNumber": "urn:uuid:...",
      "bomVersion": 1,
      "resolutionStatus": "resolved"
    }
  ],
  "edges": [
    {
      "sourceNodeId": "...",
      "targetNodeId": "...",
      "bomRef": "backend-ref",
      "resolutionStatus": "resolved"
    }
  ],
  "metadata": {
    "totalNodes": 5,
    "totalEdges": 6,
    "maxDepthReached": false,
    "maxNodesReached": false,
    "unresolvedLinks": 0,
    "cyclesDetected": 0
  }
}
```

### 8.3 循環参照対策

```go
type GraphResolver struct {
    dtClient  dtclient.DependencyTrackClient
    visited   map[string]bool  // projectUUID -> visited
    maxDepth  int
    maxNodes  int
    timeout   time.Duration
}
```

---

## 9. エラーハンドリング

### 統一エラーレスポンス

```go
type APIError struct {
    Error     string      `json:"error"`
    Message   string      `json:"message"`
    Details   interface{} `json:"details,omitempty"`
    RequestID string      `json:"requestId"`
}
```

### エラーコード

| コード | HTTP Status | 説明 |
|--------|-------------|------|
| INVALID_REQUEST | 400 | リクエスト不正 |
| PRODUCT_NOT_FOUND | 404 | Product 未発見 |
| BRANCH_NOT_FOUND | 404 | Branch 未発見 |
| RELEASE_NOT_FOUND | 404 | Release 未発見 |
| ROOT_PROJECT_NOT_FOUND | 404 | DT Project 未発見 |
| ROOT_BOM_CHANGED | 409 | BOM Hash 不一致 |
| BOM_LINK_INVALID | 422 | BOM-Link 形式不正 |
| BOM_LINK_UNRESOLVED | 422 | BOM-Link 未解決 |
| DEPENDENCY_TRACK_UNAVAILABLE | 502 | DT 接続不可 |
| GRAPH_LIMIT_EXCEEDED | 413 | Graph 探索上限超過 |
| INTERNAL_ERROR | 500 | 内部エラー |

---

## 10. Middleware / 横断関心事

### 10.1 Request ID

全リクエストに UUID を付与し、レスポンスヘッダ `X-Request-ID` に返却。ログにも記録。

### 10.2 Logging

slog を利用した構造化ログ:
- requestId, method, path, status, duration, dependencyTrackRequestDuration

### 10.3 認証プレースホルダ

```go
// internal/handler/middleware/auth_placeholder.go

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // TODO: 将来 JWT / API Key 認証を実装
        // 現時点ではすべてのリクエストを通過
        return next(c)
    }
}
```

---

## 11. Docker Compose (Release Catalog)

```yaml
# docker-compose.yml
services:
  release-catalog:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - RC_DATABASE_DRIVER=sqlite
      - RC_DATABASE_DSN=/data/release-catalog.db
      - RC_DT_BASE_URL=http://host.docker.internal:8081
      - RC_DT_API_KEY=${DT_API_KEY:-}
      - RC_DT_STUB_MODE=true
      - RC_SERVER_PORT=8080
    volumes:
      - rc-data:/data

volumes:
  rc-data:
```

---

## 12. 実装ステップ

### Phase 1: Branch / Release 基盤 (推定 8-10 日)

| Step | タスク | 依存 |
|------|--------|------|
| 1.1 | プロジェクト初期化 (go mod, Echo, Makefile, Dockerfile) | - |
| 1.2 | config パッケージ (.env 読み込み) | 1.1 |
| 1.3 | domain モデル定義 | 1.1 |
| 1.4 | DB マイグレーション (SQLite DDL) | 1.2 |
| 1.5 | Repository Interface 定義 | 1.3 |
| 1.6 | SQLite Repository 実装 (Product) | 1.4, 1.5 |
| 1.7 | SQLite Repository 実装 (Branch, CurrentState, Snapshot) | 1.6 |
| 1.8 | DT Client Interface + Stub 実装 | 1.3 |
| 1.9 | ProductService + Handler | 1.6, 1.8 |
| 1.10 | BranchService + Handler | 1.7 |
| 1.11 | CurrentState Handler (PUT/GET) | 1.7, 1.8 |
| 1.12 | ReleaseService + Handler (Snapshot/Release 作成) | 1.7 |
| 1.13 | Middleware (RequestID, Logging, Auth placeholder) | 1.1 |
| 1.14 | OpenAPI spec 作成 | 1.9-1.12 |
| 1.15 | エラーハンドリング統一 | 1.9-1.12 |
| 1.16 | テスト (Repository + Service + Handler) | 全ステップ |
| 1.17 | Docker Compose 起動確認 | 全ステップ |

### Phase 2: BOM-Link / Graph (推定 5-7 日)

| Step | タスク | 依存 |
|------|--------|------|
| 2.1 | CycloneDX BOM-Link パーサー | Phase 1 |
| 2.2 | BOM-Link Resolver (再帰解決) | 2.1 |
| 2.3 | Graph Builder (nodes/edges 構築) | 2.2 |
| 2.4 | 循環参照検出 + maxDepth/maxNodes | 2.3 |
| 2.5 | Graph API Handler | 2.4 |
| 2.6 | BOM-Link Index テーブル (オプション Cache) | 2.4 |
| 2.7 | DT Stub への SBOM テストデータ追加 | 2.1 |
| 2.8 | テスト (パーサー + Resolver + Handler) | 全ステップ |
| 2.9 | OpenAPI spec 更新 (Graph API) | 2.5 |

---

## 13. テスト方針

| レイヤ | テスト種別 | 方法 |
|--------|-----------|------|
| Repository | ユニット | SQLite in-memory (`:memory:`) |
| Service | ユニット | Repository モック (Interface) |
| DT Client | ユニット | StubClient |
| Handler | 統合 | httptest + Echo |
| BOM-Link Parser | ユニット | テスト用 CycloneDX JSON |
| Graph Resolver | ユニット | StubClient + 循環データ |
| E2E | 統合 | Docker Compose + 実 API 呼出 |

受入条件 20 項目に対応するテストを作成する。

---

## 14. 認証・認可の将来対応設計

```
Handler ← AuthMiddleware ← RoleMiddleware ← Route

- AuthMiddleware: JWT 検証 or API Key 検証
- RoleMiddleware: Role (Viewer/Editor/ReleaseManager/Admin) チェック
- context に認証情報を格納し、Service 層で参照可能
```

Phase 1 では Middleware をスキップ（全通過）とし、Router 構造だけ Role 別に分離しておく。

---

## 15. 将来の PostgreSQL 対応

1. `internal/repository/postgres/` に同一 Interface の PostgreSQL 実装を追加
2. `config.DatabaseDriver` で `sqlite` / `postgres` を切り替え
3. `migrations/postgres/` に PostgreSQL 用 DDL を配置
4. Docker Compose に PostgreSQL サービスを追加するオプション profile

---

## 16. 成果物一覧

Phase 1 & 2 完了時点での成果物:

- [x] 動作する Go API サーバ
- [x] SQLite による永続化
- [x] DT スタブによる開発環境
- [x] OpenAPI 仕様 (JSON/YAML)
- [x] Swagger UI (開発時)
- [x] Docker Compose による起動
- [x] DT 起動方法ドキュメント
- [x] 自動テスト (受入条件 1-20 対応)
- [x] Makefile (build, test, lint, migrate, run)
