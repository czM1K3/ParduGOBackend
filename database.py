from os import getenv


def get_database():
	CONNECTION_STRING = getenv("MONGO")

	from pymongo import MongoClient
	client = MongoClient(CONNECTION_STRING)
	return client["ParduGO"]

def insert_user(email, password, first_name, last_name):
	db = get_database()
	db["users"].insert_one({"email": email, "password": password, "first_name": first_name, "last_name": last_name})

def get_user(email):
	db = get_database()
	return db["users"].find_one({"email": email})
