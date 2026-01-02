flowchart TD
    %% --- スタイル定義 ---
    classDef db fill:#e1f5fe,stroke:#0277bd,stroke-width:2px;
    classDef ext fill:#fff3e0,stroke:#ef6c00,stroke-width:2px;
    classDef app fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px;
    classDef logic fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px,stroke-dasharray: 5 5;

    %% --- ノード定義 ---
    S[Google Sheets]:::ext
    G[GAS]:::ext
    
    subgraph Backend [Go Backend]
        API[API Handler]
        DiffLogic[["差分検知 & トランザクション制御"]]:::logic
    end

    subgraph DB [PostgreSQL]
        direction TB
        Shifts[("shifts<br/>シフト本体<br/>※deleted_atあり")]:::db
        Logs[("action_log<br/>変更ログ")]:::db
        Reads[("shift_reads<br/>既読管理")]:::db
    end

    SL[Slack Service]:::ext
    App["Flutter App<br/>(UI: Card List)"]:::app

    %% --- 1. 更新フロー (Update/Create Flow) ---
    S -->|onChange| G
    G -->|"POST /api/update_shifts<br/>(New Shifts)"| API
    
    API -->|"1. SELECT Old"| Shifts
    API -->|"2. Diff (Old vs New)"| DiffLogic
    
    %% トランザクション処理
    DiffLogic -->|"3-A. 新規作成時 (INSERT)<br/>shifts作成 + shift_reads作成(False)"| DB
    DiffLogic -->|"3-B. 更新時 (UPDATE)<br/>shifts更新 + shift_readsをFalseへ"| DB
    
    %% ログと削除
    DiffLogic -->|"4. 変更ログ保存 (INSERT)"| Logs
    DiffLogic -.->|"※削除時 (Logical Delete)<br/>UPDATE deleted_at = NOW()"| Shifts

    %% 通知送信
    DiffLogic -->|"5. Send Notification<br/>(Diff Payload)"| SL

    %% --- 2. アプリ表示フロー (View Flow) ---
    App -->|GET /api/shifts| API
    API -->|"6. JOIN shifts + shift_reads<br/>WHERE deleted_at IS NULL"| DB
    DB --o|"Return JSON<br/>{...shift, is_read: false}"| App

    %% --- 3. 既読アクション (Read Action) ---
    App -->|"Tap Card<br/>POST /read"| API
    API -->|"7. UPDATE is_read=TRUE"| Reads

    %% --- 凡例 ---
    linkStyle default stroke:#333,stroke-width:1px;
    
    %% 接続線の補足
    linkStyle 4 stroke:#2e7d32,stroke-width:2px;
    linkStyle 5 stroke:#ef6c00,stroke-width:2px;

erDiagram
    %% --- 1. 物理テーブル ---

    users {
        BIGSERIAL id PK
        VARCHAR name
        VARCHAR slack_user_id
        TIMESTAMP created_at
    }

    shifts {
        BIGSERIAL id PK
        INT user_id FK
        DATE date
        INT time_id
        VARCHAR task_name
        TIMESTAMP created_at
        TIMESTAMP updated_at
        TIMESTAMP deleted_at
    }

    action_log {
        BIGSERIAL id PK
        BIGINT shift_id FK
        VARCHAR action_type "CREATE/UPDATE/DELETE"
        JSONB diff_payload "変更内容(右記参照)"
        TIMESTAMP created_at
    }

    shift_reads {
        BIGSERIAL id PK
        INT user_id FK
        BIGINT shift_id FK
        BOOLEAN is_read
        TIMESTAMP updated_at
    }

    %% --- 2. JSONデータ構造 (論理定義) ---
    
    %% diff_payloadの中身
    json_diff_structure {
        STRING user "変更対象のユーザー名"
        STRING date "対象日付"
        STRING slot "時間枠 (例: 10:00-11:00)"
        ARRAY changes "変更点のリスト"
    }

    %% changes配列の中身
    json_change_item {
        STRING field "変更項目 (role, status等)"
        STRING old "変更前の値"
        STRING new "変更後の値"
    }

    %% --- 3. リレーション定義 ---

    users ||--o{ shifts : "1人が複数のシフトを持つ"
    shifts ||--o{ action_log : "履歴を持つ"
    
    users ||--o{ shift_reads : "既読ステータスを保持する"
    shifts ||--o{ shift_reads : "既読状態を持つ"

    %% JSON構造との紐付け (点線で表現)
    action_log ||..|| json_diff_structure : "diff_payloadに格納"
    json_diff_structure ||..|{ json_change_item : "changes[]"


sequenceDiagram
    autonumber
    actor Admin as Google Sheets (GAS)
    participant API as Go Backend
    participant Logic as 差分検知ロジック
    participant DB as PostgreSQL
    participant Queue as Notification Queue
    participant Slack as Slack API

    Note over Admin, API: 1. シフト更新フロー

    Admin->>API: POST /api/updateShifts (全シフトデータ)
    
    activate API
    
    API->>Logic: データ渡し
    activate Logic
    Logic->>DB: 現在のシフトを取得
    DB-->>Logic: シフト一覧
    
    Logic->>Logic: 差分比較 (Diff Check)
    
    alt 差分あり
        Logic->>DB: UPSERT shifts (シフト更新)
        Logic->>DB: INSERT action_log (履歴保存)
        Logic->>DB: UPSERT shift_reads (is_read=false)
        
        Logic->>Queue: 通知タスクをエンキュー
        deactivate Logic
        
        API-->>Admin: 200 OK (更新完了)
        deactivate API

        Note right of Queue: ここから非同期処理

        Queue->>Queue: タスク取り出し
        activate Queue
        Queue->>Slack: POST chat.postMessage (DM/Channel)
        Slack-->>Queue: ok
        deactivate Queue
    else 差分なし
        activate Logic
        Logic-->>API: 変更なし
        deactivate Logic
        activate API
        API-->>Admin: 200 OK (変更なし)
        deactivate API
    end

sequenceDiagram
    autonumber
    actor User as App User
    participant App as Flutter App
    participant API as Go Backend
    participant DB as PostgreSQL

    Note over User, DB: 2. 閲覧・既読フロー

    %% 一覧取得
    User->>App: アプリ起動 / リフレッシュ
    activate App
    App->>API: GET /api/shifts (user_id)
    activate API
    API->>DB: SELECT shifts JOIN shift_reads
    DB-->>API: シフト + 既読フラグ(false)
    API-->>App: JSONデータ
    deactivate API
    App->>User: シフト一覧表示 (Newバッジ付き)
    deactivate App

    %% 既読アクション
    User->>App: シフトをタップ
    activate App
    App->>App: 詳細画面へ遷移 / ダイアログ表示
    
    par 裏側で既読APIを叩く
        App->>API: POST /api/shifts/{id}/read
        activate API
        API->>DB: UPDATE shift_reads SET is_read = true
        DB-->>API: Success
        API-->>App: 200 OK
        deactivate API
    and UIはレスポンス待たずにNew削除
        App->>User: リストのNewバッジを消す
    end
    deactivate App