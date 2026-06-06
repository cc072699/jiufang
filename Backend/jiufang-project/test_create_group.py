#!/usr/bin/env python3
import requests
import json

# Login to get JWT token
login_url = "http://localhost:8080/api/v1/auth/login"
login_data = {
    "username": "admin1",
    "password": "214510115lhl"
}

print("Step 1: Login to get JWT token")
response = requests.post(login_url, json=login_data)
print(f"Login response status: {response.status_code}")
print(f"Login response body: {response.text}")

if response.status_code == 200:
    login_result = response.json()
    token = login_result.get("data", {}).get("token")
    print(f"JWT token: {token}")
    
    # Create user group
    print("\nStep 2: Create user group")
    create_group_url = "http://localhost:8080/api/v1/groups"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json; charset=utf-8"
    }
    group_data = {
        "name": "运营组",
        "description": "运营人员都放进这个组内"
    }
    
    # Use data parameter with explicit encoding instead of json parameter
    response = requests.post(create_group_url, data=json.dumps(group_data).encode('utf-8'), headers=headers)
    print(f"Create group response status: {response.status_code}")
    print(f"Create group response body: {response.text}")
else:
    print("Login failed, cannot test create group")