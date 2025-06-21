// MongoDB initialization script
// This script runs when MongoDB container starts

db = db.getSiblingDB('backend_template');

// Create collections and indexes
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "username": 1 }, { unique: true });

// You can add more initialization here