import EmojiDB from './index.js';
import path from 'path';
import os from 'os';
import { fileURLToPath } from 'url';

// Define __dirname equivalent for ESM
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function runTest() {
    try {
        console.log('--- EMOJIDB NODE.JS TEST (ESM) ---');

        // Use the locally built binary for testing new features!
        const platform = os.platform();
        const arch = os.arch();
        const engineName = `emojidb-${platform}-${arch}${platform === 'win32' ? '.exe' : ''}`;
        const localEnginePath = path.resolve(__dirname, '../bin/engines', engineName);

        console.log(`🧪 Testing with local engine: ${localEnginePath}`);
        const db = new EmojiDB({ enginePath: localEnginePath });

        const status = await db.connect();
        console.log('1. Engine Status:', status);
        // Output: { status: 'connected', pid: 12345 }

        await db.open('node_showcase.db', 'node-secret-2025');
        console.log('2. Database Opened');

        await db.defineSchema('users', [
            { Name: 'id', Type: 0, Unique: true },
            { Name: 'username', Type: 1, Unique: false }
        ]);
        console.log('3. Schema Defined');

        await db.insert('users', { id: 101, username: 'emoji_king' });
        await db.insert('users', { id: 102, username: 'node_master' });
        console.log('4. Data Inserted');

        const results = await db.query('users', { id: 101 });
        console.log('5. Query Results:', results);

        if (results.length > 0 && results[0].username === 'emoji_king') {
            console.log('✅ TEST PASSED: Node.js successfully interacted with Go core via ESM!');
        } else {
            console.log('❌ TEST FAILED: Data mismatch');
        }

        // 6. Test Error Handling (Stack Trace Validation)
        try {
            console.log('6. Testing Error Stack Trace...');
            await db.insert('users', { id: 101, username: 'duplicate_king' });
        } catch (err) {
            console.log('✅ Error Caught:', err); // Should look cleaner now
            if (err.code === 'UNIQUE_CONSTRAINT') {
                console.log('✅ Structured Error Code Verified: UNIQUE_CONSTRAINT');
            } else {
                console.log('❌ Failed to parse error structure:', err);

            }
        }

        // 6.5 Test Bulk Insert
        console.log('7. Testing Bulk Insert...');
        const batchUsers = [
            { id: 200, username: 'batch_1' },
            { id: 201, username: 'batch_2' },
            { id: 202, username: 'batch_3' }
        ];
        await db.batchInsert('users', batchUsers);
        console.log('✅ Bulk Insert Successful!');

        await db.close();
        console.log('8. Connection Closed');

    } catch (err) {
        console.error('❌ TEST ERROR:', err.message);
    }
}

runTest();
