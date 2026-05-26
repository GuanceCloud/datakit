import pymongo
import time
import threading
import random
import os
import logging
from datetime import datetime
from pymongo import MongoClient

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - Thread %(thread)d - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# MongoDB configuration
MONGO_CONFIG = {
    "host": os.getenv("MONGO_HOST", "mongodb"),
    "port": int(os.getenv("MONGO_PORT", "27017")),
    "username": os.getenv("MONGO_USER", "admin"),
    "password": os.getenv("MONGO_PASSWORD", "admin"),
    "authSource": "admin"
}

# Test configuration - increase for higher load
NUM_THREADS = int(os.getenv("NUM_THREADS", "20"))
RUN_DURATION = int(os.getenv("RUN_DURATION", "600"))  # in seconds - 10 minutes

# Operation types and their weights (probability) - increase insert/update operations
OPERATIONS = [
    ("insert", 25),
    ("find", 20),
    ("update", 25),
    ("delete", 15),
    ("aggregate", 15),
    ("bulk_write", 20),
    ("index_operation", 10),
    ("long_transaction", 10),
    ("large_data", 20),
]

# Total weight for probability calculation
TOTAL_WEIGHT = sum(weight for _, weight in OPERATIONS)

# Generate random data for operations
def generate_random_user():
    return {
        "username": f"user_{random.randint(1000, 9999)}",
        "email": f"user_{random.randint(1000, 9999)}@example.com",
        "created_at": datetime.now(),
        "age": random.randint(18, 70),
        "city": random.choice(["Beijing", "Shanghai", "Guangzhou", "Shenzhen"]),
        "interests": random.sample(["reading", "sports", "music", "movies", "travel"], random.randint(1, 3))
    }

def generate_random_product():
    return {
        "name": f"Product_{random.randint(1000, 9999)}",
        "price": round(random.uniform(10, 1000), 2),
        "category": random.choice(["Electronics", "Books", "Clothing", "Home", "Sports"]),
        "stock": random.randint(0, 1000),
        "created_at": datetime.now()
    }

# Operation functions

