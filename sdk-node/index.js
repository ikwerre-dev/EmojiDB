import { spawn } from 'child_process';
import path from 'path';
import readline from 'readline';
import fs from 'fs';
import https from 'https';
import os from 'os';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Custom Error for cleaner structured errors
export class EmojiDBError extends Error {
    constructor(message, originalStack) {
        super(message);
        this.name = 'EmojiDBError';
        this.code = 'UNKNOWN_ERROR';
        this.details = {};

        // Parse Known Go Errors
        if (message.includes('table not found:')) {
            this.code = 'TABLE_NOT_FOUND';
            this.details = { table: message.split(': ')[1] };
        } else if (message.includes('missing field:')) {
            this.code = 'MISSING_FIELD';
            this.details = { field: message.split(': ')[1] };
        } else if (message.includes('unique constraint violation:')) {
            this.code = 'UNIQUE_CONSTRAINT';
            this.details = { field: message.split(': ')[1] };
        } else if (message.includes('type mismatch')) {
            this.code = 'TYPE_MISMATCH';
            // "type mismatch for field: name" -> name
            const parts = message.split(': ');
            if (parts.length > 1) this.details = { field: parts[1] };
        }

        // Hide the stack trace from default console.log if requested
        // by overriding inspect custom or just keeping it clean.
        // We stitch the stack for debugging if needed, but the user wants "direct error in json".

        if (originalStack) {
            // We keep the stack for debugging but maybe we don't show it by default
            this.stack = `${this.name} [${this.code}]: ${this.message}\n${originalStack.substring(originalStack.indexOf('\n') + 1)}`;
        }
    }

    // This method is called by Node.js console.log()
    [Symbol.for('nodejs.util.inspect.custom')](depth, options) {
        return {
            error: this.name,
            code: this.code,
            message: this.message,
            ...this.details
        };
    }

    // For JSON.stringify(err)
    toJSON() {
        return {
            error: this.name,
            code: this.code,
            message: this.message,
            ...this.details
        };
    }
}

export class BinaryManager {
    constructor() {
        this.platform = os.platform();
        this.arch = os.arch();
        this.binDir = path.join(__dirname, 'bin');
        this.engineName = `emojidb-${this.platform}-${this.arch}${this.platform === 'win32' ? '.exe' : ''}`;
        this.enginePath = path.join(this.binDir, this.engineName);
    }

    async download(url, dest) {
        return new Promise((resolve, reject) => {
            const file = fs.createWriteStream(dest);
            const request = (requestUrl) => {
                https.get(requestUrl, (response) => {
                    // Handle All Redirects (301, 302, 307, 308)
                    if ([301, 302, 307, 308].includes(response.statusCode) && response.headers.location) {
                        request(response.headers.location);
                        return;
                    }

                    if (response.statusCode !== 200) {
                        fs.unlink(dest, () => { });
                        reject(new Error(`Server returned HTTP ${response.statusCode}. Please ensure a Release exists with the correctly named binary: ${this.engineName}`));
                        return;
                    }

                    response.pipe(file);
                    file.on('finish', () => {
                        file.close();
                        if (this.platform !== 'win32') {
                            fs.chmodSync(dest, 0o755);
                        }
                        resolve(dest);
                    });
                }).on('error', (err) => {
                    fs.unlink(dest, () => { });
                    reject(err);
                });
            };
            request(url);
        });
    }

    async ensureBinary() {
        if (fs.existsSync(this.enginePath)) return this.enginePath;

        console.log(`🚀 EmojiDB: Engine not found for ${this.platform}-${this.arch}. Attempting download...`);

        if (!fs.existsSync(this.binDir)) {
            fs.mkdirSync(this.binDir, { recursive: true });
        }

        const url = `https://github.com/ikwerre-dev/EmojiDB/releases/latest/download/${this.engineName}`;
        try {
            await this.download(url, this.enginePath);
            console.log('✅ EmojiDB: Engine ready.');
            return this.enginePath;
        } catch (err) {
            console.error(`\n❌ EmojiDB Setup Error: ${err.message}`);
            console.error(`💡 FIX: You MUST compile and upload '${this.engineName}' to your GitHub Release for this to work standalone.`);
            throw err;
        }
    }
}

class EmojiDB {
    constructor(options = {}) {
        this.manager = new BinaryManager();
        this.enginePath = options.enginePath || null;
        this.process = null;
        this.rl = null;
        this.pending = new Map();
    }

