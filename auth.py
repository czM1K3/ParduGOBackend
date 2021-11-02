from flask import request
from werkzeug.exceptions import abort
import jwt
from database import insert_user, get_user
import bcrypt


def getuser():
    if not("Authorization" in request.headers):
        abort(401, description="Missing header")
    raw_header = request.headers["Authorization"]
    split_header = raw_header.split(" ")
    if len(split_header) != 2:
        abort(401, description="Wrong header")

    return "none"


def login():
    email = request.form.get("email")
    password = request.form.get("password")
    if email is None:
        abort(400, description="Missing email")
    if password is None:
        abort(400, description="Missing password")
    user = get_user(email)
    if user is None:
        abort(401, description="Wrong email")
    if not bcrypt.checkpw(password.encode("utf-8"), user["password"].encode("utf-8")):
        abort(401, description="Wrong password")
    encoded = jwt.encode({"id": str(user["_id"])}, "secret", algorithm="HS256")
    return "Bearer " + encoded


def register():
    email = request.form.get("email")
    password = request.form.get("password")
    first_name = request.form.get("first_name")
    last_name = request.form.get("last_name")
    if email is None:
        abort(400, description="Missing email")
    if password is None:
        abort(400, description="Missing password")
    if first_name is None:
        abort(400, description="Missing first name")
    if last_name is None:
        abort(400, description="Missing last name")
    salt = bcrypt.gensalt()
    hashed_password = bcrypt.hashpw(password.encode("utf-8"), salt)
    insert_user(email, hashed_password.decode("utf-8"), first_name, last_name)
    return login()