def insert_operation(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    collection_name = random.choice(collections)
    collection = db[collection_name]
    
    if collection_name == "users":
        data = generate_random_user()
    elif collection_name == "products":
        data = generate_random_product()
    else:
        data = {
            "user_id": random.randint(1, 1000),
            "product_id": random.randint(1, 1000),
            "quantity": random.randint(1, 10),
            "status": random.choice(["pending", "shipped", "delivered", "cancelled"]),
            "order_date": datetime.now()
        }
    
    result = collection.insert_one(data)
    logger.debug(f"Inserted document into {collection_name} with id: {result.inserted_id}")

def find_operation(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    collection_name = random.choice(collections)
    collection = db[collection_name]
    
    if collection_name == "users":
        query = {"city": random.choice(["Beijing", "Shanghai", "Guangzhou", "Shenzhen"])}
    elif collection_name == "products":
        query = {"category": random.choice(["Electronics", "Books", "Clothing", "Home", "Sports"])}
    else:
        query = {"status": random.choice(["pending", "shipped", "delivered", "cancelled"])}
    
    results = collection.find(query).limit(5)
    for doc in results:
        logger.debug(f"Found document in {collection_name}: {doc}")

def update_operation(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    collection_name = random.choice(collections)
    collection = db[collection_name]
    
    if collection_name == "users":
        query = {"age": {"$gt": 30}}
        update = {"$set": {"updated_at": datetime.now()}}
    elif collection_name == "products":
        query = {"stock": {"$gt": 0}}
        update = {"$inc": {"stock": -1}}
    else:
        query = {"status": "pending"}
        update = {"$set": {"status": "shipped"}}
    
    result = collection.update_one(query, update)
    logger.debug(f"Updated {result.modified_count} documents in {collection_name}")

def delete_operation(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    collection_name = random.choice(collections)
    collection = db[collection_name]
    
    if collection_name == "users":
        query = {"age": {"$lt": 20}}
    elif collection_name == "products":
        query = {"stock": 0}
    else:
        query = {"status": "cancelled"}
    
    result = collection.delete_one(query)
    logger.debug(f"Deleted {result.deleted_count} documents from {collection_name}")

def aggregate_operation(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    collection_name = random.choice(collections)
    collection = db[collection_name]
    
    if collection_name == "users":
        pipeline = [
            {"$group": {"_id": "$city", "count": {"$sum": 1}}},
            {"$sort": {"count": -1}}
        ]
    elif collection_name == "products":
        pipeline = [
            {"$group": {"_id": "$category", "avg_price": {"$avg": "$price"}}},
            {"$sort": {"avg_price": -1}}
        ]
    else:
        pipeline = [
            {"$group": {"_id": "$status", "total_orders": {"$sum": 1}}}
        ]
    
    results = collection.aggregate(pipeline)
    for doc in results:
        logger.debug(f"Aggregation result in {collection_name}: {doc}")

def bulk_write_operation(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    collection_name = random.choice(collections)
    collection = db[collection_name]
    
    operations = []
    for _ in range(10):
        if collection_name == "users":
            operations.append(pymongo.InsertOne(generate_random_user()))
        else:
            operations.append(pymongo.InsertOne(generate_random_product()))
    
    result = collection.bulk_write(operations)
    logger.debug(f"Bulk write completed: {result.inserted_count} documents inserted into {collection_name}")

def index_operation(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    collection_name = random.choice(collections)
    collection = db[collection_name]
    
    # Randomly choose to create or drop an index
    if random.choice([True, False]):
        index_name = f"temp_index_{random.randint(1000, 9999)}"
        collection.create_index([(index_name, 1)], expireAfterSeconds=3600)
        logger.debug(f"Created index {index_name} in {collection_name}")
    else:
        # Drop a random index if exists
        indexes = collection.list_indexes()
        index_names = [index["name"] for index in indexes if index["name"] not in ["_id_", "username_1", "email_1", "name_1", "price_1", "user_id_1", "status_1"]]
        if index_names:
            index_to_drop = random.choice(index_names)
            collection.drop_index(index_to_drop)
            logger.debug(f"Dropped index {index_to_drop} from {collection_name}")

def long_transaction(db):
    # Get all collections in the database
    collections = db.list_collection_names()
    if not collections:
        logger.warning("No collections found in database")
        return
    
    with client.start_session() as session:
        with session.start_transaction():
            # Perform multiple operations in a transaction
            for _ in range(10):  # More operations per transaction
                collection = db[random.choice(collections)]
                if random.choice([True, False]):
                    # Insert operation with random data
                    data = {
                        "test_id": random.randint(1, 100000),
                        "value": random.random() * 10000,
                        "timestamp": datetime.now(),
                        "transaction_data": "x" * 1000  # 1KB of data per operation
                    }
                    collection.insert_one(data, session=session)
                else:
                    query = {"_id": {"$exists": True}}
                    collection.delete_one(query, session=session)
            
            # Simulate a long transaction
            time.sleep(random.uniform(0.2, 1.0))
            
            session.commit_transaction()
    
    logger.debug("Long transaction completed")

def large_data(db):
    collection_name = "large_data_collection"
    collection = db[collection_name]
    
    # Create a large document
    large_doc = {
        "name": "Large Document",
        "data": "x" * 100000,  # 100KB of data
        "metadata": {f"field_{i}": f"value_{i}" for i in range(100)},
        "created_at": datetime.now()
    }
    
    collection.insert_one(large_doc)
    logger.debug(f"Inserted large document into {collection_name}")

# Thread function to run operations
def run_operations(thread_id):
    logger.info(f"Thread {thread_id} started")
    
    client = MongoClient(**MONGO_CONFIG)
    
    # Get all available databases
    all_dbs = client.list_database_names()
    logger.info(f"Available databases: {all_dbs}")
    
    # Filter out system databases
    test_dbs = [db for db in all_dbs if not db.startswith("system") and not db.startswith("admin") and not db.startswith("local")]
    if not test_dbs:
        logger.error("No test databases found")
        return
    
    start_time = time.time()
    
    while time.time() - start_time < RUN_DURATION:
        # Randomly choose a database for each operation
        db_name = random.choice(test_dbs)
        db = client[db_name]
        
        # Choose an operation based on probability
        rand = random.randint(1, TOTAL_WEIGHT)
        current = 0
        operation = None
        
        for op_name, weight in OPERATIONS:
            current += weight
            if rand <= current:
                operation = op_name
                break
        
        try:
            if operation == "insert":
                insert_operation(db)
            elif operation == "find":
                find_operation(db)
            elif operation == "update":
                update_operation(db)
            elif operation == "delete":
                delete_operation(db)
            elif operation == "aggregate":
                aggregate_operation(db)
            elif operation == "bulk_write":
                bulk_write_operation(db)
            elif operation == "index_operation":
                index_operation(db)
            elif operation == "long_transaction":
                long_transaction(db)
            elif operation == "large_data":
                large_data(db)
            
            # Random delay between operations
            time.sleep(random.uniform(0.01, 0.1))
            
        except Exception as e:
            logger.error(f"Thread {thread_id} error: {e}")
            # Short delay after error
            time.sleep(0.1)
    
    client.close()
    logger.info(f"Thread {thread_id} completed")

# Main function
if __name__ == "__main__":
    logger.info("Starting MongoDB test...")
    logger.info(f"Configuration: {NUM_THREADS} threads, {RUN_DURATION} seconds")
    
    # Connect to MongoDB and get information about databases
    client = MongoClient(**MONGO_CONFIG)
    
    # Get all available databases
    all_dbs = client.list_database_names()
    logger.info(f"All databases: {all_dbs}")
    
    # Filter out system databases
    test_dbs = [db for db in all_dbs if not db.startswith("system") and not db.startswith("admin") and not db.startswith("local")]
    logger.info(f"Test databases: {test_dbs}")
    
    # Count total collections across all test databases
    total_collections = 0
    for db_name in test_dbs:
        db = client[db_name]
        collections = db.list_collection_names()
        total_collections += len(collections)
        logger.info(f"Database {db_name} has {len(collections)} collections")
    
    logger.info(f"Total test collections: {total_collections}")
    
    client.close()
    
    # Create and start threads
    threads = []
    for i in range(NUM_THREADS):
        thread = threading.Thread(target=run_operations, args=(i+1,))
        threads.append(thread)
        thread.start()
    
    # Wait for all threads to complete
    for thread in threads:
        thread.join()
    
    logger.info("MongoDB test completed!")