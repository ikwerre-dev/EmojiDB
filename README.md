# 🦄 EmojiDB: Legend of the Encrypted Ledger

> **"The Database That Smiles Back"** 🛡️💎🌈

EmojiDB is a high-performance, embedded database engine written in Go. It encrypts **everything**—data, headers, and schemas—into 100% valid Emoji sequences. Now featuring a powerful **Node.js SDK** for modern web development.

---

## 🚀 Quick Start

### 1. Installation
```bash
npm install @ikwerre-dev/emojidb
```
*(The engine binary is automatically downloaded for your platform: Mac, Linux, or Windows).*

### 2. Basic Usage
```javascript
import EmojiDB from '@ikwerre-dev/emojidb';

const db = new EmojiDB();
await db.connect();
await db.open('my_app.db', 'super-secret-key');
```

---

## 📐 Defining Schemas & Creating Tables

Before you can store data, you must define the **Schema**. This automatically creates the table and enforces data integrity.

### The `defineSchema` Method
```javascript
await db.defineSchema('users', [
    { Name: 'id',       Type: 0, Unique: true  },
    { Name: 'username', Type: 1, Unique: true  }
]);
```

### 🔄 Schema Evolution (Migrate & Pull)
**1. Migrate using Schema File**
```javascript
// Edits to 'emojidb/*.schema.json' are applied automatically
await db.migrate(); 
```

**2. Push Changes Explicitly**
```javascript
await db.migrate('users', [
    { Name: 'id',       Type: 0, Unique: true  },
    { Name: 'username', Type: 1, Unique: true  },
    { Name: 'email',    Type: 1, Unique: true  } 
]);
```

**3. Pull Schema (`pull`)**
```javascript
await db.pull(); // Regenerates local schema files from DB state
```

### 📝 Field Types
| Type ID | Data Type | Example |
| :--- | :--- | :--- |
| `0` | **Integer** | `123` |
| `1` | **String** | `"robinson"` |
| `2` | **Boolean** | `true` |
| `3` | **Float** | `10.5` |
| `4` | **Map** | `{ "a": 1 }` |

> **💡 Note:** Schemas are persisted to disk as readable JSON files (e.g., `emojidb/my_app.db.schema.json`).

---

## 🛠️ Data Operations (CRUD)

### Insert Data
```javascript
await db.insert('users', {
    id: 1,
    username: 'emoji_king',
    active: true
});
```

### Query Data
```javascript
// Find user with id 1
const users = await db.query('users', { id: 1 });
console.log(users); 
// Output: [{ id: 1, username: 'emoji_king', active: true }]
```

### Update Data
```javascript
// Rename user 1
await db.update('users', { id: 1 }, { username: 'robinson_honour' });
```

### Delete Data
```javascript
// Remove user 1
await db.delete('users', { id: 1 });
```

---

## 🔐 Military-Grade Security

EmojiDB isn't just cute; it's a fortress.
- **AES-GCM Encryption**: All data is encrypted at rest.
- **Emoji Encoding**: Ciphertext is encoded into emojis (e.g., 🔒🦄🌵), making it visually distinct and obfuscated.
- **Master Key Rotation**: Built-in support for re-keying your entire database (`db.rekey()`).

### Security Files
All database artifacts are stored in the `emojidb/` directory:
- `*.db`: The encrypted data.
- `*.safety`: Crash recovery logs.
- `secure.pem`: (Optional) Master key file for managed security.

---

## 🌍 Platform Support
Our automated build system supports:
- **macOS**: ARM64 (M1/M2/M3) & Intel x64
- **Linux**: x64 & ARM64
- **Windows**: x64 & ARM64

---

*Built by Robinson Honour. 🚀*