// MongoDB initialization script
// This script will be executed when the MongoDB container starts

// Connect to admin database and authenticate
const adminDb = db.getSiblingDB('admin');
adminDb.auth('admin', 'admin');

// Create datakit user with specified roles
adminDb.createUser({
  "user": "datakit",
  "pwd": "123456",
  "roles": [
    { role: "read", db: "admin" },
    { role: "clusterMonitor", db: "admin" },
    { role: "backup", db: "admin" },
    { role: "read", db: "local" }
  ]
});

print('Created datakit user with required roles');

// Create test database
const testDb = db.getSiblingDB('test_db');

// Create multiple test collections - increase from 5 to 50 collections
const collections = [];
for (let i = 1; i <= 50; i++) {
    collections.push(`col_${i}`);
}

// Create collections and insert initial data
collections.forEach(colName => {
    const collection = testDb.getCollection(colName);
    
    // Create indexes
    collection.createIndex({ id: 1 });
    collection.createIndex({ name: 1 });
    collection.createIndex({ timestamp: 1 });
    
    // Insert sample data - more data per collection
    const docs = [];
    for (let j = 0; j < 100; j++) {
        docs.push({
            id: j,
            name: `item_${j}`,
            value: Math.random() * 1000,
            category: `cat_${Math.floor(Math.random() * 10)}`,
            timestamp: new Date(),
            description: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.'
        });
    }
    collection.insertMany(docs);
    
    print(`Initialized collection: ${colName}`);
});

// Create additional databases with many collections
const dbs = ['app_db', 'audit_db', 'analytics_db', 'test_db_2', 'test_db_3'];
dbs.forEach(dbName => {
    const currentDb = db.getSiblingDB(dbName);
    for (let i = 1; i <= 30; i++) {
        const col = currentDb.getCollection(`col_${i}`);
        col.createIndex({ id: 1 });
        // Insert some data
        const docs = [];
        for (let j = 0; j < 50; j++) {
            docs.push({
                id: j,
                value: Math.random() * 1000,
                timestamp: new Date()
            });
        }
        col.insertMany(docs);
    }
    print(`Initialized database: ${dbName} with 30 collections`);
});

// Create a large collection with many operations
const largeCol = testDb.getCollection('large_collection');
largeCol.createIndex({ id: 1 });
largeCol.createIndex({ timestamp: 1 });
largeCol.createIndex({ value: 1 });

// Insert many documents into large collection
const largeDocs = [];
for (let i = 0; i < 1000; i++) {
    largeDocs.push({
        id: i,
        timestamp: new Date(),
        value: Math.random() * 10000,
        data: Array(100).fill().map(() => ({ field: Math.random(), value: Math.random() }))
    });
}
largeCol.insertMany(largeDocs);
print('Created large_collection with 1000 documents');

print('MongoDB initialization completed successfully!');