    async connect() {
        if (!this.enginePath) {
            this.enginePath = await this.manager.ensureBinary();
        }

        return new Promise((resolve, reject) => {
            this.process = spawn(this.enginePath);

            this.rl = readline.createInterface({
                input: this.process.stdout,
                terminal: false
            });

            this.rl.on('line', (line) => {
                try {
                    const res = JSON.parse(line);
                    const p = this.pending.get(res.id);
                    if (p) {
                        if (res.error) {
                            // Rehydrate the error with the original stack trace
                            p.reject(new EmojiDBError(res.error, p.stack));
                        }
                        else p.resolve(res.data);
                        this.pending.delete(res.id);
                    }
                } catch (e) {
                    console.error('Failed to parse engine response:', e);
                }
            });

            this.process.stderr.on('data', (data) => {
                console.error(`Engine Error: ${data}`);
            });

            this.process.on('error', (err) => {
                reject(new Error(`Failed to start engine: ${err.message}`));
            });

            setTimeout(() => {
                resolve({ status: 'connected', pid: this.process.pid });
            }, 100);
        });
    }

    get status() {
        if (this.process && !this.process.killed) {
            return { status: 'connected', pid: this.process.pid };
        }
        return { status: 'disconnected' };
    }

    async send(method, params = {}) {
        const id = Math.random().toString(36).substring(7);
        // Capture stack trace at the call site (sync)
        const stackContainer = {};
        Error.captureStackTrace(stackContainer);

        return new Promise((resolve, reject) => {
            this.pending.set(id, { resolve, reject, stack: stackContainer.stack });
            const payload = JSON.stringify({ id, method, params });
            if (!this.process || this.process.killed) {
                return reject(new Error("Database not connected. Call db.connect() first."));
            }
            this.process.stdin.write(payload + '\n');
        });
    }

    async open(dbPath, key) {
        this.dbPath = dbPath;
        return this.send('open', { path: dbPath, key });
    }

    async defineSchema(table, fields) {
        return this.send('define_schema', { table, fields });
    }

    async insert(table, row) {
        return this.send('insert', { table, row });
    }

    async batchInsert(table, rows) {
        return this.send('batch_insert', { table, records: rows });
    }

    async query(table, match = {}) {
        return this.send('query', { table, match });
    }

    async migrate(table, fieldsOrForce, forceArg = false) {
        let fields = null;
        let force = forceArg;

        if (Array.isArray(fieldsOrForce)) {
            fields = fieldsOrForce;
        } else if (typeof fieldsOrForce === 'boolean') {
            force = fieldsOrForce;
        }

        // Case 3: Explicit Migration (table + fields provided)
        if (table && fields) {
            return this.send('sync_schema', { table, fields, force });
        }

        // Case 1 & 2: File-Based Migration
        if (!this.dbPath) {
            throw new Error("Database not open. Call open() first.");
        }

        // Construct path: emojidb/[basename].schema.json
        const baseName = path.basename(this.dbPath);
        const schemaPath = path.join('emojidb', baseName + '.schema.json');

        if (!fs.existsSync(schemaPath)) {
            throw new Error(`Schema file not found at ${schemaPath}. Cannot migrate from file.`);
        }

        const schemaContent = fs.readFileSync(schemaPath, 'utf8');
        let schema;
        try {
            schema = JSON.parse(schemaContent);
        } catch (e) {
            throw new Error(`Failed to parse schema file: ${e.message}`);
        }

        // Case 2: Migrate Single Table from File
        if (table) {
            if (!schema[table]) {
                throw new Error(`Table '${table}' not defined in schema file.`);
            }
            return this.send('sync_schema', { table, fields: schema[table].Fields, force });
        }

        // Case 1: Migrate All Tables from File
        const tables = Object.keys(schema);
        if (tables.length === 0) return "No tables to migrate.";

        const results = [];
        for (const t of tables) {
            // We run them sequentially to be safe
            await this.send('sync_schema', { table: t, fields: schema[t].Fields, force });
            results.push(t);
        }
        return `Migrated ${results.length} tables: ${results.join(', ')}`;
    }

    async pull() {
        return this.send('pull_schema');
    }

    async count(table, match = {}) {
        return this.send('count', { table, match });
    }

    async dropTable(table) {
        return this.send('drop_table', { table });
    }

    async flush(table) {
        console.log(`💾 EmojiDB: Persisting '${table}' to disk...`);
        const res = await this.send('flush', { table });
        console.log(`✅ EmojiDB: '${table}' persisted successfully.`);
        return res;
    }

    async update(table, match, updateData) {
        return this.send('update', { table, match, update: updateData });
    }

    async delete(table, match) {
        return this.send('delete', { table, match });
    }

    async secure() {
        return this.send('secure');
    }

    async rekey(newKey, masterKey) {
        return this.send('rekey', { new_key: newKey, master_key: masterKey });
    }

    async close() {
        if (this.process && !this.process.killed) {
            await this.send('close');
            this.process.kill();
        }
    }
}

export default EmojiDB;
