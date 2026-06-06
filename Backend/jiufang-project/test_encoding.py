#!/usr/bin/env python3
import psycopg2

# Connect to PostgreSQL
conn = psycopg2.connect(
    host="localhost",
    port=5432,
    user="postgres",
    password="214Dahai",
    dbname="jiufang"
)

# Check database encoding
cur = conn.cursor()
cur.execute("SHOW SERVER_ENCODING;")
server_encoding = cur.fetchone()[0]
print(f"Server encoding: {server_encoding}")

cur.execute("SHOW CLIENT_ENCODING;")
client_encoding = cur.fetchone()[0]
print(f"Client encoding: {client_encoding}")

# Test inserting Chinese characters
cur.execute("INSERT INTO user_groups (snowflake_id, name, description, created_at, updated_at) VALUES (9999999999999999999, '测试组', '这是一个测试组', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING name;")
result = cur.fetchone()[0]
print(f"Inserted name: {result}")

# Query Chinese characters
cur.execute("SELECT name FROM user_groups WHERE snowflake_id = 9999999999999999999;")
query_result = cur.fetchone()[0]
print(f"Queried name: {query_result}")

# Clean up
cur.execute("DELETE FROM user_groups WHERE snowflake_id = 9999999999999999999;")
conn.commit()

cur.close()
conn.close()