#  EmojiDB Node.js SDK

High-performance, standalone Node.js SDK for the EmojiDB encrypted engine.

**Built by Robinson Honour**

## 🚀 Features
- **Zero-Dependency**: No Go installation required.
- **Standalone**: Automatically downloads the required engine for your platform.
- **Prisma-like Evolution**: Built-in schema persistence and diffing.
- **Military-Grade**: AES-GCM encryption with 100% Emoji-encoded persistence.

## 📦 Installation
```bash
npm install @ikwerre-dev/emojidb
```

## 🛠️ Data Management Guide

### 1. Opening & Persistence
EmojiDB stores everything in an `emojidb/` folder in your project root.
- `[dbname].db`: Encrypted data.
- `[dbname].schema.json`: Readable schema.
- `[dbname].safety`: Crash recovery buffer.

```javascript
import EmojiDB from '@ikwerre-dev/emojidb';
const db = new EmojiDB();
await db.connect();
await db.open('prod.db', 'secret-key');
```

### The `defineSchema` Method
Use this for the initial table creation.
```javascript
await db.defineSchema('users', [
    { Name: 'id',       Type: 0, Unique: true  },
    { Name: 'username', Type: 1, Unique: true  }
]);
```

### 🔄 Schema Evolution (Migrate & Pull)
EmojiDB supports Prisma-like schema evolution.

**1. Apply Changes from Schema File (`db.migrate()`)**  
Edit your `emojidb/*.schema.json` file manually, then run:
```javascript
// Syncs ALL tables from file
await db.migrate(); 

// OR sync just one table from the file
await db.migrate('users');
```

**2. Push Changes Explicitly**  
```javascript
await db.migrate('users', [
    { Name: 'id',       Type: 0, Unique: true  },
    { Name: 'username', Type: 1, Unique: true  },
    { Name: 'email',    Type: 1, Unique: true  } // New field
]);
```

**⚠️ Force Migration (Destructive)**  
If you have data that violates the new schema (e.g., changing type from String to Int), you can **force** the migration.
> **Warning**: This will DELETE any rows that do not match the new schema.
```javascript
// Force migrate 'users' table from file
await db.migrate('users', true);

// Force migrate with explicit fields
await db.migrate('users', [...fields], true);
```

**3. Pull Schema (`pull`)**  
Force-regenerate your local schema file from the actual database state.
```javascript
await db.pull();
```

### 📊 Aggregations & Utilities
```javascript
// Count records
const numActive = await db.count('users', { active: true });

// Drop Table (Destructive)
await db.dropTable('old_logs');

// 💾 Flush to Disk
// Configures explicit persistence (Good logs provided)
await db.flush('users');
```

### 📝 Field Types
Schemas are required before any data operations.
```javascript
await db.defineSchema('users', [
    { Name: 'id', Type: 0, Unique: true },
    { Name: 'username', Type: 1, Unique: false }
]);
```

### 3. CRUD Operations

#### Insert Data
```javascript
await db.insert('users', {
    id: 1,
    username: 'emoji_king',
    active: true
});
```

### 🚚 Batch Insert (Bulk)
Insert multiple records in a single atomic operation for high performance.
```javascript
const users = [
    { id: 2, username: 'alice' },
    { id: 3, username: 'bob' },
    { id: 4, username: 'charlie' }
];

await db.batchInsert('users', users);
```

### Query Data
```javascript
const results = await db.query('users', { id: 1 });
```

#### **Update**
Update records matching a filter.
```javascript
// Change username to 'robinson_honour' where id is 1
await db.update('users', { id: 1 }, { username: 'robinson_honour' });
```

#### **Delete**
Remove records matching a filter.
```javascript
await db.delete('users', { id: 1 });
```

## � File Architecture
EmojiDB follows a strict **Consolidated Directory** pattern:
- All database files live in `emojidb/`.
- If you delete the `emojidb/` folder, the database is wiped.
- To backup, simply zip and move the `emojidb/` folder.

---
*Created by Robinson Honour.*
