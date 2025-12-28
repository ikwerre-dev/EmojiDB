export interface EmojiDBOptions {
    enginePath?: string;
}

export interface ConnectionStatus {
    status: 'connected' | 'disconnected';
    pid?: number;
}

export interface Field {
    Name: string;
    Type: number; // 0 for int, 1 for string, etc.
    Unique: boolean;
}

export interface Schema {
    table: string;
    fields: Field[];
}

export default class EmojiDB {
    constructor(options?: EmojiDBOptions);

    /**
     * Connects to the underlying EmojiDB engine process.
     * Auto-downloads the binary if not found (in standalone mode).
     */
    connect(): Promise<ConnectionStatus>;

    /**
     * Gets the current status of the engine process.
     */
    get status(): ConnectionStatus;

    /**
     * Opens or creates a database at the specified path.
     * @param dbPath Path to the database file (relative or absolute).
     * @param key Secret key for encryption/decryption.
     */
    open(dbPath: string, key: string): Promise<string>;

    /**
     * Defines a schema for a table.
     * @param table Name of the table.
     * @param fields Array of field definitions.
     */
    defineSchema(table: string, fields: Field[]): Promise<string>;

    /**
     * Inserts a row into a table.
     * @param table Name of the table.
     * @param row Key-value pair object representing the row data.
     */
    insert(table: string, row: Record<string, any>): Promise<string>;

    /**
     * Applies schema changes to the database (Migration).
     * @param table (Optional) Name of the table. If omitted, migrates ALL tables from the local schema file.
     * @param fields (Optional) New field definitions. If omitted, reads from the local schema file.
     */
    migrate(table?: string, fields?: Field[]): Promise<string>;

    /**
     * Forces the engine to regenerate the local schema file based on the database content (Pull).
     */
    pull(): Promise<string>;

    /**
     * Queries a table for rows matching the criteria.
     * @param table Name of the table.
     * @param match (Optional) Filter object to match rows.
     */
    query(table: string, match?: Record<string, any>): Promise<any[]>;

    /**
     * Updates rows in a table that match the criteria.
     * @param table Name of the table.
     * @param match Filter object to select rows to update.
     * @param updateData Object containing the new values.
     */
    update(table: string, match: Record<string, any>, updateData: Record<string, any>): Promise<string>;

    /**
     * Deletes rows from a table that match the criteria.
     * @param table Name of the table.
     * @param match Filter object to select rows to delete.
     */
    delete(table: string, match: Record<string, any>): Promise<string>;

    /**
     * Secures the database by generating a one-time master key.
     */
    secure(): Promise<string>;

    /**
     * Rotates the database encryption key.
     * @param newKey The new key to re-encrypt data with.
     * @param masterKey The master key for authorization.
     */
    rekey(newKey: string, masterKey: string): Promise<string>;

    /**
     * Closes the connection to the database engine.
     */
    close(): Promise<void>;
}
