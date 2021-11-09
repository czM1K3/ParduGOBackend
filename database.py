from os import getenv
from bson import ObjectId

cache = None

def get_database():
	global cache
	if cache is not None:
		return cache
	CONNECTION_STRING = getenv("MONGO")

	from pymongo import MongoClient
	client = MongoClient(CONNECTION_STRING)
	cache = client["ParduGO"]
	return cache

def insert_user(email, password, nickname):
	db = get_database()
	db["users"].insert_one({"email": email, "password": password, "nickname": nickname})

def get_user(email):
	db = get_database()
	return db["users"].find_one({"email": email})

def create_point(user_id, name, description, latitude, longitude, type):
	db = get_database()
	db["points"].insert_one({"user_id": ObjectId(user_id), "name": name, "description": description, "type": type, "location": { "type": "Point", "coordinates": [float(longitude), float(latitude)] }})

def get_points(latitude, longitude, radius):
	db = get_database()
	return db["points"].find({"location": {"$near": {"$geometry": {"type": "Point", "coordinates": [float(longitude), float(latitude)]}, "$maxDistance": float(radius)}}})
